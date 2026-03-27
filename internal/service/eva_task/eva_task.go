package eva_task

import (
	context "context"
	eva_task2 "edu-evaluation-backed/api/v1/eva_task"
	"edu-evaluation-backed/internal/biz/eva_task"
	"edu-evaluation-backed/internal/data/dal"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
)

// EvaTaskService 评教任务服务
// 提供评教任务的创建、详情查询、列表查询、状态修改等功能
type EvaTaskService struct {
	taskDal *dal.TaskDal
	taskUC  *eva_task.EvaTaskUseCase
}

func (e EvaTaskService) SubmitEvaluation(ctx context.Context, req *eva_task2.SubmitEvaluationReq) (*eva_task2.SubmitEvaluationResp, error) {
	err := e.taskDal.SubmitEvaluation(uint(req.TaskId), uint(req.CourseId), uint(req.TeacherId), req.StuNo, req.DetailScore, req.Comment, int(req.Score))
	if err != nil {
		return nil, err
	}
	return &eva_task2.SubmitEvaluationResp{
		Message: "提交成功",
	}, nil
}

// StudentTaskDetail 获取学生评教任务详情
func (e EvaTaskService) StudentTaskDetail(ctx context.Context, req *eva_task2.StuTaskDetailReq) (*eva_task2.StuTaskDetailRes, error) {
	c, err := e.taskDal.StudentTaskDetail(req.StuNo, uint(req.TaskId))
	log.Info(c)
	if err != nil {
		return nil, err
	}
	cour := make([]*eva_task2.CourseInfo, 0)

	for _, c := range c {
		cour = append(cour, &eva_task2.CourseInfo{
			Id:      strconv.Itoa(int(c.ID)),
			Name:    c.CourseName + " - " + c.ClassName,
			Teacher: make([]*eva_task2.CourseInfo_TeacherInfo, 0),
		})
		for _, t := range c.Teachers {
			// 去查询该老师是否已评价过
			det, _ := e.taskDal.GetTaskEvaluationDetail(uint(req.TaskId), c.ID, req.StuNo, t.ID)
			cour[len(cour)-1].Teacher = append(cour[len(cour)-1].Teacher, &eva_task2.CourseInfo_TeacherInfo{
				Id:            strconv.Itoa(int(t.ID)),
				Name:          t.Name,
				HasEvaluation: det.ID != 0,
			})
		}
	}
	resp := &eva_task2.StuTaskDetailRes{
		Message: "success",
		Course:  cour,
	}
	return resp, nil
}

// CreateTask 创建评教任务
// 根据名称和选定的课程ID列表创建一个新的评教任务
// ctx: 上下文
// req: 创建评教任务请求，包含任务名称和课程ID列表
// 返回值: 创建成功响应，包含新创建的任务ID，错误信息
func (e EvaTaskService) CreateTask(ctx context.Context, req *eva_task2.CreateTaskReq) (*eva_task2.CreateTaskResp, error) {
	id, err := e.taskUC.CreateEvaTask(req.Name, req.CourseIds)
	if err != nil {
		return nil, err
	}
	resp := &eva_task2.CreateTaskResp{
		Data: &eva_task2.CreateTaskRespD{
			Id: strconv.Itoa(int(id)),
		},
		Message: "创建成功",
	}
	return resp, nil
}

// Detail 获取评教任务详情
// 根据任务ID获取评教任务的详细信息，包含每个课程的已评教人数、评教平均分等信息
// ctx: 上下文
// req: 获取任务详情请求，包含任务ID（字符串格式）
// 返回值: 评教任务详情，包含每个课程的统计信息，错误信息
func (e EvaTaskService) Detail(ctx context.Context, req *eva_task2.GetTaskReq) (*eva_task2.TaskInfo, error) {
	taskID, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, err
	}
	task, err := e.taskUC.GetTaskDetail(uint(taskID))
	if err != nil {
		return nil, err
	}

	// 转换为proto结构
	var courses []*eva_task2.CourseInfo
	for _, course := range task.Courses {
		courseInfo := &eva_task2.CourseInfo{
			Id:              strconv.Itoa(int(course.ID)),
			Name:            course.CourseName + " - " + course.ClassName,
			EvaluationScore: int32(course.EvaluationScore),
			EvaluationNum:   int32(course.EvaluationNum),
			TotalNum:        int32(len(course.Students)),
		}
		courses = append(courses, courseInfo)
	}

	resp := &eva_task2.TaskInfo{
		Id:     strconv.Itoa(int(task.ID)),
		Name:   task.Title,
		Status: int32(task.Status),
		Course: courses,
	}
	return resp, nil
}

// List 获取评教任务列表
// 支持分页查询，并可以按状态筛选
// ctx: 上下文
// req: 获取任务列表请求，包含页码、每页条数、状态筛选条件
// 返回值: 评教任务列表响应，包含数据列表和总数，错误信息
func (e EvaTaskService) List(ctx context.Context, req *eva_task2.GetTaskListReq) (*eva_task2.GetTaskListResp, error) {
	tasks, total, err := e.taskUC.GetTaskList(int(req.Page), int(req.PageSize), int(req.Status))
	if err != nil {
		return nil, err
	}
	var taskInfos []*eva_task2.TaskInfo
	for _, task := range *tasks {
		taskInfo := &eva_task2.TaskInfo{
			Id:     strconv.Itoa(int(task.ID)),
			Name:   task.Title,
			Status: int32(task.Status),
		}
		taskInfos = append(taskInfos, taskInfo)
	}

	return &eva_task2.GetTaskListResp{
		Message: "success",
		Data: &eva_task2.GetTaskListRespD{
			Total: total,
			Tasks: taskInfos,
		},
	}, nil
}

// ChangeStatus 修改评教任务状态
// 修改指定评教任务的状态（进行中/已结束）
// ctx: 上下文
// req: 修改任务状态请求，包含任务ID和新状态
// 返回值: 修改成功响应，错误信息
func (e EvaTaskService) ChangeStatus(ctx context.Context, req *eva_task2.ChangeTaskStatusReq) (*eva_task2.ChangeTaskStatusResp, error) {
	err := e.taskUC.ChangeTaskStatus(uint(req.Id), int(req.Status))
	if err != nil {
		return nil, err
	}
	return &eva_task2.ChangeTaskStatusResp{
		Message: "修改成功",
	}, nil
}

// NewEvaTaskService 创建评教任务服务实例
// taskDal: 评教任务数据访问层
// taskUC: 评教任务业务用例
// 返回值: 评教任务服务实例指针
func NewEvaTaskService(taskDal *dal.TaskDal, taskUC *eva_task.EvaTaskUseCase) *EvaTaskService {
	return &EvaTaskService{
		taskDal: taskDal,
		taskUC:  taskUC,
	}
}
