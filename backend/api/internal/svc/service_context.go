package svc

import (
	"github.com/guojia99/cubing-pro/backend/api/internal/config"
	"github.com/guojia99/cubing-pro/backend/api/internal/middleware"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config              config.Config
	TokenInterceptor    rest.Middleware
	UserAuthMiddleware  rest.Middleware
	UserLevelMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:              c,
		TokenInterceptor:    middleware.NewTokenInterceptorMiddleware().Handle,
		UserAuthMiddleware:  middleware.NewUserAuthMiddleware().Handle,
		UserLevelMiddleware: middleware.NewUserLevelMiddleware().Handle,
	}
}
