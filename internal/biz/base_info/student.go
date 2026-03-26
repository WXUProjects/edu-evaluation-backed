package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

type StudentUseCase struct {
	studentDal *dal.BaseInfoDal
}

func NewStudentUseCase(studentDal *dal.BaseInfoDal) *StudentUseCase {
	return &StudentUseCase{
		studentDal: studentDal,
	}
}

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
	submit(tmp)
	return nil
}

// UpdateStudent 更新学生信息
func (s StudentUseCase) UpdateStudent(id uint, name, sex, studentNo, idCardNo *string) (*model.Student, error) {
	return s.studentDal.UpdateStudent(id, name, sex, studentNo, idCardNo)
}

// DeleteStudent 删除学生
func (s StudentUseCase) DeleteStudent(id uint) error {
	return s.studentDal.DeleteStudent(id)
}

// GetStudentByID 获取学生详情
func (s StudentUseCase) GetStudentByID(id uint) (*model.Student, error) {
	return s.studentDal.GetStudentByID(id)
}
