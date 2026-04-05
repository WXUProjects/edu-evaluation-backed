// Package base_info 提供基础信息（课程、学生、教师）的业务逻辑层实现。
package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"fmt"
	"mime/multipart"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xuri/excelize/v2"
)

// CourseUseCase 课程信息业务用例，负责课程数据的导入、删除等业务操作。
type CourseUseCase struct {
	courseDal *dal.CourseDal
}

// courseItem 课程分组键，用于按课程名称和班级名称进行分组。
type courseItem struct {
	courseName string
	className  string
}

// Import 从 Excel 文件导入课程数据。
// 解析 Excel 中的 Sheet1，按课程名称和班级名称分组后批量创建课程并关联学生。
// 返回导入过程中的错误信息汇总字符串；无错误时返回空字符串。
func (c CourseUseCase) Import(f multipart.File) string {
	list, err := excelize.OpenReader(f)
	if err != nil {
		return err.Error()
	}
	defer list.Close()

	rows, err := list.GetRows("Sheet1")
	if err != nil {
		return err.Error()
	}

	// 按课程名称+班级名称分组
	courseMap := make(map[courseItem][]string)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		item := courseItem{courseName: row[1], className: row[2]}
		courseMap[item] = append(courseMap[item], row[3])
	}

	var logMsg string
	for k, v := range courseMap {
		id, err := c.courseDal.CreateCourse(k.courseName, k.className)
		if err != nil {
			logMsg += fmt.Sprintf("课程:%s,班级:%s,错误:班级已存在\n", k.courseName, k.className)
			continue
		}
		if err := c.courseDal.AddStudent(id, v); err != nil {
			log.Info(err)
			logMsg += fmt.Sprintf("课程:%s,班级:%s,添加学生错误:%s\n", k.courseName, k.className, err.Error())
		}
	}
	return logMsg
}

// DeleteCourse 根据课程 ID 删除指定课程。
// 参数 id 为目标课程的主键 ID。
// 删除成功返回 nil，否则返回对应的错误。
func (c CourseUseCase) DeleteCourse(id uint) error {
	return c.courseDal.DeleteCourse(id)
}

// NewCourseUseCase 创建并返回 CourseUseCase 实例。
// 参数 courseDal 为课程数据访问层对象。
func NewCourseUseCase(courseDal *dal.CourseDal) *CourseUseCase {
	return &CourseUseCase{courseDal: courseDal}
}
