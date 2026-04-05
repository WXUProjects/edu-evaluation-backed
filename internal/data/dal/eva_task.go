package dal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"edu-evaluation-backed/internal/common/data/cache"
	"edu-evaluation-backed/internal/common/utils"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TaskDal 评教任务数据访问层
type TaskDal struct {
	db  *gorm.DB
	rdb *redis.Client
	hc  *cache.HealthChecker
}

// CreateTask 创建评教任务
func (d *TaskDal) CreateTask(title string, courses []model.Course) (uint, error) {
	task := &model.EvaluationTask{
		Title:   title,
		Courses: courses,
		Status:  0,
	}
	err := d.db.Create(task).Error
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

// taskListResult 任务列表缓存数据
type taskListResult struct {
	Tasks []model.EvaluationTask `json:"tasks"`
	Total int64                  `json:"total"`
}

// GetTaskList 获取评教任务列表（带缓存）
func (d *TaskDal) GetTaskList(page, pageSize, status int) (*[]model.EvaluationTask, int64, error) {
	ctx := context.Background()
	key := cache.TaskListKey(status, page, pageSize)

	result, err := cache.Get[taskListResult](ctx, d.rdb, d.hc, key, 5*time.Minute, func() (*taskListResult, error) {
		var tasks []model.EvaluationTask
		var total int64
		page, pageSize = utils.PageNumHandle(page, pageSize)
		baseQ := d.db.Model(&model.EvaluationTask{})
		if status != -1 {
			baseQ = baseQ.Where("status = ?", status)
		}
		if err := baseQ.Count(&total).Order("id desc").Limit(pageSize).Offset(utils.CalculateOffset(page, pageSize)).Find(&tasks).Error; err != nil {
			return nil, err
		}
		return &taskListResult{Tasks: tasks, Total: total}, nil
	})
	if err != nil {
		return nil, 0, err
	}
	return &result.Tasks, result.Total, nil
}

// GetTaskDetail 获取评教任务详情
func (d *TaskDal) GetTaskDetail(taskID uint) (*model.EvaluationTask, error) {
	var task model.EvaluationTask
	err := d.db.Where("id = ?", taskID).Preload("Courses").Preload("Courses.Students").Preload("Courses.Teachers").First(&task).Error
	return &task, err
}

// ChangeTaskStatus 修改评教任务状态
func (d *TaskDal) ChangeTaskStatus(taskID uint, status int) error {
	err := d.db.Model(&model.EvaluationTask{}).Where("id = ?", taskID).Update("status", status).Error
	if err != nil {
		return err
	}
	// 清除任务列表缓存
	ctx := context.Background()
	go cache.DeleteByPattern(ctx, d.rdb, d.hc, cache.TaskListPattern())
	return nil
}

// DeleteTaskDetails 删除任务的评教详情（硬删除）
func (d *TaskDal) DeleteTaskDetails(taskID uint, courseIDs []uint) error {
	if len(courseIDs) == 0 {
		return nil
	}
	return d.db.Unscoped().Where("task_id = ? AND course_id IN ?", taskID, courseIDs).Delete(&model.EvaluationDetail{}).Error
}

// StudentTaskDetail 获取学生在某任务下的课程列表（带缓存）
func (d *TaskDal) StudentTaskDetail(studentNo string, taskID uint) ([]model.Course, error) {
	ctx := context.Background()
	key := cache.TaskCoursesKey(taskID, studentNo)

	result, err := cache.Get[[]model.Course](ctx, d.rdb, d.hc, key, 10*time.Minute, func() (*[]model.Course, error) {
		var courses []model.Course
		err := d.db.Model(&model.Course{}).
			Preload("Teachers").
			Joins("JOIN evaluation_courses ec ON ec.course_id = courses.id").
			Joins("JOIN course_students cs ON cs.course_id = courses.id").
			Where("ec.evaluation_task_id = ? AND cs.student_student_no = ?", taskID, studentNo).
			Find(&courses).Error
		if err != nil {
			return nil, err
		}
		return &courses, nil
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// EvalCheckData 评教状态检查缓存数据
type EvalCheckData struct {
	ID        uint   `json:"id"`
	HasEval   bool   `json:"has_eval"`
	Score     int    `json:"score,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Summary   string `json:"summary,omitempty"`
}

// BatchGetEvalChecks 批量查询评教状态（使用 Redis MGET）
// keys: cache key 列表, params: 每个元素是 (taskID, courseID, studentNo, teacherID) 的元组
func (d *TaskDal) BatchGetEvalChecks(ctx context.Context, keys []string, params []struct {
	TaskID     uint
	CourseID   uint
	StudentNo  string
	TeacherID  uint
}) (map[string]*EvalCheckData, error) {
	return cache.MGet[EvalCheckData](ctx, d.rdb, d.hc, keys, func(missingKeys []string) (map[string]*EvalCheckData, error) {
		// 构建缺失 key 对应的参数索引
		keySet := make(map[string]int)
		for i, k := range keys {
			keySet[k] = i
		}

		result := make(map[string]*EvalCheckData)
		for _, mk := range missingKeys {
			idx, ok := keySet[mk]
			if !ok {
				continue
			}
			p := params[idx]
			// 查库
			data, err := d.getEvalCheckFromDB(p.TaskID, p.CourseID, p.StudentNo, p.TeacherID)
			if err != nil {
				return nil, err
			}
			result[mk] = data
		}
		return result, nil
	}, 5*time.Minute)
}

// getEvalCheckFromDB 从数据库查询单个评教状态
func (d *TaskDal) getEvalCheckFromDB(taskID, courseID uint, studentNo string, teacherID uint) (*EvalCheckData, error) {
	student := model.Student{}
	if err := d.db.Where("student_no = ?", studentNo).First(&student).Error; err != nil {
		return &EvalCheckData{ID: 0, HasEval: false}, nil
	}

	r := model.EvaluationDetail{}
	err := d.db.Where("task_id = ? AND course_id = ? AND student_id = ? AND teacher_id = ?",
		taskID, courseID, student.ID, teacherID).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &EvalCheckData{ID: 0, HasEval: false}, nil
	}
	if err != nil {
		return &EvalCheckData{ID: 0, HasEval: false}, err
	}
	return &EvalCheckData{
		ID:      r.ID,
		HasEval: true,
		Score:   r.Score,
		Detail:  r.Detail,
		Summary: r.Summary,
	}, nil
}

// GetTaskEvaluationDetail 获取任务评价详情（保留兼容性）
func (d *TaskDal) GetTaskEvaluationDetail(taskID uint, courseID uint, studentNo string, teacherId uint) (model.EvaluationDetail, error) {
	student := model.Student{}
	err := d.db.Where("student_no = ?", studentNo).First(&student).Error
	r := model.EvaluationDetail{}
	err = d.db.Where("task_id = ? AND course_id = ? AND student_id = ? AND teacher_id = ?", taskID, courseID, student.ID, teacherId).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.ID = 0
		return r, nil
	}
	return r, err
}

// SubmitEvaluation 提交评价（写DB + 缓存失效）
func (d *TaskDal) SubmitEvaluation(taskID uint, courseID, teacherId uint, studentNo string, detail, summary string, score int) error {
	// 先查询有没有评价过
	r, err := d.GetTaskEvaluationDetail(taskID, courseID, studentNo, teacherId)
	if err != nil {
		return err
	}
	if r.ID != 0 {
		return errors.New("已评价过")
	}
	student := model.Student{}
	d.db.Where("student_no = ?", studentNo).First(&student)
	dc := model.EvaluationDetail{
		TaskId:    taskID,
		CourseId:  courseID,
		StudentId: student.ID,
		TeacherId: teacherId,
		Detail:    detail,
		Score:     score,
		Summary:   summary,
	}
	// 获取课程
	course := model.Course{}
	d.db.Where("id = ?", courseID).First(&course)
	course.EvaluationNum += 1
	course.EvaluationScore += score
	d.db.Save(&course)

	if err := d.db.Create(&dc).Error; err != nil {
		return err
	}

	// 异步失效缓存
	go func() {
		ctx := context.Background()
		// 删除评教状态缓存
		evalKey := cache.EvalCheckKey(taskID, courseID, studentNo, teacherId)
		cache.Delete(ctx, d.rdb, d.hc, evalKey)
		// 删除学生课程列表缓存
		coursesKey := cache.TaskCoursesKey(taskID, studentNo)
		cache.Delete(ctx, d.rdb, d.hc, coursesKey)
		// 删除课程详情缓存（计数器变了）
		courseKey := cache.CourseDetailKey(courseID)
		cache.Delete(ctx, d.rdb, d.hc, courseKey)
	}()

	return nil
}

// InvalidateTaskCachesByTaskStudent 清除某学生在某任务下的所有缓存（预热用）
func (d *TaskDal) InvalidateTaskCachesByTaskStudent(ctx context.Context, taskID uint, studentNo string) {
	// 删除评教状态缓存（该学生在此任务下的所有组合）
	go cache.DeleteByPattern(ctx, d.rdb, d.hc, cache.EvalCheckPattern(taskID, studentNo))
	// 删除学生课程列表缓存
	go cache.Delete(ctx, d.rdb, d.hc, cache.TaskCoursesKey(taskID, studentNo))
}

// PreloadEvalCheck 预热评教状态缓存（预热时设置"未评教"状态）
func (d *TaskDal) PreloadEvalCheck(ctx context.Context, taskID, courseID uint, studentNo string, teacherID uint) error {
	key := cache.EvalCheckKey(taskID, courseID, studentNo, teacherID)
	data := &EvalCheckData{ID: 0, HasEval: false}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return d.rdb.Set(ctx, key, b, 5*time.Minute).Err()
}

// PreloadTaskCourses 预热学生课程列表缓存
func (d *TaskDal) PreloadTaskCourses(ctx context.Context, taskID uint, studentNo string) error {
	var courses []model.Course
	err := d.db.Model(&model.Course{}).
		Preload("Teachers").
		Joins("JOIN evaluation_courses ec ON ec.course_id = courses.id").
		Joins("JOIN course_students cs ON cs.course_id = courses.id").
		Where("ec.evaluation_task_id = ? AND cs.student_student_no = ?", taskID, studentNo).
		Find(&courses).Error
	if err != nil {
		return err
	}

	key := cache.TaskCoursesKey(taskID, studentNo)
	b, err := json.Marshal(courses)
	if err != nil {
		return err
	}
	return d.rdb.Set(ctx, key, b, 10*time.Minute).Err()
}

// NewTaskDal 创建评教任务数据访问层实例
func NewTaskDal(data *data.Data) *TaskDal {
	return &TaskDal{
		db:  data.DB,
		rdb: data.RDB,
		hc:  data.HC,
	}
}

// TeacherEvaluationResult 教师评教结果（用于导出）
type TeacherEvaluationResult struct {
	TeacherName    string
	WorkNo        string
	CourseName    string
	ClassName     string
	TotalScore    float64
	AvgScore      float64
	Rank          int
	QuestionScores [][]int
}

// TeacherEvaluationDetail 教师评教详情（用于生成PDF）
type TeacherEvaluationDetail struct {
	TeacherName  string
	WorkNo       string
	CourseName   string
	ClassName    string
	AvgScore     float64
	Rank         int
	TotalTeachers int
	Comments     []string
	Summaries    []string
}

// GetTaskEvaluationResults 获取任务下所有教师的评教结果
func (d *TaskDal) GetTaskEvaluationResults(taskID uint) ([]TeacherEvaluationResult, error) {
	var results []TeacherEvaluationResult

	rows, err := d.db.Raw(fmt.Sprintf(`
		SELECT
			ed.id,
			ed.teacher_id,
			ed.course_id,
			ed.detail,
			ed.score,
			t.name as teacher_name,
			t.work_no as teacher_work_no,
			c.course_name,
			c.class_name
		FROM evaluation_details ed
		INNER JOIN teachers t ON ed.teacher_id = t.id
		INNER JOIN courses c ON ed.course_id = c.id
		WHERE ed.task_id = %d
	`, taskID)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type groupKey struct {
		teacherID uint
		courseID  uint
	}
	type groupData struct {
		teacherName string
		workNo     string
		courseName  string
		className   string
		details    []struct {
			detail string
			score  int
		}
	}
	groupMap := make(map[groupKey]*groupData)

	for rows.Next() {
		var id, teacherID, courseID uint
		var detail string
		var score int
		var teacherName, workNo, courseName, className string

		if err := rows.Scan(&id, &teacherID, &courseID, &detail, &score, &teacherName, &workNo, &courseName, &className); err != nil {
			return nil, err
		}

		key := groupKey{teacherID: teacherID, courseID: courseID}
		if _, ok := groupMap[key]; !ok {
			groupMap[key] = &groupData{
				teacherName: teacherName,
				workNo:     workNo,
				courseName:  courseName,
				className:   className,
				details:     make([]struct{ detail string; score int }, 0),
			}
		}
		groupMap[key].details = append(groupMap[key].details, struct {
			detail string
			score  int
		}{detail: detail, score: score})
	}

	if len(groupMap) == 0 {
		return nil, errors.New("暂无评教数据")
	}

	for _, data := range groupMap {
		if len(data.details) == 0 {
			continue
		}

		var totalScore float64
		questionScores := make([][]int, len(data.details))

		for i, d := range data.details {
			totalScore += float64(d.score)
			questionScores[i] = parseDetailScores(d.detail)
		}

		result := TeacherEvaluationResult{
			TeacherName:    data.teacherName,
			WorkNo:        data.workNo,
			CourseName:    data.courseName,
			ClassName:     data.className,
			TotalScore:    totalScore,
			AvgScore:      totalScore / float64(len(data.details)),
			QuestionScores: questionScores,
		}
		results = append(results, result)
	}

	return results, nil
}

// GetTeacherEvaluationDetailsForPDF 获取教师评教详情（用于生成PDF）
func (d *TaskDal) GetTeacherEvaluationDetailsForPDF(taskID uint) ([]TeacherEvaluationDetail, error) {
	rows, err := d.db.Raw(fmt.Sprintf(`
		SELECT
			t.name as teacher_name,
			t.work_no as teacher_work_no,
			c.course_name,
			c.class_name,
			AVG(ed.score) as avg_score,
			ed.detail,
			ed.summary
		FROM evaluation_details ed
		INNER JOIN teachers t ON ed.teacher_id = t.id
		INNER JOIN courses c ON ed.course_id = c.id
		WHERE ed.task_id = %d
		GROUP BY t.id, t.name, t.work_no, c.id, c.course_name, c.class_name, ed.detail, ed.summary
		ORDER BY t.name, c.class_name, avg_score DESC
	`, taskID)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type teacherCourseData struct {
		TeacherName string
		WorkNo     string
		CourseName string
		ClassName  string
		TotalScore float64
		Count      int
		Comments   []string
		Summaries  []string
	}

	dataMap := make(map[string]*teacherCourseData)

	for rows.Next() {
		var teacherName, workNo, courseName, className, detail, summary string
		var avgScore float64

		if err := rows.Scan(&teacherName, &workNo, &courseName, &className, &avgScore, &detail, &summary); err != nil {
			continue
		}

		key := teacherName + "_" + className
		if _, ok := dataMap[key]; !ok {
			dataMap[key] = &teacherCourseData{
				TeacherName: teacherName,
				WorkNo:     workNo,
				CourseName: courseName,
				ClassName:  className,
				Comments:   make([]string, 0),
				Summaries:  make([]string, 0),
			}
		}

		dataMap[key].TotalScore += avgScore
		dataMap[key].Count++
		if detail != "" {
			dataMap[key].Comments = append(dataMap[key].Comments, detail)
		}
		if summary != "" {
			dataMap[key].Summaries = append(dataMap[key].Summaries, summary)
		}
	}

	if len(dataMap) == 0 {
		return nil, errors.New("暂无评教数据")
	}

	type teacherScore struct {
		key      string
		avgScore float64
	}
	var scores []teacherScore
	for key, td := range dataMap {
		scores = append(scores, teacherScore{key: key, avgScore: td.TotalScore / float64(td.Count)})
	}

	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].avgScore > scores[i].avgScore {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	var details []TeacherEvaluationDetail
	for rank, s := range scores {
		td := dataMap[s.key]
		detail := TeacherEvaluationDetail{
			TeacherName:   td.TeacherName,
			WorkNo:        td.WorkNo,
			CourseName:    td.CourseName,
			ClassName:     td.ClassName,
			AvgScore:      td.TotalScore / float64(td.Count),
			Rank:          rank + 1,
			TotalTeachers: len(scores),
			Comments:      td.Comments,
			Summaries:     td.Summaries,
		}
		details = append(details, detail)
	}

	return details, nil
}

// parseDetailScores 解析 Detail 字段的 JSON 数组字符串
func parseDetailScores(detail string) []int {
	if len(detail) < 2 {
		return nil
	}
	detail = detail[1 : len(detail)-1]
	if detail == "" {
		return nil
	}

	var scores []int
	for _, s := range splitAndTrim(detail, ",") {
		var score int
		for _, c := range s {
			if c >= '0' && c <= '9' {
				score = score*10 + int(c-'0')
			}
		}
		scores = append(scores, score)
	}
	return scores
}

// splitAndTrim 分割字符串并去除空白
func splitAndTrim(s string, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}
