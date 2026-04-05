package dal

import (
	"context"
	"errors"
	"time"

	"edu-evaluation-backed/internal/common/data/cache"
	"edu-evaluation-backed/internal/common/utils"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BaseInfoDal 基础信息数据访问层
type BaseInfoDal struct {
	db  *gorm.DB
	rdb *redis.Client
	hc  *cache.HealthChecker
}

// NewBaseInfoDal 创建基础信息数据访问层实例
func NewBaseInfoDal(data *data.Data) *BaseInfoDal {
	return &BaseInfoDal{
		db:  data.DB,
		rdb: data.RDB,
		hc:  data.HC,
	}
}

// InsertStudent 批量插入学生数据（UPSERT策略）
func (d *BaseInfoDal) InsertStudent(students []*model.Student) error {
	return d.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: "student_no"}},
		}).Create(students).Error
}

// QueryStudent 查询学生列表，支持分页和模糊搜索
func (d *BaseInfoDal) QueryStudent(page, size int, studentNo, name string) (*[]model.Student, int64, error) {
	var students []model.Student
	page, size = utils.PageNumHandle(page, size)
	var tot int64
	query := d.db.Model(&model.Student{})
	if studentNo != "" {
		query = query.Where("student_no LIKE ?", studentNo+"%")
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	err := query.Count(&tot).Order("id desc").Limit(size).Offset(utils.CalculateOffset(page, size)).Find(&students).Error
	return &students, tot, err
}

// InsertTeacher 批量插入教师数据（UPSERT策略）
func (d *BaseInfoDal) InsertTeacher(teachers []*model.Teacher) error {
	return d.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: "work_no"}},
		}).Create(teachers).Error
}

