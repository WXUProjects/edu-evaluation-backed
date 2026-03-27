package dal

import (
	"edu-evaluation-backed/internal/common/utils"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BaseInfoDal 基础信息数据访问层
// 处理学生和教师信息的数据库操作，包括增删改查
type BaseInfoDal struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewBaseInfoDal 创建基础信息数据访问层实例
// data: 数据层上下文，包含数据库连接和Redis客户端
// 返回值: 基础信息数据访问层实例指针
func NewBaseInfoDal(data *data.Data) *BaseInfoDal {
	return &BaseInfoDal{
		db:  data.DB,
		rdb: data.RDB,
	}
}

// InsertStudent 批量插入学生数据
// students: 学生数据列表
// 使用UPSERT策略，当学号冲突时自动跳过（DoNothing）
// 返回值: 插入失败返回错误，成功返回nil
func (d *BaseInfoDal) InsertStudent(students []*model.Student) error {
	return d.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: "student_no"}},
		}).Create(students).Error
}

// QueryStudent 查询学生列表，支持分页和按学号、姓名模糊搜索
// page: 当前页码，size: 每页条数
// studentNo: 学号搜索关键词（前缀匹配），为空不搜索
// name: 姓名搜索关键词（模糊匹配），为空不搜索
// 返回值: 学生列表指针，总记录数，错误信息
// 结果按ID降序排列，保证最新添加的学生排在前面
func (d *BaseInfoDal) QueryStudent(page, size int, studentNo, name string) (*[]model.Student, int64, error) {
	var modelStudent []model.Student
	page, size = utils.PageNumHandle(page, size)
	var tot int64
	baseQ := d.db.Model(model.Student{})
	if studentNo != "" {
		baseQ = baseQ.Where("student_no like ?", studentNo+"%")
	}
	if name != "" {
		baseQ = baseQ.Where("name like ?", "%"+name+"%")
	}
	baseQ.Count(&tot).Order("id desc").Limit(size).Offset(utils.CalculateOffset(page, size)).Find(&modelStudent)
	return &modelStudent, tot, baseQ.Error
}

// InsertTeacher 批量插入教师数据
// teachers: 教师数据列表
// 使用UPSERT策略，当工号冲突时自动跳过（DoNothing）
// 返回值: 插入失败返回错误，成功返回nil
func (d *BaseInfoDal) InsertTeacher(teachers []*model.Teacher) error {
	return d.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: "work_no"}},
		}).Create(teachers).Error
}

// QueryTeacher 查询教师列表，支持分页和按工号、姓名模糊搜索
// page: 当前页码，size: 每页条数
// workNo: 工号搜索关键词（前缀匹配），为空不搜索
// name: 姓名搜索关键词（模糊匹配），为空不搜索
// 返回值: 教师列表指针，总记录数，错误信息
// 结果按ID降序排列，保证最新添加的教师排在前面
func (d *BaseInfoDal) QueryTeacher(page, size int, workNo, name string) (*[]model.Teacher, int64, error) {
	var modelTeacher []model.Teacher
	page, size = utils.PageNumHandle(page, size)
	baseQ := d.db.Model(model.Teacher{})
	if workNo != "" {
		baseQ = baseQ.Where("work_no like ?", workNo+"%")
	}
	if name != "" {
		baseQ = baseQ.Where("name like ?", "%"+name+"%")
	}
	var tot int64
	baseQ.Count(&tot).Limit(size).Offset(utils.CalculateOffset(page, size)).Order("id desc").Find(&modelTeacher)
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
// id: 学生ID
// 删除前清除学生与所有课程的多对多关联，课程保留，只删除关联关系
// 关联清除完成后删除学生本身
// 返回值: 删除成功返回nil，错误信息
func (d *BaseInfoDal) DeleteStudent(id uint) error {
	var student model.Student
	if err := d.db.First(&student, id).Error; err != nil {
		return err
	}

	// 清除与课程的关联，课程保留
	_ = d.db.Model(&student).Association("Courses").Clear()

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

// AdminLogin 管理员登录验证
// username: 管理员用户名
// password: 密码
// 返回值: 管理员信息，错误信息
func (d *BaseInfoDal) AdminLogin(username, password string) (*model.Admin, error) {
	var admin model.Admin
	err := d.db.Where("username = ? AND password = ?", username, password).First(&admin).Error
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return &admin, nil
}

// StudentLogin 学生登录验证
// stuNo: 学号
// cardNo: 身份证号
// taskId: 评教任务ID
// 返回值: 学生信息，错误信息
// 只有学生属于该task中任意一门课程时才能登录成功
func (d *BaseInfoDal) StudentLogin(stuNo, cardNo string, taskId uint) (*model.Student, error) {
	// 1. 先验证学生身份（学号和身份证）
	var student model.Student
	err := d.db.Where("student_no = ? AND id_card_no = ?", stuNo, cardNo).First(&student).Error
	if err != nil {
		return nil, errors.New("学号或身份证号错误")
	}

	// 2. 核心：一条 SQL 验证该学生是否在指定 Task 的范围内
	// 逻辑：寻找一门课，它既在 Task 关联中，又在学生的选课名单中
	var count int64
	err = d.db.Table("courses c").
		// 关联任务中间表
		Joins("INNER JOIN evaluation_courses ec ON c.id = ec.course_id").
		// 关联学生中间表（注意字段名是 student_student_no）
		Joins("INNER JOIN course_students cs ON c.id = cs.course_id").
		Where("ec.evaluation_task_id = ? AND cs.student_student_no = ?", taskId, student.StudentNo).
		Count(&count).Error

	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, errors.New("您不在本次评教范围内或该任务暂无课程")
	}

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
