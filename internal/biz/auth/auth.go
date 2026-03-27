package auth

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
)

// AuthUseCase 认证业务用例
// 处理管理员和学生登录验证
type AuthUseCase struct {
	baseDal *dal.BaseInfoDal
}

// NewAuthUseCase 创建认证业务用例实例
func NewAuthUseCase(baseDal *dal.BaseInfoDal) *AuthUseCase {
	return &AuthUseCase{
		baseDal: baseDal,
	}
}

// AdminLogin 管理员登录
func (a *AuthUseCase) AdminLogin(username, password string) (*model.Admin, error) {
	return a.baseDal.AdminLogin(username, password)
}

// StudentLogin 学生登录
// stuNo: 学号
// cardNo: 身份证号
// taskId: 评教任务ID
func (a *AuthUseCase) StudentLogin(stuNo, cardNo string, taskId uint) (*model.Student, error) {
	return a.baseDal.StudentLogin(stuNo, cardNo, taskId)
}

// GetStudentInfo 获取学生信息
func (a *AuthUseCase) GetStudentInfo(stuNo string) (*model.Student, error) {
	return a.baseDal.GetStudentByStudentNo(stuNo)
}