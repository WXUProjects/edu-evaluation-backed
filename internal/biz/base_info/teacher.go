package base_info

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

type TeacherUseCase struct {
	baseDal *dal.BaseInfoDal
}

func NewTeacherUseCase(baseDal *dal.BaseInfoDal) *TeacherUseCase {
	return &TeacherUseCase{
		baseDal: baseDal,
	}
}

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
	submit(tmp)
	return nil
}
