package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

// TeacherUseCase 教师信息业务用例
// 处理教师信息相关的业务逻辑，包括Excel导入、更新、删除、详情查询
type TeacherUseCase struct {
	baseDal *dal.BaseInfoDal
}

// NewTeacherUseCase 创建教师信息业务用例实例
// baseDal: 基础信息数据访问层
// 返回值: 教师信息业务用例实例指针
func NewTeacherUseCase(baseDal *dal.BaseInfoDal) *TeacherUseCase {
	return &TeacherUseCase{
		baseDal: baseDal,
	}
}

// Import 从Excel文件导入教师数据
// f: 上传的Excel文件句柄
// 从Sheet1读取数据，第一行为表头，从第二行开始为数据
// 每200行批量插入一次数据库，工号冲突时自动跳过
// 返回值: 导入过程中的错误，如果成功返回nil
func (s TeacherUseCase) Import(f multipart.File) error {
	list, err := excelize.OpenReader(f)
	if err != nil {
		return err
	}
	defer func() {
		_ = list.Close()
	}()
	rows, err := list.GetRows("Sheet1")
	var tmp []*model.Teacher
	submit := func(data []*model.Teacher) {
		s.baseDal.InsertTeacher(data)
	}
	if err != nil {
		return err
	}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		tmp = append(tmp, &model.Teacher{
			WorkNo: row[1],
			Name:   row[2],
			Sex:    row[3],
			Email:  row[4],
		})
		// 每隔200行提交一次数据
		if i%200 == 0 {
			submit(tmp)
			tmp = []*model.Teacher{}
		}
	}
	// 提交剩余不足200行的数据
	submit(tmp)
	return nil
}

// UpdateTeacher 更新教师信息
// id: 教师ID
// name: 姓名指针，为nil表示不更新该字段
// sex: 性别指针，为nil表示不更新该字段
// workNo: 工号指针，为nil表示不更新该字段
// email: 邮箱指针，为nil表示不更新该字段
// 返回值: 更新后的教师信息，错误信息
// 如果工号发生变更，会检查是否与其他教师冲突
func (s TeacherUseCase) UpdateTeacher(id uint, name, sex, workNo, email *string) (*model.Teacher, error) {
	return s.baseDal.UpdateTeacher(id, name, sex, workNo, email)
}

// DeleteTeacher 删除教师
// id: 要删除的教师ID
// 删除前会清除教师与所有课程的关联，教师关联清除后删除教师本身
// 返回值: 删除成功返回nil，错误信息
func (s TeacherUseCase) DeleteTeacher(id uint) error {
	return s.baseDal.DeleteTeacher(id)
}

// GetTeacherByID 根据ID获取教师详情
// id: 教师ID
// 返回值: 教师信息指针，错误信息
func (s TeacherUseCase) GetTeacherByID(id uint) (*model.Teacher, error) {
	return s.baseDal.GetTeacherByID(id)
}
