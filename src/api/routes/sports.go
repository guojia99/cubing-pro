package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/sports"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func SportsRoutes(router *gin.RouterGroup, svc *svc.Svc) {
	sportsAdmin := router.Group("/sports/admin",
		middleware.JWT().MiddlewareFunc(),
		middleware.CheckAuthMiddlewareFunc(user2.AuthSuperAdmin),
	)
	{
		sportsAdmin.GET("/events", sports.ListSportEvents(svc))
		sportsAdmin.POST("/events", sports.CreateSportEvent(svc))
		sportsAdmin.DELETE("/events/:id", sports.DeleteSportEvent(svc))

		sportsAdmin.GET("/results", sports.ListSportResults(svc))
		sportsAdmin.POST("/results", sports.CreateSportResult(svc))
		sportsAdmin.DELETE("/results/:id", sports.DeleteSportResult(svc))
	}
}
