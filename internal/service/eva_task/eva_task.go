package eva_task

import (
	context "context"
	eva_task2 "edu-evaluation-backed/api/v1/eva_task"
	"edu-evaluation-backed/internal/biz/eva_task"
	"edu-evaluation-backed/internal/data/dal"
	"strconv"
)

type EvaTaskService struct {
	taskDal *dal.TaskDal
	taskUC  *eva_task.EvaTaskUseCase
}

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
	var courses []*eva_task2.TaskInfo_CourseInfo
	for _, course := range task.Courses {
		courseInfo := &eva_task2.TaskInfo_CourseInfo{
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

func (e EvaTaskService) ChangeStatus(ctx context.Context, req *eva_task2.ChangeTaskStatusReq) (*eva_task2.ChangeTaskStatusResp, error) {
	err := e.taskUC.ChangeTaskStatus(uint(req.Id), int(req.Status))
	if err != nil {
		return nil, err
	}
	return &eva_task2.ChangeTaskStatusResp{
		Message: "修改成功",
	}, nil
}

func NewEvaTaskService(taskDal *dal.TaskDal, taskUC *eva_task.EvaTaskUseCase) *EvaTaskService {
	return &EvaTaskService{
		taskDal: taskDal,
		taskUC:  taskUC,
	}
}
