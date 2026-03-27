package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"fmt"
	"mime/multipart"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xuri/excelize/v2"
)

// CourseUseCase 课程信息业务用例
// 处理课程信息相关的业务逻辑，包括Excel导入、删除课程
type CourseUseCase struct {
	courseDal *dal.CourseDal
}

// courseItem 课程分组键
// 由课程名称和班级名称唯一确定一个课程
type courseItem struct {
	courseName string
	className  string
}

// Import 从Excel文件导入课程数据
// f: 上传的Excel文件句柄
// 从Sheet1读取数据，第一行为表头，从第二行开始为数据
// Excel格式：第一列课程名称，第二列班级名称，第四列学生学号
// 相同课程名称+班级名称的学生会自动分组合并到同一个课程中
// 返回值: 导入过程中的错误日志字符串，为空表示没有错误
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
	// 按课程名称+班级名称分组，收集所有学生学号
	courseClass := make(map[courseItem][]string)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		t := courseItem{
			courseName: row[1],
			className:  row[2],
		}
		courseClass[t] = append(courseClass[t], row[3])
	}
	// 为每个分组创建课程并添加学生
	for k, v := range courseClass {
		id, err := c.courseDal.CreateCourse(k.courseName, k.className)
		if err != nil {
			iLog += fmt.Sprintf("课程:%s,班级:%s,错误:%s\n", k.courseName, k.className, "已经存在此班级")
			continue
		}
		err = c.courseDal.AddStudent(id, v)
		log.Info(err)
		if err != nil {
			iLog += fmt.Sprintf("课程:%s,班级:%s,添加学生错误:%s\n", k.courseName, k.className, err.Error())
		}
	}
	return iLog
}

// DeleteCourse 删除课程
// id: 要删除的课程ID
// 删除前会清除课程与所有教师和学生的关联，关联清除后删除课程本身
// 返回值: 删除成功返回nil，错误信息
func (c CourseUseCase) DeleteCourse(id uint) error {
	return c.courseDal.DeleteCourse(id)
}

// NewCourseUseCase 创建课程信息业务用例实例
// courseDal: 课程数据访问层
// 返回值: 课程信息业务用例实例指针
func NewCourseUseCase(courseDal *dal.CourseDal) *CourseUseCase {
	return &CourseUseCase{
		courseDal: courseDal,
	}
}
