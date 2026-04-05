// Package base_info 提供基础信息（课程、学生、教师）的业务逻辑层实现。
package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

// StudentUseCase 学生信息业务用例，负责学生数据的导入、查询、更新和删除等业务操作。
type StudentUseCase struct {
	studentDal *dal.BaseInfoDal
}

// NewStudentUseCase 创建并返回 StudentUseCase 实例。
// 参数 studentDal 为基础信息数据访问层对象。
func NewStudentUseCase(studentDal *dal.BaseInfoDal) *StudentUseCase {
	return &StudentUseCase{studentDal: studentDal}
}

// ImportStudent 从 Excel 文件批量导入学生数据。
// 解析 Excel 中的 Sheet1，按每批 200 条进行批量插入。
// 参数 f 为上传的 Excel 文件句柄。
// 返回文件解析或读取过程中的错误；导入成功返回 nil。
func (s StudentUseCase) ImportStudent(f multipart.File) error {
	list, err := excelize.OpenReader(f)
	if err != nil {
		return err
	}
	defer list.Close()

	rows, err := list.GetRows("Sheet1")
	if err != nil {
		return err
	}

	var batch []*model.Student
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		batch = append(batch, &model.Student{
			StudentNo: row[1],
			Name:      row[2],
			Sex:       row[3],
			IdCardNo:  row[4],
		})
		if len(batch) >= 200 {
			s.studentDal.InsertStudent(batch)
			batch = nil
		}
	}
	if len(batch) > 0 {
		s.studentDal.InsertStudent(batch)
	}
	return nil
}

// UpdateStudent 更新指定学生的信息。
// 参数 id 为目标学生的主键 ID；name、sex、studentNo、idCardNo 为需要更新的字段指针，传 nil 表示不更新该字段。
// 返回更新后的学生对象和可能的错误。
func (s StudentUseCase) UpdateStudent(id uint, name, sex, studentNo, idCardNo *string) (*model.Student, error) {
	return s.studentDal.UpdateStudent(id, name, sex, studentNo, idCardNo)
}

// DeleteStudent 根据学生 ID 删除指定学生。
// 参数 id 为目标学生的主键 ID。
// 删除成功返回 nil，否则返回对应的错误。
func (s StudentUseCase) DeleteStudent(id uint) error {
	return s.studentDal.DeleteStudent(id)
}

// GetStudentByID 根据学生 ID 查询学生详情。
// 参数 id 为目标学生的主键 ID。
// 返回查询到的学生对象和可能的错误。
func (s StudentUseCase) GetStudentByID(id uint) (*model.Student, error) {
	return s.studentDal.GetStudentByID(id)
}
