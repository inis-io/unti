package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"inis/app/api/controller"
	private "inis/app/api/middleware"
	global "inis/app/middleware"
)

func Route(engine *gin.Engine) {

	// 全局中间件 global.QpsGlobal(),
	api := engine.Group("/api/").Use(global.Cors(), global.QpsPoint(), global.QpsGlobal())

	// 公共接口 - 无权限
	allow := map[string]controller.ApiInterface{
		"comm": &controller.Comm{},
		"test": &controller.Test{},
	}

	// 动态配置路由
	for key, val := range allow {
		api.GET(key, val.INDEX)
		api.GET(fmt.Sprintf("%s/:method"   , key), val.IGET)
		api.PUT(fmt.Sprintf("%s/:method"   , key), val.IPUT)
		api.POST(fmt.Sprintf("%s/:method"  , key), val.IPOST)
		api.DELETE(fmt.Sprintf("%s/:method", key), val.IDEL)
	}

	// 需要权限的接口
	permission := map[string]controller.ApiInterface{
		"users": &controller.Users{},
		"oauth": &controller.OAuth{},
		"auth-group" : &controller.AuthGroup{},
		"auth-rules" : &controller.AuthRules{},
		"file-system": &controller.FileSystem{},
	}

	// 动态配置路由
	for key, val := range permission {

		// 追加中间件
		item := api.Use(private.Jwt())

		item.GET(key, val.INDEX)
		item.GET(fmt.Sprintf("%s/:method"   , key), val.IGET)
		item.PUT(fmt.Sprintf("%s/:method"   , key), val.IPUT)
		item.POST(fmt.Sprintf("%s/:method"  , key), val.IPOST)
		item.DELETE(fmt.Sprintf("%s/:method", key), val.IDEL)
	}
}