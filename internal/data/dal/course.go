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

// CreateCourse 创建课程
func (c CourseDal) CreateCourse(courseName, className string) (uint, error) {
	// 班级名称是唯一的
	cs := model.Course{
		Status:     0,
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

func NewCourseDal(data *data.Data) *CourseDal {
	return &CourseDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}
