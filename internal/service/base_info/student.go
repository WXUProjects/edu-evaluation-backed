package base_info

import (
	"context"
	baseinfo2 "edu-evaluation-backed/api/v1/base_info/student"
	"edu-evaluation-backed/internal/biz/base_info"
	"edu-evaluation-backed/internal/data/dal"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// StudentService 学生信息服务
// 提供学生信息的查询、导入、更新、删除等功能
type StudentService struct {
	studentUc *base_info.StudentUseCase
	baseDal   *dal.BaseInfoDal
}

// List 获取学生列表
// 支持分页查询和按学号、姓名模糊搜索
// ctx: 上下文
// req: 获取学生列表请求，包含页码、每页条数、搜索关键词
// 返回值: 学生列表响应，包含数据列表和总数，错误信息
func (s StudentService) List(ctx context.Context, req *baseinfo2.GetStudentListReq) (*baseinfo2.GetStudentListResp, error) {
	// 包含分页 模糊查询功能
	studentList, tot, err := s.baseDal.QueryStudent(int(req.Page), int(req.PageSize), req.StudentNo, req.Name)
	if err != nil {
		return nil, err
	}
	res := &baseinfo2.GetStudentListResp{
		Message: "success",
		Data:    make([]*baseinfo2.StudentInfo, 0),
		Total:   tot,
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

// Import 导入学生信息Excel文件
// 从Excel的Sheet1中读取学生数据，批量导入数据库
// Excel格式要求：第2列开始依次是学号、姓名、性别、身份证号，第一行为表头
// ctx: HTTP上下文，包含上传的文件
// 返回值: 错误信息，成功返回nil
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

// Update 更新学生信息
// 根据学生ID更新学生的姓名、性别、学号、身份证号
// 只更新请求中提供的非空字段
// ctx: 上下文
// req: 更新学生请求，包含学生ID和要更新的字段
// 返回值: 更新成功后返回更新后的学生信息，错误信息
func (s StudentService) Update(ctx context.Context, req *baseinfo2.UpdateStudentReq) (*baseinfo2.UpdateStudentResp, error) {
	// 调用biz层更新
	student, err := s.studentUc.UpdateStudent(uint(req.Id), req.Name, req.Sex, req.StudentNo, req.IdCardNo)
	if err != nil {
		return nil, err
	}

	// 返回更新后的数据
	return &baseinfo2.UpdateStudentResp{
		Message: "修改成功",
		Data: &baseinfo2.StudentInfo{
			Id:        uint32(student.ID),
			Name:      student.Name,
			Sex:       student.Sex,
			StudentNo: student.StudentNo,
			IdCardNo:  student.IdCardNo,
		},
	}, nil
}

// Delete 删除学生
// 根据学生ID删除学生，删除前会清除学生与课程的关联
// ctx: 上下文
// req: 删除学生请求，包含学生ID
// 返回值: 删除成功响应，错误信息
func (s StudentService) Delete(ctx context.Context, req *baseinfo2.DeleteStudentReq) (*baseinfo2.DeleteStudentResp, error) {
	// 调用biz层删除
	err := s.studentUc.DeleteStudent(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return &baseinfo2.DeleteStudentResp{
		Message: "删除成功",
	}, nil
}

// NewStudentService 创建学生信息服务实例
// studentUc: 学生信息业务用例
// baseDal: 基础信息数据访问层
// 返回值: 学生信息服务实例指针
func NewStudentService(studentUc *base_info.StudentUseCase, baseDal *dal.BaseInfoDal) *StudentService {
	return &StudentService{studentUc: studentUc, baseDal: baseDal}
}
