package service

import (
	"edu-evaluation-backed/internal/service/auth"
	"edu-evaluation-backed/internal/service/base_info"
	"edu-evaluation-backed/internal/service/eva_task"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	auth.NewAuthService,
	base_info.NewStudentService,
	base_info.NewTeacherService,
	base_info.NewCourseService,
	eva_task.NewEvaTaskService,
)
