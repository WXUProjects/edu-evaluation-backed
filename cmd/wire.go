//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"edu-evaluation-backed/internal/biz"
	"edu-evaluation-backed/internal/conf"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/dal"
	"edu-evaluation-backed/internal/server"
	"edu-evaluation-backed/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, service.ProviderSet, data.ProviderSet, biz.ProviderSet, dal.ProviderSet, newApp))
}
