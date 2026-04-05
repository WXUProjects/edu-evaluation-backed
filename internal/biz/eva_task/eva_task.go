package eva_task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xuri/excelize/v2"
)

// EvaTaskUseCase 评教任务业务用例
type EvaTaskUseCase struct {
	baseDal   *dal.BaseInfoDal
	courseDal *dal.CourseDal
	taskDal   *dal.TaskDal
}

// CreateEvaTask 创建评教任务
func (e EvaTaskUseCase) CreateEvaTask(title string, courses []int32) (int32, error) {
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
func (e EvaTaskUseCase) GetTaskList(page int, pageSize int, status int) (*[]model.EvaluationTask, int64, error) {
	return e.taskDal.GetTaskList(page, pageSize, status)
}

// GetTaskDetail 获取评教任务详情
func (e EvaTaskUseCase) GetTaskDetail(taskID uint) (*model.EvaluationTask, error) {
	return e.taskDal.GetTaskDetail(taskID)
}

// ChangeTaskStatus 修改评教任务状态
func (e EvaTaskUseCase) ChangeTaskStatus(taskID uint, status int) error {
	task, err := e.taskDal.GetTaskDetail(taskID)
	if err != nil {
		return err
	}
	oldStatus := task.Status

	if err := e.taskDal.ChangeTaskStatus(taskID, status); err != nil {
		return err
	}

	if status == 2 {
		for _, course := range task.Courses {
			if err := e.courseDal.UpdateCourseStatus(course.ID, 2); err != nil {
				return err
			}
		}
	}

	if oldStatus == 2 && status == 0 {
		courseIDs := make([]uint, 0, len(task.Courses))
		for _, course := range task.Courses {
			courseIDs = append(courseIDs, course.ID)
		}

		if err := e.taskDal.DeleteTaskDetails(taskID, courseIDs); err != nil {
			return err
		}

		for _, course := range task.Courses {
			if err := e.courseDal.UpdateCourseStatus(course.ID, 1); err != nil {
				return err
			}
			if err := e.courseDal.ResetEvaluationStats(course.ID); err != nil {
				return err
			}
		}
	}

	// 当任务状态变为"进行中"时，异步预热缓存
	if status == 1 {
		go e.PreheatTask(taskID)
	}

	return nil
}

// NewEvaTaskUseCase 创建评教任务业务用例实例
func NewEvaTaskUseCase(baseDal *dal.BaseInfoDal, evaTaskDal *dal.TaskDal, courseDal *dal.CourseDal) *EvaTaskUseCase {
	return &EvaTaskUseCase{
		baseDal:   baseDal,
		taskDal:   evaTaskDal,
		courseDal: courseDal,
	}
}

// GetTaskEvaluationResults 获取任务评教结果
func (e EvaTaskUseCase) GetTaskEvaluationResults(taskID uint) ([]dal.TeacherEvaluationResult, error) {
	return e.taskDal.GetTaskEvaluationResults(taskID)
}

// ExportTaskResults 导出任务评教结果
func (e EvaTaskUseCase) ExportTaskResults(taskID uint) (*ExportResult, error) {
	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)
	tmpDir := filepath.Join(baseDir, "tmp")
	resDir := filepath.Join(baseDir, "res")

	os.MkdirAll(tmpDir, 0755)
	os.MkdirAll(resDir, 0755)

	results, err := e.GetTaskEvaluationResults(taskID)
	if err != nil {
		return nil, err
	}

	xlsxPath := filepath.Join(tmpDir, "评教结果.xlsx")
	if err := generateXlsx(results, xlsxPath); err != nil {
		return nil, err
	}

	details, err := e.taskDal.GetTeacherEvaluationDetailsForPDF(taskID)
	if err != nil {
		return nil, err
	}

	pdfPaths := generateAllPDFs(details, tmpDir)

	zipName := fmt.Sprintf("%d_%s.zip", time.Now().UnixNano(), randomString(8))
	zipPath := filepath.Join(resDir, zipName)
	allFiles := []string{"评教结果.xlsx"}
	for _, p := range pdfPaths {
		parts := strings.Split(p, string(filepath.Separator))
		allFiles = append(allFiles, parts[len(parts)-1])
	}
	if err := zipFiles(tmpDir, zipPath, allFiles...); err != nil {
		return nil, err
	}

	return &ExportResult{
		XlsxPath: xlsxPath,
		PdfPaths: pdfPaths,
		ZipPath:  fmt.Sprintf("res/%s", zipName),
	}, nil
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// generateXlsx 生成 xlsx 文件
func generateXlsx(results []dal.TeacherEvaluationResult, xlsxPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "评教结果"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	f.SetColWidth(sheetName, "A", "A", 8)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 25)
	f.SetColWidth(sheetName, "D", "D", 20)
	f.SetColWidth(sheetName, "E", "E", 15)

	maxQuestions := 0
	for _, r := range results {
		for _, scores := range r.QuestionScores {
			if len(scores) > maxQuestions {
				maxQuestions = len(scores)
			}
		}
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    []excelize.Border{{Type: "left", Style: 1}, {Type: "right", Style: 1}, {Type: "top", Style: 1}, {Type: "bottom", Style: 1}},
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    []excelize.Border{{Type: "left", Style: 1}, {Type: "right", Style: 1}, {Type: "top", Style: 1}, {Type: "bottom", Style: 1}},
	})

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

	return f.SaveAs(xlsxPath)
}

// PreheatTask 预热任务相关缓存
// 预热内容：
// 1. 任务列表前3页（status=1）
// 2. 遍历任务关联课程 → 所有学生 → 预缓存 task:courses
// 3. 预设所有 eval:check 为未评教状态
func (e EvaTaskUseCase) PreheatTask(taskID uint) {
	log.Infof("starting cache preheat for task %d", taskID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1. 获取任务详情（含关联课程）
	task, err := e.taskDal.GetTaskDetail(taskID)
	if err != nil {
		log.Warnf("preheat: failed to get task detail: %v", err)
		return
	}

	// 2. 预热任务列表前3页
	for page := 1; page <= 3; page++ {
		// GetTaskList 内部会自动缓存
		_, _, _ = e.taskDal.GetTaskList(page, 10, 1)
	}

	// 3. 并发预热课程列表和评教状态，限制并发数
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, course := range task.Courses {
		for _, student := range course.Students {
			studentNo := student.StudentNo
			courseID := course.ID

			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

				// 预热学生课程列表缓存
				if err := e.taskDal.PreloadTaskCourses(ctx, taskID, studentNo); err != nil {
					log.Warnf("preheat: failed to preload courses for student %s: %v", studentNo, err)
					return
				}

				// 预热每个教师对应的评教状态（设置为"未评教"）
				for _, teacher := range course.Teachers {
					if err := e.taskDal.PreloadEvalCheck(ctx, taskID, courseID, studentNo, teacher.ID); err != nil {
						log.Warnf("preheat: failed to preload eval check: %v", err)
					}
				}
			}()
		}
	}

	wg.Wait()
	log.Infof("cache preheat completed for task %d", taskID)
}
