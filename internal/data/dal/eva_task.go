package dal

import (
	"edu-evaluation-backed/internal/common/utils"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TaskDal 评教任务数据访问层
// 处理评教任务相关的数据库操作，包括创建、查询列表、查询详情、修改状态
type TaskDal struct {
	db  *gorm.DB
	rdb *redis.Client
}

// CreateTask 创建评教任务
// title: 评教任务名称
// courses: 参与评教的课程列表
// 创建时初始状态为0（未开始）
// 返回值: 新创建的评教任务ID，错误信息
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

// GetTaskList 获取评教任务列表
// page: 当前页码，pageSize: 每页条数
// status: 状态筛选，-1表示不筛选，只返回指定状态的任务
// 返回值: 评教任务列表指针，总记录数，错误信息
// 结果按ID降序排列，保证最新的任务排在前面
func (d *TaskDal) GetTaskList(page, pageSize, status int) (*[]model.EvaluationTask, int64, error) {
	var tasks []model.EvaluationTask
	var total int64
	page, pageSize = utils.PageNumHandle(page, pageSize)
	baseQ := d.db.Model(&model.EvaluationTask{})
	if status != -1 {
		baseQ = baseQ.Where("status = ?", status)
	}
	err := baseQ.Count(&total).Order("id desc").Limit(pageSize).Offset(utils.CalculateOffset(page, pageSize)).Find(&tasks).Error
	return &tasks, total, err
}

// GetTaskDetail 获取评教任务详情
// taskID: 评教任务ID
// 预加载课程列表，以及课程的学生和教师关联信息
// 返回值: 评教任务信息指针，错误信息
func (d *TaskDal) GetTaskDetail(taskID uint) (*model.EvaluationTask, error) {
	var task model.EvaluationTask
	err := d.db.Where("id = ?", taskID).Preload("Courses").Preload("Courses.Students").Preload("Courses.Teachers").First(&task).Error
	return &task, err
}

// ChangeTaskStatus 修改评教任务状态
// taskID: 评教任务ID
// status: 新状态值
// 直接更新任务的status字段
// 返回值: 修改成功返回nil，错误信息
func (d *TaskDal) ChangeTaskStatus(taskID uint, status int) error {
	err := d.db.Model(&model.EvaluationTask{}).Where("id = ?", taskID).Update("status", status).Error
	return err
}

func (d *TaskDal) StudentTaskDetail(studentNo string, taskID uint) ([]model.Course, error) {
	var courses []model.Course

	// 1. 直接查询 Course 表
	err := d.db.Debug().Model(&model.Course{}).
		// 预加载老师信息（这是你需要的）
		Preload("Teachers").
		// 关键：关联评价任务表并过滤 taskID
		Joins("JOIN evaluation_courses ec ON ec.course_id = courses.id").
		// 关键：关联学生选课表并过滤 studentNo
		Joins("JOIN course_students cs ON cs.course_id = courses.id").
		Where("ec.evaluation_task_id = ? AND cs.student_student_no = ?", taskID, studentNo).
		Find(&courses).Error

	return courses, err
}

// GetTaskEvaluationDetail 获取任务评价详情
func (d *TaskDal) GetTaskEvaluationDetail(taskID uint, courseID uint, studentNo string, teacherId uint) (model.EvaluationDetail, error) {
	// 根据 studentNo 获取学生ID
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

// SubmitEvaluation 提交评价
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
	return d.db.Create(&dc).Error
}

// NewTaskDal 创建评教任务数据访问层实例
// data: 数据层上下文，包含数据库连接和Redis客户端
// 返回值: 评教任务数据访问层实例指针
func NewTaskDal(data *data.Data) *TaskDal {
	return &TaskDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}

// TeacherEvaluationResult 教师评教结果（用于导出）
type TeacherEvaluationResult struct {
	TeacherName    string   // 教师姓名
	WorkNo        string   // 教师工号
	CourseName    string   // 课程名称
	ClassName     string   // 班级名称
	TotalScore    float64  // 总分
	AvgScore      float64  // 平均分
	Rank          int     // 排名
	QuestionScores [][]int  // 每道题的分数组（每个学生一行）
}

// TeacherEvaluationDetail 教师评教详情（用于生成PDF）
type TeacherEvaluationDetail struct {
	TeacherName    string   // 教师姓名
	WorkNo        string   // 教师工号
	CourseName     string   // 课程名称
	ClassName     string   // 班级名称
	AvgScore      float64  // 平均分
	Rank          int      // 排名
	TotalTeachers int      // 总教师数
	Comments      []string // 学生评价列表
	Summary       string   // 学生总结
}

// GetTaskEvaluationResults 获取任务下所有教师的评教结果
func (d *TaskDal) GetTaskEvaluationResults(taskID uint) ([]TeacherEvaluationResult, error) {
	var results []TeacherEvaluationResult

	// 使用 SQL 关联查询获取评教详情及教师、课程信息
	rows, err := d.db.Raw(`
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
		WHERE ed.task_id = ?
	`, taskID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 按教师+课程分组
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

	// 构建结果
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
// 按教师+班级分组，返回每个组合的评教详情
func (d *TaskDal) GetTeacherEvaluationDetailsForPDF(taskID uint) ([]TeacherEvaluationDetail, error) {
	// 查询每个教师-课程组合的评教数据
	rows, err := d.db.Raw(`
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
		WHERE ed.task_id = ?
		GROUP BY t.id, t.name, t.work_no, c.id, c.course_name, c.class_name, ed.detail, ed.summary
		ORDER BY t.name, c.class_name, avg_score DESC
	`, taskID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 按教师+班级分组收集数据
	type teacherCourseData struct {
		TeacherName  string
		WorkNo      string
		CourseName  string
		ClassName   string
		TotalScore  float64
		AvgScore    float64
		Count       int
		Comments    []string
		Summary     string
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
			}
		}

		dataMap[key].TotalScore += avgScore
		dataMap[key].Count++
		if detail != "" {
			dataMap[key].Comments = append(dataMap[key].Comments, detail)
		}
		if summary != "" {
			dataMap[key].Summary = summary
		}
	}

	if len(dataMap) == 0 {
		return nil, errors.New("暂无评教数据")
	}

	// 计算每个教师的平均分并排序
	type teacherScore struct {
		key      string
		avgScore float64
	}
	var scores []teacherScore
	for key, td := range dataMap {
		td.AvgScore = td.TotalScore / float64(td.Count)
		scores = append(scores, teacherScore{key: key, avgScore: td.AvgScore})
	}

	// 按平均分降序排序
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].avgScore > scores[i].avgScore {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// 构建结果
	var details []TeacherEvaluationDetail
	for rank, s := range scores {
		td := dataMap[s.key]
		detail := TeacherEvaluationDetail{
			TeacherName:    td.TeacherName,
			WorkNo:        td.WorkNo,
			CourseName:    td.CourseName,
			ClassName:     td.ClassName,
			AvgScore:      td.AvgScore,
			Rank:          rank + 1,
			TotalTeachers: len(scores),
			Comments:      td.Comments,
			Summary:       td.Summary,
		}
		details = append(details, detail)
	}

	return details, nil
}

// parseDetailScores 解析 Detail 字段的 JSON 数组字符串
func parseDetailScores(detail string) []int {
	// Detail 格式类似 "[1,2,3,4,5,3,2]"
	if len(detail) < 2 {
		return nil
	}
	// 去掉首尾的 []
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
