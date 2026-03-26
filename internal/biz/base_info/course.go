package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"fmt"
	"mime/multipart"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xuri/excelize/v2"
)

type CourseUseCase struct {
	courseDal *dal.CourseDal
}
type courseItem struct {
	courseName string
	className  string
}

// Import 导入课程
func (c CourseUseCase) Import(f multipart.File) string {
	list, err := excelize.OpenReader(f)
	if err != nil {
		return err.Error()
	}
	defer func() {
		_ = list.Close()
	}()
	rows, err := list.GetRows("Sheet1")
	if err != nil {
		return err.Error()
	}
	iLog := ""
	courseClass := make(map[courseItem][]string)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		t := courseItem{
			courseName: row[1],
			className:  row[2],
		}
		courseClass[t] = append(courseClass[t], row[4])
	}
	for k, v := range courseClass {
		id, err := c.courseDal.CreateCourse(k.courseName, k.className)
		if err != nil {
			iLog += fmt.Sprintf("课程:%s,班级:%s,错误:%s\n", k.courseName, k.className, "已经存在此班级")
			continue
		}
		err = c.courseDal.AddStudent(id, v)
		log.Info(err)
	}
	return iLog
}
// DeleteCourse 删除课程
func (c CourseUseCase) DeleteCourse(id uint) error {
	return c.courseDal.DeleteCourse(id)
}

func NewCourseUseCase(courseDal *dal.CourseDal) *CourseUseCase {
	return &CourseUseCase{
		courseDal: courseDal,
	}
}
