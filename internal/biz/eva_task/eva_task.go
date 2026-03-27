package eva_task

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
)

// EvaTaskUseCase 评教任务业务用例
// 处理评教任务相关的业务逻辑，包括创建任务、查询列表、查询详情、修改状态
type EvaTaskUseCase struct {
	baseDal   *dal.BaseInfoDal
	courseDal *dal.CourseDal
	taskDal   *dal.TaskDal
}

// CreateEvaTask 创建评教任务
// title: 评教任务名称
// courses: 要加入评教的课程ID列表
// 首先根据ID列表查询课程信息，然后创建评教任务并关联这些课程
// 返回值: 新创建的评教任务ID，错误信息
func (e EvaTaskUseCase) CreateEvaTask(title string, courses []int32) (int32, error) {
	// 根据课程 ID 查询课程信息
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

// GetTaskList 获取评教任务列表
// page: 当前页码，pageSize: 每页条数
// status: 状态筛选，-1表示不筛选
// 返回值: 评教任务列表指针，总记录数，错误信息
func (e EvaTaskUseCase) GetTaskList(page int, pageSize int, status int) (*[]model.EvaluationTask, int64, error) {
	return e.taskDal.GetTaskList(page, pageSize, status)
}

// GetTaskDetail 获取评教任务详情
// taskID: 评教任务ID
// 返回值: 评教任务信息，包含关联的课程列表和每个课程的评教统计信息，错误信息
func (e EvaTaskUseCase) GetTaskDetail(taskID uint) (*model.EvaluationTask, error) {
	return e.taskDal.GetTaskDetail(taskID)
}

// ChangeTaskStatus 修改评教任务状态
// taskID: 评教任务ID
// status: 新状态值（1: 进行中, 2: 已结束）
// 返回值: 修改成功返回nil，错误信息
func (e EvaTaskUseCase) ChangeTaskStatus(taskID uint, status int) error {
	return e.taskDal.ChangeTaskStatus(taskID, status)
}

// NewEvaTaskUseCase 创建评教任务业务用例实例
// baseDal: 基础信息数据访问层
// evaTaskDal: 评教任务数据访问层
// courseDal: 课程数据访问层
// 返回值: 评教任务业务用例实例指针
func NewEvaTaskUseCase(baseDal *dal.BaseInfoDal, evaTaskDal *dal.TaskDal, courseDal *dal.CourseDal) *EvaTaskUseCase {
	return &EvaTaskUseCase{
		baseDal:   baseDal,
		taskDal:   evaTaskDal,
		courseDal: courseDal,
	}
}
