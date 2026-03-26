package base_info

import (
	"context"
	baseinfo2 "edu-evaluation-backed/api/v1/base_info/student"
	"edu-evaluation-backed/internal/biz/base_info"
	"edu-evaluation-backed/internal/data/dal"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/transport/http"
)

type StudentService struct {
	studentUc *base_info.StudentUseCase
	baseDal   *dal.BaseInfoDal
}

func (s StudentService) List(ctx context.Context, req *baseinfo2.GetStudentListReq) (*baseinfo2.GetStudentListResp, error) {
	// 包含分页 模糊查询功能
	studentList, err := s.baseDal.QueryStudent(int(req.Page), int(req.PageSize), req.StudentNo, req.Name)
	if err != nil {
		return nil, err
	}
	res := &baseinfo2.GetStudentListResp{
		Message: "success",
		Data:    make([]*baseinfo2.StudentInfo, 0),
	}
	for _, v := range *studentList {
		res.Data = append(res.Data, &baseinfo2.StudentInfo{
			Id:        uint32(v.ID),
			Name:      v.Name,
			Sex:       v.Sex,
			StudentNo: v.StudentNo,
			IdCardNo:  v.IdCardNo,
		})
	}
	return res, nil
}

func (s StudentService) Import(ctx http.Context) error {
	req := ctx.Request()
	file, _, err := req.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()
	err = s.studentUc.ImportStudent(file)
	if err != nil {
		return err
	}
	w := ctx.Response()
	w.WriteHeader(200)
	resp, _ := json.Marshal(map[string]string{
		"message": "导入成功",
	})
	_, _ = w.Write(resp)
	return nil
}

func NewStudentService(studentUc *base_info.StudentUseCase, baseDal *dal.BaseInfoDal) *StudentService {
	return &StudentService{studentUc: studentUc, baseDal: baseDal}
}
