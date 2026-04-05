package eva_task

import (
	"context"
	"strconv"

	eva_task2 "edu-evaluation-backed/api/v1/eva_task"
	"edu-evaluation-backed/internal/common/data/cache"
	"edu-evaluation-backed/internal/biz/eva_task"
	"edu-evaluation-backed/internal/data/dal"
)

// EvaTaskService 评教任务服务
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

// StudentTaskDetail 获取学生评教任务详情（使用批量 MGET 消除 N+1）
func (e EvaTaskService) StudentTaskDetail(ctx context.Context, req *eva_task2.StuTaskDetailReq) (*eva_task2.StuTaskDetailRes, error) {
	c, err := e.taskDal.StudentTaskDetail(req.StuNo, uint(req.TaskId))
	if err != nil {
		return nil, err
	}

	// 收集所有需要检查评教状态的 key 和参数
	var evalKeys []string
	var evalParams []struct {
		TaskID    uint
		CourseID  uint
		StudentNo string
		TeacherID uint
	}

	for _, course := range c {
		for _, teacher := range course.Teachers {
			key := cache.EvalCheckKey(uint(req.TaskId), course.ID, req.StuNo, teacher.ID)
			evalKeys = append(evalKeys, key)
			evalParams = append(evalParams, struct {
				TaskID    uint
				CourseID  uint
				StudentNo string
				TeacherID uint
			}{
				TaskID:    uint(req.TaskId),
				CourseID:  course.ID,
				StudentNo: req.StuNo,
				TeacherID: teacher.ID,
			})
		}
	}

	// 批量查询所有评教状态（一次 MGET）
	evalResults := make(map[string]*dal.EvalCheckData)
	if len(evalKeys) > 0 {
		evalResults, err = e.taskDal.BatchGetEvalChecks(ctx, evalKeys, evalParams)
		if err != nil {
			// 降级：批量查询失败不影响主流程
			evalResults = make(map[string]*dal.EvalCheckData)
		}
	}

	// 组装响应
	cour := make([]*eva_task2.CourseInfo, 0)
	for _, c := range c {
		cour = append(cour, &eva_task2.CourseInfo{
			Id:      strconv.Itoa(int(c.ID)),
			Name:    c.CourseName + " - " + c.ClassName,
			Teacher: make([]*eva_task2.CourseInfo_TeacherInfo, 0),
		})
		for _, t := range c.Teachers {
			key := cache.EvalCheckKey(uint(req.TaskId), c.ID, req.StuNo, t.ID)
			hasEval := false
			if evalData, ok := evalResults[key]; ok {
				hasEval = evalData.HasEval
			}
			cour[len(cour)-1].Teacher = append(cour[len(cour)-1].Teacher, &eva_task2.CourseInfo_TeacherInfo{
				Id:            strconv.Itoa(int(t.ID)),
				Name:          t.Name,
				HasEvaluation: hasEval,
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
func (e EvaTaskService) Detail(ctx context.Context, req *eva_task2.GetTaskReq) (*eva_task2.TaskInfo, error) {
	taskID, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, err
	}
	task, err := e.taskUC.GetTaskDetail(uint(taskID))
	if err != nil {
		return nil, err
	}

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
func NewEvaTaskService(taskDal *dal.TaskDal, taskUC *eva_task.EvaTaskUseCase) *EvaTaskService {
	return &EvaTaskService{
		taskDal: taskDal,
		taskUC:  taskUC,
	}
}

// ExportTaskResults 导出任务评教结果
func (e EvaTaskService) ExportTaskResults(ctx context.Context, req *eva_task2.ExportTaskResultsReq) (*eva_task2.ExportTaskResultsResp, error) {
	taskID := uint(req.TaskId)

	result, err := e.taskUC.ExportTaskResults(taskID)
	if err != nil {
		return nil, err
	}

	return &eva_task2.ExportTaskResultsResp{
		Message: "导出成功",
		ZipPath: result.ZipPath,
	}, nil
}
