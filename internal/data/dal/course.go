package dal

import (
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CourseDal struct {
	db  *gorm.DB
	rdb *redis.Client
}

// Detail 获取课程详情
func (c CourseDal) Detail(courseID uint) (*model.Course, error) {
	course := model.Course{}
	err := c.db.Where("id = ?", courseID).Preload("Teachers").Preload("Students").First(&course).Error
	return &course, err
}

// CreateCourse 创建课程
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

// List 获取课程列表
func (c CourseDal) List(page, pageSize int) (*[]model.Course, int64, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	var courses []model.Course
	var tot int64
	err := c.db.Limit(pageSize).Preload("Teachers").Offset((page - 1) * pageSize).Find(&courses).Count(&tot).Error
	return &courses, tot, err
}

// UpdateCourse 更新课程信息
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

// AddTeachers 添加教师到课程
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

func NewCourseDal(data *data.Data) *CourseDal {
	return &CourseDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}
