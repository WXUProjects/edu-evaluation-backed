package dal

import (
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"
	"errors"

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

func (d *BaseInfoDal) QueryStudent(page, size int, studentNo, name string) (*[]model.Student, int64, error) {
	var modelStudent []model.Student
	page, size = pageNumHandle(page, size)
	var tot int64
	baseQ := d.db.Model(model.Student{})
	if studentNo != "" {
		baseQ = baseQ.Where("student_no like ?", studentNo+"%")
	}
	if name != "" {
		baseQ = baseQ.Where("name like ?", "%"+name+"%")
	}
	baseQ.Count(&tot).Order("id desc").Limit(size).Offset((page - 1) * size).Find(&modelStudent)
	return &modelStudent, tot, baseQ.Error
}

func (d *BaseInfoDal) InsertTeacher(teachers []*model.Teacher) error {
	return d.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: "work_no"}},
		}).Create(teachers).Error
}

func (d *BaseInfoDal) QueryTeacher(page, size int, workNo, name string) (*[]model.Teacher, int64, error) {
	var modelTeacher []model.Teacher
	page, size = pageNumHandle(page, size)
	baseQ := d.db.Model(model.Teacher{})
	if workNo != "" {
		baseQ = baseQ.Where("work_no like ?", workNo+"%")
	}
	if name != "" {
		baseQ = baseQ.Where("name like ?", "%"+name+"%")
	}
	var tot int64
	baseQ.Count(&tot).Limit(size).Offset((page - 1) * size).Order("id desc").Find(&modelTeacher)
	return &modelTeacher, tot, baseQ.Error
}

// GetStudentByID 根据ID获取学生
func (d *BaseInfoDal) GetStudentByID(id uint) (*model.Student, error) {
	var student model.Student
	err := d.db.First(&student, id).Error
	return &student, err
}

// UpdateStudent 更新学生信息
func (d *BaseInfoDal) UpdateStudent(id uint, name, sex, studentNo, idCardNo *string) (*model.Student, error) {
	// 先查询学生是否存在
	var student model.Student
	if err := d.db.First(&student, id).Error; err != nil {
		return nil, err
	}

	// 如果修改了学号，检查是否与其他学生冲突
	if studentNo != nil && *studentNo != student.StudentNo {
		var count int64
		err := d.db.Model(&model.Student{}).Where("student_no = ? AND id != ?", *studentNo, id).Count(&count).Error
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("学号已存在")
		}
	}

	// 构建更新map，只更新非空字段
	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if sex != nil {
		updates["sex"] = *sex
	}
	if studentNo != nil {
		updates["student_no"] = *studentNo
	}
	if idCardNo != nil {
		updates["id_card_no"] = *idCardNo
	}

	if len(updates) == 0 {
		return &student, nil
	}

	// 执行更新
	if err := d.db.Model(&student).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 查询更新后的数据
	if err := d.db.First(&student, id).Error; err != nil {
		return nil, err
	}

	return &student, nil
}

// DeleteStudent 删除学生（清除课程关联后删除）
func (d *BaseInfoDal) DeleteStudent(id uint) error {
	var student model.Student
	if err := d.db.First(&student, id).Error; err != nil {
		return err
	}

	// 清除与课程的关联，课程保留
	if err := d.db.Model(&student).Association("Courses").Clear(); err != nil {
		//return err
	}

	// 删除学生本身
	if err := d.db.Delete(&student).Error; err != nil {
		return err
	}

	return nil
}

// GetTeacherByID 根据ID获取教师
func (d *BaseInfoDal) GetTeacherByID(id uint) (*model.Teacher, error) {
	var teacher model.Teacher
	err := d.db.First(&teacher, id).Error
	return &teacher, err
}

// UpdateTeacher 更新教师信息
func (d *BaseInfoDal) UpdateTeacher(id uint, name, sex, workNo, email *string) (*model.Teacher, error) {
	// 先查询教师是否存在
	var teacher model.Teacher
	if err := d.db.First(&teacher, id).Error; err != nil {
		return nil, err
	}

	// 如果修改了工号，检查是否与其他教师冲突
	if workNo != nil && *workNo != teacher.WorkNo {
		var count int64
		err := d.db.Model(&model.Teacher{}).Where("work_no = ? AND id != ?", *workNo, id).Count(&count).Error
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("工号已存在")
		}
	}

	// 构建更新map，只更新非空字段
	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if sex != nil {
		updates["sex"] = *sex
	}
	if workNo != nil {
		updates["work_no"] = *workNo
	}
	if email != nil {
		updates["email"] = *email
	}

	if len(updates) == 0 {
		return &teacher, nil
	}

	// 执行更新
	if err := d.db.Model(&teacher).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 查询更新后的数据
	if err := d.db.First(&teacher, id).Error; err != nil {
		return nil, err
	}

	return &teacher, nil
}

// DeleteTeacher 删除教师（清除课程关联后删除）
func (d *BaseInfoDal) DeleteTeacher(id uint) error {
	var teacher model.Teacher
	if err := d.db.First(&teacher, id).Error; err != nil {
		return err
	}

	// 清除与课程的关联，课程保留
	if err := d.db.Model(&teacher).Association("Courses").Clear(); err != nil {
		return err
	}

	// 删除教师本身
	if err := d.db.Delete(&teacher).Error; err != nil {
		return err
	}

	return nil
}
