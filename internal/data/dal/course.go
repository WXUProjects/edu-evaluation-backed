package dal

import (
	"edu-evaluation-backed/internal/common/utils"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CourseDal 课程数据访问层
// 处理课程相关的数据库操作，包括详情查询、创建、更新、删除、列表查询等
type CourseDal struct {
	db  *gorm.DB
	rdb *redis.Client
}

// Detail 获取课程详情，预加载教师和学生关联信息
// courseID: 课程ID
// 返回值: 课程信息指针，错误信息
func (c CourseDal) Detail(courseID uint) (*model.Course, error) {
	course := model.Course{}
	err := c.db.Where("id = ?", courseID).Preload("Teachers").Preload("Students").First(&course).Error
	return &course, err
}

// CreateCourse 创建新课程
// courseName: 课程名称
// className: 班级名称（唯一标识）
// 创建时默认状态为1（正常）
// 返回值: 新创建的课程ID，错误信息
func (c CourseDal) CreateCourse(courseName, className string) (uint, error) {
	// 班级名称是唯一的
	cs := model.Course{
		Status:     1,
		CourseName: courseName,
		ClassName:  className,
	}
	err := c.db.Create(&cs)
	return cs.ID, err.Error
}

// AddStudent 添加学生到课程
// courseID: 课程ID
// studentNos: 要添加的学生学号列表
// 根据学号列表查询学生，然后添加到课程的学生关联中
// 返回值: 添加成功返回nil，错误信息（如未找到学生）
func (c CourseDal) AddStudent(courseID uint, studentNos []string) error {
	course := model.Course{}
	course.ID = courseID
	var students []model.Student
	if err := c.db.Where("student_no IN ?", studentNos).Find(&students).Error; err != nil {
		return err
	}
	if len(students) == 0 {
		return errors.New("未找到匹配的学生信息")
	}
	err := c.db.Model(&course).Association("Students").Append(&students)
	if err != nil {
		return err
	}
	return nil
}

// List 获取课程列表，支持分页，预加载教师关联信息
// page: 当前页码，pageSize: 每页条数
// 返回值: 课程列表指针，总记录数，错误信息
func (c CourseDal) List(page, pageSize int) (*[]model.Course, int64, error) {
	page, pageSize = utils.PageNumHandle(page, pageSize)
	var courses []model.Course
	var tot int64
	err := c.db.Model(&model.Course{}).Count(&tot).Limit(pageSize).Preload("Teachers").Offset(utils.CalculateOffset(page, pageSize)).Find(&courses).Error
	return &courses, tot, err
}

// QueryCourseByIds 批量获取课程信息根据ID列表
// ids: 课程ID列表
// 返回值: 课程列表指针，错误信息
func (c CourseDal) QueryCourseByIds(ids []int32) (*[]model.Course, error) {
	var courses []model.Course
	err := c.db.Where("id IN ?", ids).Find(&courses).Error
	return &courses, err
}

// UpdateCourse 更新课程基本信息
// courseID: 课程ID
// courseName: 新课程名称，为空不更新
// className: 新班级名称，为空不更新
// 如果更新班级名称，会检查是否与其他课程冲突
// 返回值: 更新成功返回nil，错误信息
func (c CourseDal) UpdateCourse(courseID uint, courseName, className string) error {
	// 检查className是否已被其他课程使用
	var count int64
	err := c.db.Model(&model.Course{}).Where("class_name = ? AND id != ?", className, courseID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("班级名称已存在")
	}

	updates := make(map[string]interface{})
	if courseName != "" {
		updates["course_name"] = courseName
	}
	if className != "" {
		updates["class_name"] = className
	}

	if len(updates) == 0 {
		return nil
	}

	err = c.db.Model(&model.Course{}).Where("id = ?", courseID).Updates(updates).Error
	return err
}

// AddTeachers 绑定教师到课程（先清除原有绑定再重新绑定）
// courseID: 课程ID
// teacherWorkNos: 教师ID列表
// 先清除课程原有的所有教师关联，然后添加新的教师关联
// 返回值: 添加成功返回nil，错误信息（如未找到教师）
func (c CourseDal) AddTeachers(courseID uint, teacherWorkNos []int32) error {
	// 第一步，清除课程的教师关联
	err := c.db.Model(&model.Course{Model: gorm.Model{ID: courseID}}).
		Association("Teachers").
		Clear()
	if err != nil {
		return err
	}
	course := model.Course{}
	course.ID = courseID
	var teachers []model.Teacher
	if err := c.db.Where("id IN ?", teacherWorkNos).Find(&teachers).Error; err != nil {
		return err
	}
	if len(teachers) == 0 {
		return errors.New("未找到匹配的教师信息")
	}
	err = c.db.Model(&course).Association("Teachers").Append(&teachers)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCourse 删除课程（清除教师和学生关联后删除）
func (c CourseDal) DeleteCourse(id uint) error {
	var course model.Course
	if err := c.db.First(&course, id).Error; err != nil {
		return err
	}

	// 清除教师关联，教师保留
	if err := c.db.Model(&course).Association("Teachers").Clear(); err != nil {
		return err
	}

	// 清除学生关联，学生保留
	if err := c.db.Model(&course).Association("Students").Clear(); err != nil {
		return err
	}

	// 删除课程本身
	if err := c.db.Delete(&course).Error; err != nil {
		return err
	}

	return nil
}

// NewCourseDal 创建课程数据访问层实例
// data: 数据层上下文，包含数据库连接和Redis客户端
// 返回值: 课程数据访问层实例指针
func NewCourseDal(data *data.Data) *CourseDal {
	return &CourseDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}