// QueryTeacher 查询教师列表，支持分页和模糊搜索
func (d *BaseInfoDal) QueryTeacher(page, size int, workNo, name string) (*[]model.Teacher, int64, error) {
	var teachers []model.Teacher
	page, size = utils.PageNumHandle(page, size)
	var tot int64
	query := d.db.Model(&model.Teacher{})
	if workNo != "" {
		query = query.Where("work_no LIKE ?", workNo+"%")
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	err := query.Count(&tot).Limit(size).Offset(utils.CalculateOffset(page, size)).Order("id desc").Find(&teachers).Error
	return &teachers, tot, err
}

// GetStudentByID 根据ID获取学生
func (d *BaseInfoDal) GetStudentByID(id uint) (*model.Student, error) {
	var student model.Student
	return &student, d.db.First(&student, id).Error
}

// UpdateStudent 更新学生信息
func (d *BaseInfoDal) UpdateStudent(id uint, name, sex, studentNo, idCardNo *string) (*model.Student, error) {
	var student model.Student
	if err := d.db.First(&student, id).Error; err != nil {
		return nil, err
	}

	if studentNo != nil && *studentNo != student.StudentNo {
		var count int64
		if err := d.db.Model(&model.Student{}).Where("student_no = ? AND id != ?", *studentNo, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("学号已存在")
		}
	}

	updates := buildUpdateMap(map[string]*string{
		"name":       name,
		"sex":        sex,
		"student_no": studentNo,
		"id_card_no": idCardNo,
	})
	if len(updates) > 0 {
		if err := d.db.Model(&student).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	d.db.First(&student, id)
	return &student, nil
}

// DeleteStudent 删除学生（清除课程关联后删除）
func (d *BaseInfoDal) DeleteStudent(id uint) error {
	var student model.Student
	if err := d.db.First(&student, id).Error; err != nil {
		return err
	}
	_ = d.db.Model(&student).Association("Courses").Clear()
	return d.db.Delete(&student).Error
}

// GetTeacherByID 根据ID获取教师
func (d *BaseInfoDal) GetTeacherByID(id uint) (*model.Teacher, error) {
	var teacher model.Teacher
	return &teacher, d.db.First(&teacher, id).Error
}

// UpdateTeacher 更新教师信息
func (d *BaseInfoDal) UpdateTeacher(id uint, name, sex, workNo, email *string) (*model.Teacher, error) {
	var teacher model.Teacher
	if err := d.db.First(&teacher, id).Error; err != nil {
		return nil, err
	}

	if workNo != nil && *workNo != teacher.WorkNo {
		var count int64
		if err := d.db.Model(&model.Teacher{}).Where("work_no = ? AND id != ?", *workNo, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("工号已存在")
		}
	}

	updates := buildUpdateMap(map[string]*string{
		"name":    name,
		"sex":     sex,
		"work_no": workNo,
		"email":   email,
	})
	if len(updates) > 0 {
		if err := d.db.Model(&teacher).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	d.db.First(&teacher, id)
	return &teacher, nil
}

// DeleteTeacher 删除教师（清除课程关联后删除）
func (d *BaseInfoDal) DeleteTeacher(id uint) error {
	var teacher model.Teacher
	if err := d.db.First(&teacher, id).Error; err != nil {
		return err
	}
	if err := d.db.Model(&teacher).Association("Courses").Clear(); err != nil {
		return err
	}
	return d.db.Delete(&teacher).Error
}

// buildUpdateMap 构建更新map，只包含非空字段
func buildUpdateMap(fields map[string]*string) map[string]interface{} {
	updates := make(map[string]interface{})
	for k, v := range fields {
		if v != nil && *v != "" {
			updates[k] = *v
		}
	}
	return updates
}

// AdminLogin 管理员登录验证
func (d *BaseInfoDal) AdminLogin(username, password string) (*model.Admin, error) {
	var admin model.Admin
	err := d.db.Where("username = ? AND password = ?", username, password).First(&admin).Error
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return &admin, nil
}

// StudentLogin 学生登录验证（带限流）
func (d *BaseInfoDal) StudentLogin(stuNo, cardNo string, taskId uint) (*model.Student, error) {
	// 登录限流：15分钟内最多5次
	ctx := context.Background()
	rateKey := cache.LoginRateKey(stuNo)
	count, allowed, err := cache.CheckAndSetRateLimit(ctx, d.rdb, d.hc, rateKey, 5, 15*time.Minute)
	if err != nil {
		// 限流检查失败不阻塞登录
		_ = err
	}
	if !allowed {
		return nil, errors.New("登录尝试过于频繁，请15分钟后再试")
	}

	// 1. 验证学生身份
	var student model.Student
	err = d.db.Where("student_no = ? AND id_card_no = ?", stuNo, cardNo).First(&student).Error
	if err != nil {
		return nil, errors.New("学号或身份证号错误")
	}

	// 2. 验证学生是否在指定 Task 范围内
	var count2 int64
	err = d.db.Table("courses c").
		Joins("INNER JOIN evaluation_courses ec ON c.id = ec.course_id").
		Joins("INNER JOIN course_students cs ON c.id = cs.course_id").
		Where("ec.evaluation_task_id = ? AND cs.student_student_no = ?", taskId, student.StudentNo).
		Count(&count2).Error

	if err != nil {
		return nil, err
	}

	if count2 == 0 {
		return nil, errors.New("您不在本次评教范围内或该任务暂无课程")
	}

	_ = count
	return &student, nil
}

// GetStudentByStudentNo 根据学号获取学生信息
func (d *BaseInfoDal) GetStudentByStudentNo(stuNo string) (*model.Student, error) {
	var student model.Student
	err := d.db.Where("student_no = ?", stuNo).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// AdminChangePassword 管理员修改密码
func (d *BaseInfoDal) AdminChangePassword(username, oldPassword, newPassword string) error {
	var admin model.Admin
	err := d.db.Where("username = ? AND password = ?", username, oldPassword).First(&admin).Error
	if err != nil {
		return errors.New("用户名或旧密码错误")
	}
	return d.db.Model(&admin).Update("password", newPassword).Error
}
