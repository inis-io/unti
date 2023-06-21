package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"inis/app/dev/controller"
	global "inis/app/middleware"
)

func Route(Gin *gin.Engine) {

	// 全局中间件
	install := Gin.Group("/dev/").Use(
		global.Params(),	// 解析参数
	)

	// 动态配置路由 - 允许动态挂载的路由
	for key, item := range map[string]controller.ApiInterface{
		"info":  &controller.Info{},
	} {
		install.Any(key, item.INDEX)
		install.GET(fmt.Sprintf("%s/:method", key), item.IGET)
		install.PUT(fmt.Sprintf("%s/:method", key), item.IPUT)
		install.POST(fmt.Sprintf("%s/:method", key), item.IPOST)
		install.DELETE(fmt.Sprintf("%s/:method", key), item.IDEL)
	}
}
