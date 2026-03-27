package server

import (
	"edu-evaluation-backed/api/v1/base_info/course"
	"edu-evaluation-backed/api/v1/base_info/student"
	"edu-evaluation-backed/api/v1/base_info/teacher"
	eva_task2 "edu-evaluation-backed/api/v1/eva_task"
	"edu-evaluation-backed/internal/conf"
	"edu-evaluation-backed/internal/service/auth"
	"edu-evaluation-backed/internal/service/base_info"
	"edu-evaluation-backed/internal/service/eva_task"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server,
	authService *auth.AuthService,
	studentService *base_info.StudentService,
	teacherService *base_info.TeacherService,
	courseService *base_info.CourseService,
	evaService *eva_task.EvaTaskService,
	logger log.Logger,
) *http.Server {
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
	b := srv.Route("/api/v1")
	// 认证路由
	b.POST("/auth/admin/login", authService.AdminLogin)
	b.POST("/auth/student/login", authService.StudentLogin)
	b.GET("/auth/student/info", authService.StudentInfo)
	// 上传路由
	b.POST("/base-info/student/import", studentService.Import)
	b.POST("/base-info/teacher/import", teacherService.Import)
	b.POST("/base-info/course/import", courseService.Import)
	student_i.RegisterStudentHTTPServer(srv, studentService)
	teacher_i.RegisterTeacherHTTPServer(srv, teacherService)
	course.RegisterCourseHTTPServer(srv, courseService)
	eva_task2.RegisterTaskHTTPServer(srv, evaService)

	return srv
}
