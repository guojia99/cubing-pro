package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"

	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/api/routes"
)

type API struct {
	Svc *svc.Svc

	engine *gin.Engine
}

func NewAPI(svc *svc.Svc) *API {
	a := &API{
		Svc:    svc,
		engine: gin.Default(),
	}

	// init middleware
	middleware.InitJWT(svc)
	middleware.InitCheckAuth(svc)

	// init routers
	group := a.engine.Group("/v3/cube-api")
	routes.AuthRouters(group, svc)
	routes.AdminRouters(group, svc)
	routes.UserRouters(group, svc)
	routes.CompRouters(group, svc)
	routes.PostRouters(group, svc)
	routes.PublicRouters(group, svc)
	return a
}

func (a *API) Run(host string, post int) error {
	return a.engine.Run(fmt.Sprintf("%s:%d", host, post))
}