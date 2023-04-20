package route

import (
	"github.com/gin-gonic/gin"
	"inis/app/socket/controller"
	"inis/app/socket/middleware"
)

func Route(engine *gin.Engine) {
	socket := engine.Group("/socket", middleware.App)
	{
		class := controller.Index{}
		socket.GET("", class.Connect)
		socket.GET("/", class.Connect)
	}
}
