package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

// StudentUseCase 学生信息业务用例
// 处理学生信息相关的业务逻辑，包括Excel导入、更新、删除、详情查询
type StudentUseCase struct {
	studentDal *dal.BaseInfoDal
}

// NewStudentUseCase 创建学生信息业务用例实例
// studentDal: 基础信息数据访问层
// 返回值: 学生信息业务用例实例指针
func NewStudentUseCase(studentDal *dal.BaseInfoDal) *StudentUseCase {
	return &StudentUseCase{
		studentDal: studentDal,
	}
}

// ImportStudent 从Excel文件导入学生数据
// f: 上传的Excel文件句柄
// 从Sheet1读取数据，第一行为表头，从第二行开始为数据
// 每200行批量插入一次数据库，学号冲突时自动跳过
// 返回值: 导入过程中的错误，如果成功返回nil
func (s StudentUseCase) ImportStudent(f multipart.File) error {
	list, err := excelize.OpenReader(f)
	if err != nil {
		return err
	}
	defer func() {
		_ = list.Close()
	}()
	rows, err := list.GetRows("Sheet1")
	var tmp []*model.Student
	submit := func(data []*model.Student) {
		s.studentDal.InsertStudent(data)
	}
	if err != nil {
		return err
	}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		tmp = append(tmp, &model.Student{
			StudentNo: row[1],
			Name:      row[2],
			Sex:       row[3],
			IdCardNo:  row[4],
		})
		// 每隔200行提交一次数据
		if i%200 == 0 {
			submit(tmp)
			tmp = []*model.Student{}
		}
	}
	// 提交剩余不足200行的数据
	submit(tmp)
	return nil
}

// UpdateStudent 更新学生信息
// id: 学生ID
// name: 姓名指针，为nil表示不更新该字段
// sex: 性别指针，为nil表示不更新该字段
// studentNo: 学号指针，为nil表示不更新该字段
// idCardNo: 身份证号指针，为nil表示不更新该字段
// 返回值: 更新后的学生信息，错误信息
// 如果学号发生变更，会检查是否与其他学生冲突
func (s StudentUseCase) UpdateStudent(id uint, name, sex, studentNo, idCardNo *string) (*model.Student, error) {
	return s.studentDal.UpdateStudent(id, name, sex, studentNo, idCardNo)
}

// DeleteStudent 删除学生
// id: 要删除的学生ID
// 删除前会清除学生与所有课程的关联，学生关联清除后删除学生本身
// 返回值: 删除成功返回nil，错误信息
func (s StudentUseCase) DeleteStudent(id uint) error {
	return s.studentDal.DeleteStudent(id)
}

// GetStudentByID 根据ID获取学生详情
// id: 学生ID
// 返回值: 学生信息指针，错误信息
func (s StudentUseCase) GetStudentByID(id uint) (*model.Student, error) {
	return s.studentDal.GetStudentByID(id)
}
