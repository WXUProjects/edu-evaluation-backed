package biz

import (
	"edu-evaluation-backed/internal/biz/base_info"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(base_info.NewStudentUseCase, base_info.NewTeacherUseCase)
