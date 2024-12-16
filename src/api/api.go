package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"os"

	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/api/routes"
)

type API struct {
	Svc *svc.Svc

	engine *gin.Engine
}

func NewAPI(svc *svc.Svc) *API {
	if svc.Cfg.APIConfig.StaticPath != "" {
		_ = os.MkdirAll(svc.Cfg.APIConfig.StaticPath, os.ModePerm)
	}
	//if svc.Cfg.APIConfig.AssetsPath != "" {
	//	_ = os.MkdirAll(svc.Cfg.APIConfig.AssetsPath, os.ModePerm)
	//}

	a := &API{
		Svc:    svc,
		engine: gin.New(),
	}

	a.engine.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.CorsMiddleware(),
	)
	//a.engine.RedirectTrailingSlash = false
	//gin.SetMode(gin.ReleaseMode)

	// init middleware
	middleware.InitJWT(svc)
	middleware.InitCheckAuth(svc)

	// init routers
	group := a.engine.Group(
		"/v3/cube-api",
		//middleware.CheckHeaderMiddleware,

	)
	routes.AuthRouters(group, svc)
	routes.AdminRouters(group, svc)
	routes.CompWithOrgRouters(group, svc)
	routes.CompWithUserRouters(group, svc)
	routes.PostRouters(group, svc)
	routes.PublicRouters(group, svc)
	routes.StaticRouters(group, svc)

	//group.Static("/assets", svc.Cfg.APIConfig.AssetsPath)
	return a
}

func (a *API) Run(host string, post int) error {
	return a.engine.Run(fmt.Sprintf("%s:%d", host, post))
}
