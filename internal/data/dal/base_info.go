package dal

import (
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BaseInfoDal struct {
	db  *gorm.DB
	rdb *redis.Client
}

// pageNumHandle
func pageNumHandle(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size >= 100 {
		size = 100
	}
	return page, size
}
func NewBaseInfoDal(data *data.Data) *BaseInfoDal {
	return &BaseInfoDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}

func (d *BaseInfoDal) InsertStudent(students []*model.Student) error {
	return d.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: "student_no"}},
		}).Create(students).Error
}

func (d *BaseInfoDal) QueryStudent(page, size int, studentNo, name string) (*[]model.Student, error) {
	var modelStudent []model.Student
	page, size = pageNumHandle(page, size)
	baseQ := d.db.Limit(size).Offset((page - 1) * size)
	if studentNo != "" {
		baseQ = baseQ.Where("student_no like ?", studentNo+"%")
	}
	if name != "" {
		baseQ = baseQ.Where("name like ?", "%"+name+"%")
	}
	baseQ.Order("id desc").Find(&modelStudent)
	return &modelStudent, baseQ.Error
}

func (d *BaseInfoDal) InsertTeacher(teachers []*model.Teacher) error {
	return d.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: "work_no"}},
		}).Create(teachers).Error
}

func (d *BaseInfoDal) QueryTeacher(page, size int, workNo, name string) (*[]model.Teacher, error) {
	var modelTeacher []model.Teacher
	page, size = pageNumHandle(page, size)
	baseQ := d.db.Limit(size).Offset((page - 1) * size)
	if workNo != "" {
		baseQ = baseQ.Where("work_no like ?", workNo+"%")
	}
	if name != "" {
		baseQ = baseQ.Where("name like ?", "%"+name+"%")
	}
	baseQ.Order("id desc").Find(&modelTeacher)
	return &modelTeacher, baseQ.Error
}
