package auth

import (
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/data/model"
)

// AuthUseCase 认证业务用例，负责管理员和学生的登录验证与密码管理
type AuthUseCase struct {
	baseDal *dal.BaseInfoDal
}

// NewAuthUseCase 创建认证业务用例实例
//
// 参数:
//   - baseDal *dal.BaseInfoDal 基础信息数据访问层
//
// 返回值:
//   - *AuthUseCase
func NewAuthUseCase(baseDal *dal.BaseInfoDal) *AuthUseCase {
	return &AuthUseCase{
		baseDal: baseDal,
	}
}

// AdminLogin 管理员登录，验证用户名和密码并返回管理员信息
//
// 参数:
//   - username string 管理员用户名
//   - password string 管理员密码
//
// 返回值:
//   - *model.Admin 管理员信息
//   - error 登录失败时返回错误
func (a *AuthUseCase) AdminLogin(username, password string) (*model.Admin, error) {
	return a.baseDal.AdminLogin(username, password)
}

// StudentLogin 学生登录，通过学号、身份证号和评教任务ID验证身份
//
// 参数:
//   - stuNo string 学号
//   - cardNo string 身份证号
//   - taskId uint 评教任务ID
//
// 返回值:
//   - *model.Student 学生信息
//   - error 登录失败时返回错误
func (a *AuthUseCase) StudentLogin(stuNo, cardNo string, taskId uint) (*model.Student, error) {
	return a.baseDal.StudentLogin(stuNo, cardNo, taskId)
}

// GetStudentInfo 根据学号获取学生信息
//
// 参数:
//   - stuNo string 学号
//
// 返回值:
//   - *model.Student 学生信息
//   - error 未找到学生时返回错误
func (a *AuthUseCase) GetStudentInfo(stuNo string) (*model.Student, error) {
	return a.baseDal.GetStudentByStudentNo(stuNo)
}

// AdminChangePassword 管理员修改密码，验证旧密码后更新为新密码
//
// 参数:
//   - username string 管理员用户名
//   - oldPassword string 旧密码
//   - newPassword string 新密码
//
// 返回值:
//   - error 修改失败时返回错误（旧密码错误、用户不存在等）
func (a *AuthUseCase) AdminChangePassword(username, oldPassword, newPassword string) error {
	return a.baseDal.AdminChangePassword(username, oldPassword, newPassword)
}