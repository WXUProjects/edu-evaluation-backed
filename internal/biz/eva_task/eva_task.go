package eva_task

import (
	"fmt"
	"os"
	"strconv"

	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"

	"github.com/xuri/excelize/v2"
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

// GetTaskEvaluationResults 获取任务评教结果（用于导出）
func (e *EvaTaskUseCase) GetTaskEvaluationResults(taskID uint) ([]dal.TeacherEvaluationResult, error) {
	return e.taskDal.GetTaskEvaluationResults(taskID)
}

// ExportTaskResults 导出任务评教结果为 xlsx
// taskID: 评教任务ID
// 返回值: xlsx文件路径，错误信息
func (e *EvaTaskUseCase) ExportTaskResults(taskID uint) (string, error) {
	// 获取评教结果数据
	results, _ := e.GetTaskEvaluationResults(taskID)

	// 确保tmp目录存在
	os.MkdirAll("./tmp", 0755)

	// ========== 生成 xlsx ==========
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "评教结果"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 8)   // 序号
	f.SetColWidth(sheetName, "B", "B", 15)  // 工号
	f.SetColWidth(sheetName, "C", "C", 25)  // 教师姓名
	f.SetColWidth(sheetName, "D", "D", 20)  // 课程
	f.SetColWidth(sheetName, "E", "E", 15)  // 班级名

	// 计算最大题目数
	maxQuestions := 0
	for _, r := range results {
		for _, scores := range r.QuestionScores {
			if len(scores) > maxQuestions {
				maxQuestions = len(scores)
			}
		}
	}

	// 设置样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:   []excelize.Border{{Type: "left", Style: 1}, {Type: "right", Style: 1}, {Type: "top", Style: 1}, {Type: "bottom", Style: 1}},
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:   []excelize.Border{{Type: "left", Style: 1}, {Type: "right", Style: 1}, {Type: "top", Style: 1}, {Type: "bottom", Style: 1}},
	})

	// 写入表头
	headers := []interface{}{
		"序号",
		"工号",
		"教师姓名",
		"课程",
		"班级名",
		"平均分",
	}
	for i := 1; i <= maxQuestions; i++ {
		headers = append(headers, "问题"+strconv.Itoa(i))
	}
	headerRow := 1
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, headerRow)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 写入数据
	rowNum := 2
	for idx, r := range results {
		questionAvgs := make([]float64, maxQuestions)
		for q := 0; q < maxQuestions; q++ {
			var total float64
			count := 0
			for _, scores := range r.QuestionScores {
				if q < len(scores) {
					total += float64(scores[q])
					count++
				}
			}
			if count > 0 {
				questionAvgs[q] = total / float64(count)
			}
		}

		row := []interface{}{
			idx + 1,
			r.WorkNo,
			r.TeacherName,
			r.CourseName,
			r.ClassName,
			r.AvgScore,
		}
		for _, avg := range questionAvgs {
			if avg > 0 {
				row = append(row, fmt.Sprintf("%.1f", avg))
			} else {
				row = append(row, "-")
			}
		}

		for col, val := range row {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowNum)
			f.SetCellValue(sheetName, cell, val)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
		rowNum++
	}

	xlsxPath := "./tmp/评教结果.xlsx"
	if err := f.SaveAs(xlsxPath); err != nil {
		return "", err
	}

	return xlsxPath, nil
}
