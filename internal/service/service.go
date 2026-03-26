package service

import (
	"edu-evaluation-backed/internal/service/base_info"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(base_info.NewStudentService, base_info.NewTeacherService)
