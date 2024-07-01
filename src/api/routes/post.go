package routes

import (
	"github.com/gin-gonic/gin"
	posts "github.com/guojia99/cubing-pro/src/api/app/post"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func PostRouters(router *gin.RouterGroup, svc *svc.Svc) {

	topics := router.Group("/topic")
	{
		topics.GET("/", posts.GetTopics(svc))                            // 帖子列表
		topics.GET("/:topicId", posts.GetTopic(svc, true))               // 帖子详情
		topics.GET("/:topicId/posts", posts.GetPosts(svc, false, false)) // 评论列表
	}

	// 权限
	topicsAuth := topics.Group(
		"/auth",
		middleware.CheckAuthMiddlewareFunc(user2.AuthPlayer),
	)
	{
		topicsAuth.POST("/", posts.CreateTopic(svc))                 // 编写帖子
		topicsAuth.PUT("/:topicId", posts.ReleaseTopic(svc))         // 发布帖子
		topicsAuth.DELETE("/:topicId", posts.DeleteTopic(svc, true)) // 删除帖子

		topicsAuth.GET("/posts", posts.GetPosts(svc, true, true))                // 我的评论
		topicsAuth.POST("/:topicId/posts", posts.CreatePost(svc))                // 发表评论
		topicsAuth.POST("/:topicId/posts/:post_id/reply", posts.CreatePost(svc)) // 回复评论
		topicsAuth.DELETE("/posts/:post_id", posts.DeletePost(svc, true))        // 删除评论
	}
}
