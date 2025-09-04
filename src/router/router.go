package router

import (
	"challenge/api/controller"
	"challenge/config"
	"challenge/middleware"

	"github.com/gin-gonic/gin"
)

func Init() {
	router := NewRouter()
	router.Run(config.Appconfig.GetString("server.port"))
}

func NewRouter() *gin.Engine {
	router := gin.New()
	resource := router.Group("/api")
	resource.Use(middleware.LogRequestInfo())
	{
		resource.POST("/question", controller.SubmitQuestion)
	}
	return router
}
