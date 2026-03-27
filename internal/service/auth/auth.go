package auth

import (
	"edu-evaluation-backed/internal/biz/auth"
	"errors"
	"strconv"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// AuthService 认证服务
type AuthService struct {
	authUc *auth.AuthUseCase
}

// NewAuthService 创建认证服务实例
func NewAuthService(authUc *auth.AuthUseCase) *AuthService {
	return &AuthService{
		authUc: authUc,
	}
}

// AdminLogin 管理员登录
// ctx: HTTP上下文
// req: 管理员登录请求，包含用户名和密码
// 返回值: 错误信息，成功返回nil
func (s *AuthService) AdminLogin(ctx http.Context) error {
	req := ctx.Request()
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")

	if username == "" || password == "" {
		return errors.New("用户名和密码不能为空")
	}

	_, err := s.authUc.AdminLogin(username, password)
	if err != nil {
		return err
	}

	w := ctx.Response()
	w.WriteHeader(200)
	resp := []byte(`{"message":"登录成功"}`)
	_, _ = w.Write(resp)
	return nil
}

// StudentLogin 学生登录
// ctx: HTTP上下文
// req: 学生登录请求，包含学号、身份证号、评教任务ID
// 返回值: 错误信息，成功返回nil
func (s *AuthService) StudentLogin(ctx http.Context) error {
	req := ctx.Request()
	stuNo := req.PostFormValue("stu_no")
	cardNo := req.PostFormValue("card_no")
	taskIdStr := req.PostFormValue("task_id")

	if stuNo == "" || cardNo == "" || taskIdStr == "" {
		return errors.New("学号、身份证号和评教任务ID不能为空")
	}

	var taskId uint
	if _, err := parseUint(taskIdStr, &taskId); err != nil {
		return errors.New("评教任务ID格式错误")
	}

	student, err := s.authUc.StudentLogin(stuNo, cardNo, taskId)
	if err != nil {
		return err
	}

	w := ctx.Response()
	w.WriteHeader(200)
	resp := []byte(`{"message":"登录成功","data":{"student_no":"` + student.StudentNo + `","name":"` + student.Name + `"}}`)
	_, _ = w.Write(resp)
	return nil
}

// StudentInfo 获取学生个人信息
// ctx: HTTP上下文
// 返回值: 错误信息，成功返回nil
func (s *AuthService) StudentInfo(ctx http.Context) error {
	req := ctx.Request()
	stuNo := req.URL.Query().Get("stuNo")
	if stuNo == "" {
		return errors.New("学号不能为空")
	}

	student, err := s.authUc.GetStudentInfo(stuNo)
	if err != nil {
		return err
	}

	w := ctx.Response()
	w.WriteHeader(200)
	resp := []byte(`{"message":"success","data":{"id":` + uint32ToStr(student.ID) + `,"name":"` + student.Name + `","sex":"` + student.Sex + `","student_no":"` + student.StudentNo + `","id_card_no":"` + student.IdCardNo + `"}}`)
	_, _ = w.Write(resp)
	return nil
}

func uint32ToStr(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

// StudentInfoResp 学生信息响应
type StudentInfoResp struct {
	Id        uint32
	Name      string
	Sex       string
	StudentNo string
	IdCardNo  string
}

// parseUint 简单解析uint
func parseUint(s string, result *uint) (bool, error) {
	var v uint
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, errors.New("invalid number")
		}
		v = v*10 + uint(c-'0')
	}
	*result = v
	return true, nil
}
