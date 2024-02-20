package svc

import (
	"github.com/guojia99/cubing-pro/backend/api/internal/config"
	"github.com/guojia99/cubing-pro/backend/api/internal/middleware"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config             config.Config
	JwtInterceptor     rest.Middleware
	TokenInterceptor   rest.Middleware
	UserAuthMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:             c,
		JwtInterceptor:     middleware.NewJwtInterceptorMiddleware().Handle,
		TokenInterceptor:   middleware.NewTokenInterceptorMiddleware().Handle,
		UserAuthMiddleware: middleware.NewUserAuthMiddleware().Handle,
	}
}
