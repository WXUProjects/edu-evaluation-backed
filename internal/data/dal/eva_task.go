package dal

import (
	"edu-evaluation-backed/internal/common/utils"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"

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

// NewTaskDal 创建评教任务数据访问层实例
// data: 数据层上下文，包含数据库连接和Redis客户端
// 返回值: 评教任务数据访问层实例指针
func NewTaskDal(data *data.Data) *TaskDal {
	return &TaskDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}
