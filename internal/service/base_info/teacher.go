package base_info

import (
	"context"
	teacher_i "edu-evaluation-backed/api/v1/base_info/teacher"
	"edu-evaluation-backed/internal/biz/base_info"
	"edu-evaluation-backed/internal/data/dal"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// TeacherService 教师信息服务
// 提供教师信息的查询、导入、更新、删除等功能
type TeacherService struct {
	teacherUc *base_info.TeacherUseCase
	baseDal   *dal.BaseInfoDal
}

// List 获取教师列表
// 支持分页查询和按工号、姓名模糊搜索
// ctx: 上下文
// req: 获取教师列表请求，包含页码、每页条数、搜索关键词
// 返回值: 教师列表响应，包含数据列表和总数，错误信息
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

// Import 导入教师信息Excel文件
// 从Excel的Sheet1中读取教师数据，批量导入数据库
// Excel格式要求：第2列开始依次是工号、姓名、性别、邮箱，第一行为表头
// ctx: HTTP上下文，包含上传的文件
// 返回值: 错误信息，成功返回nil
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

// Update 更新教师信息
// 根据教师ID更新教师的姓名、性别、工号、邮箱
// 只更新请求中提供的非空字段
// ctx: 上下文
// req: 更新教师请求，包含教师ID和要更新的字段
// 返回值: 更新成功后返回更新后的教师信息，错误信息
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

// Delete 删除教师
// 根据教师ID删除教师，删除前会清除教师与课程的关联
// ctx: 上下文
// req: 删除教师请求，包含教师ID
// 返回值: 删除成功响应，错误信息
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

// NewTeacherService 创建教师信息服务实例
// teacherUc: 教师信息业务用例
// baseDal: 基础信息数据访问层
// 返回值: 教师信息服务实例指针
func NewTeacherService(teacherUc *base_info.TeacherUseCase, baseDal *dal.BaseInfoDal) *TeacherService {
	return &TeacherService{teacherUc: teacherUc, baseDal: baseDal}
}
