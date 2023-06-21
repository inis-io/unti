package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"inis/app/api/controller"
	middle "inis/app/api/middleware"
	global "inis/app/middleware"
)

func Route(Gin *gin.Engine) {

	// 全局中间件
	group := Gin.Group("/api/").Use(
		global.Params(),    // 解析参数
		middle.Jwt(),       // 验证权限
	)

	// 允许动态挂载的路由
	allow := map[string]controller.ApiInterface{
		"test":          &controller.Test{},
		"comm":          &controller.Comm{},
		"file":          &controller.File{},
		"users":         &controller.Users{},
		"proxy":         &controller.Proxy{},
	}

	// 动态配置路由
	for key, item := range allow {
		group.Any(key, item.INDEX)
		group.GET(fmt.Sprintf("%s/:method", key), item.IGET)
		group.PUT(fmt.Sprintf("%s/:method", key), item.IPUT)
		group.POST(fmt.Sprintf("%s/:method", key), item.IPOST)
		group.DELETE(fmt.Sprintf("%s/:method", key), item.IDEL)
	}
}
