package eva_task

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
)

type EvaTaskUseCase struct {
	baseDal   *dal.BaseInfoDal
	courseDal *dal.CourseDal
	taskDal   *dal.TaskDal
}

// CreateEvaTask 创建评价任务
func (e EvaTaskUseCase) CreateEvaTask(title string, courses []int32) (int32, error) {
	// 根据课程ID查询课程信息
	coursesInfo, err := e.courseDal.QueryCourseByIds(courses)
	if err != nil {
		return 0, err
	}
	id, err := e.taskDal.CreateTask(title, *coursesInfo)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

// GetTaskList 获取任务列表
func (e EvaTaskUseCase) GetTaskList(page int, pageSize int, status int) (*[]model.EvaluationTask, int64, error) {
	return e.taskDal.GetTaskList(page, pageSize, status)
}

// GetTaskDetail 获取任务详情
func (e EvaTaskUseCase) GetTaskDetail(taskID uint) (*model.EvaluationTask, error) {
	return e.taskDal.GetTaskDetail(taskID)
}

// ChangeTaskStatus 修改任务状态
func (e EvaTaskUseCase) ChangeTaskStatus(taskID uint, status int) error {
	return e.taskDal.ChangeTaskStatus(taskID, status)
}

func NewEvaTaskUseCase(baseDal *dal.BaseInfoDal, evaTaskDal *dal.TaskDal, courseDal *dal.CourseDal) *EvaTaskUseCase {
	return &EvaTaskUseCase{
		baseDal:   baseDal,
		taskDal:   evaTaskDal,
		courseDal: courseDal,
	}
}
