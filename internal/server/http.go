package server

import (
	"edu-evaluation-backed/api/v1/base_info/student"
	"edu-evaluation-backed/api/v1/base_info/teacher"
	"edu-evaluation-backed/internal/conf"
	"edu-evaluation-backed/internal/service/base_info"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server,
	studentService *base_info.StudentService,
	teacherService *base_info.TeacherService,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	b := srv.Route("/api/v1/base-info")
	// 一些上传路由
	b.POST("/student/import", studentService.Import)
	b.POST("/teacher/import", teacherService.Import)
	student_i.RegisterStudentHTTPServer(srv, studentService)
	teacher_i.RegisterTeacherHTTPServer(srv, teacherService)
	return srv
}
