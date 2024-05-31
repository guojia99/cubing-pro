package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func PostRouters(router *gin.RouterGroup, svc *svc.Svc) {
	post := router.Group("/post")
	{
		post.GET("/")           // 帖子列表
		post.GET("/:postId")    // 帖子详情
		post.POST("/")          // 发布帖子
		post.DELETE("/:postId") // 删除帖子

		comment := post.Group("/:postId/comments")
		{
			comment.GET("/")                  // 评论列表
			comment.POST("/")                 // 发表评论
			comment.DELETE("/")               // 删除评论
			comment.PUT("/:commentsId/reply") // 回复评论
		}
	}
}
