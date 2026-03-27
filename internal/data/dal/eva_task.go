package dal

import (
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type TaskDal struct {
	db  *gorm.DB
	rdb *redis.Client
}

// CreateTask 创建评价任务
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

// GetTaskList 获取评价任务列表
func (d *TaskDal) GetTaskList(page, pageSize, status int) (*[]model.EvaluationTask, int64, error) {
	var tasks []model.EvaluationTask
	var total int64
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	baseQ := d.db.Model(&model.EvaluationTask{})
	if status != -1 {
		baseQ = baseQ.Where("status = ?", status)
	}
	err := baseQ.Count(&total).Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&tasks).Error
	return &tasks, total, err
}

// GetTaskDetail 获取评价任务详情
func (d *TaskDal) GetTaskDetail(taskID uint) (*model.EvaluationTask, error) {
	var task model.EvaluationTask
	err := d.db.Where("id = ?", taskID).Preload("Courses").Preload("Courses.Students").Preload("Courses.Teachers").First(&task).Error
	return &task, err
}

// ChangeTaskStatus 修改任务状态
func (d *TaskDal) ChangeTaskStatus(taskID uint, status int) error {
	err := d.db.Model(&model.EvaluationTask{}).Where("id = ?", taskID).Update("status", status).Error
	return err
}

func NewTaskDal(data *data.Data) *TaskDal {
	return &TaskDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}
