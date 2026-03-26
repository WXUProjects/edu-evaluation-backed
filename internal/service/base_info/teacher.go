package base_info

import (
	"context"
	"edu-evaluation-backed/api/v1/base_info/teacher"
	"edu-evaluation-backed/internal/biz/base_info"
	"edu-evaluation-backed/internal/data/dal"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/transport/http"
)

type TeacherService struct {
	teacherUc *base_info.TeacherUseCase
	baseDal   *dal.BaseInfoDal
}

func (s TeacherService) List(ctx context.Context, req *teacher_i.GetTeacherListReq) (*teacher_i.GetTeacherListResp, error) {
	// 包含分页 模糊查询功能
	teacherList, tot, err := s.baseDal.QueryTeacher(int(req.Page), int(req.PageSize), req.WorkNo, req.Name)
	if err != nil {
		return nil, err
	}
	res := &teacher_i.GetTeacherListResp{
		Message: "success",
		Data:    make([]*teacher_i.TeacherInfo, 0),
		Total:   tot,
	}
	for _, v := range *teacherList {
		res.Data = append(res.Data, &teacher_i.TeacherInfo{
			Id:     uint32(v.ID),
			Name:   v.Name,
			Sex:    v.Sex,
			WorkNo: v.WorkNo,
			Email:  v.Email,
		})
	}
	return res, nil
}

func (s TeacherService) Import(ctx http.Context) error {
	req := ctx.Request()
	file, _, err := req.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()
	err = s.teacherUc.Import(file)
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

func (s TeacherService) Update(ctx context.Context, req *teacher_i.UpdateTeacherReq) (*teacher_i.UpdateTeacherResp, error) {
	// 调用biz层更新
	teacher, err := s.teacherUc.UpdateTeacher(uint(req.Id), req.Name, req.Sex, req.WorkNo, req.Email)
	if err != nil {
		return nil, err
	}

	// 返回更新后的数据
	return &teacher_i.UpdateTeacherResp{
		Message: "修改成功",
		Data: &teacher_i.TeacherInfo{
			Id:     uint32(teacher.ID),
			Name:   teacher.Name,
			Sex:    teacher.Sex,
			WorkNo: teacher.WorkNo,
			Email:  teacher.Email,
		},
	}, nil
}

func (s TeacherService) Delete(ctx context.Context, req *teacher_i.DeleteTeacherReq) (*teacher_i.DeleteTeacherResp, error) {
	// 调用biz层删除
	err := s.teacherUc.DeleteTeacher(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return &teacher_i.DeleteTeacherResp{
		Message: "删除成功",
	}, nil
}

func NewTeacherService(teacherUc *base_info.TeacherUseCase, baseDal *dal.BaseInfoDal) *TeacherService {
	return &TeacherService{teacherUc: teacherUc, baseDal: baseDal}
}
