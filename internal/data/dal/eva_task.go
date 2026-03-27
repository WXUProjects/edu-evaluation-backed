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
