// Package base_info 提供基础信息（课程、学生、教师）的业务逻辑层实现。
package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

// TeacherUseCase 教师信息业务用例，负责教师数据的导入、查询、更新和删除等业务操作。
type TeacherUseCase struct {
	baseDal *dal.BaseInfoDal
}

// NewTeacherUseCase 创建并返回 TeacherUseCase 实例。
// 参数 baseDal 为基础信息数据访问层对象。
func NewTeacherUseCase(baseDal *dal.BaseInfoDal) *TeacherUseCase {
	return &TeacherUseCase{baseDal: baseDal}
}

// Import 从 Excel 文件批量导入教师数据。
// 解析 Excel 中的 Sheet1，按每批 200 条进行批量插入。
// 参数 f 为上传的 Excel 文件句柄。
// 返回文件解析或读取过程中的错误；导入成功返回 nil。
func (s TeacherUseCase) Import(f multipart.File) error {
	list, err := excelize.OpenReader(f)
	if err != nil {
		return err
	}
	defer list.Close()

	rows, err := list.GetRows("Sheet1")
	if err != nil {
		return err
	}

	var batch []*model.Teacher
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		batch = append(batch, &model.Teacher{
			WorkNo: row[1],
			Name:   row[2],
			Sex:    row[3],
			Email:  row[4],
		})
		if len(batch) >= 200 {
			s.baseDal.InsertTeacher(batch)
			batch = nil
		}
	}
	if len(batch) > 0 {
		s.baseDal.InsertTeacher(batch)
	}
	return nil
}

// UpdateTeacher 更新指定教师的信息。
// 参数 id 为目标教师的主键 ID；name、sex、workNo、email 为需要更新的字段指针，传 nil 表示不更新该字段。
// 返回更新后的教师对象和可能的错误。
func (s TeacherUseCase) UpdateTeacher(id uint, name, sex, workNo, email *string) (*model.Teacher, error) {
	return s.baseDal.UpdateTeacher(id, name, sex, workNo, email)
}

// DeleteTeacher 根据教师 ID 删除指定教师。
// 参数 id 为目标教师的主键 ID。
// 删除成功返回 nil，否则返回对应的错误。
func (s TeacherUseCase) DeleteTeacher(id uint) error {
	return s.baseDal.DeleteTeacher(id)
}

// GetTeacherByID 根据教师 ID 查询教师详情。
// 参数 id 为目标教师的主键 ID。
// 返回查询到的教师对象和可能的错误。
func (s TeacherUseCase) GetTeacherByID(id uint) (*model.Teacher, error) {
	return s.baseDal.GetTeacherByID(id)
}
