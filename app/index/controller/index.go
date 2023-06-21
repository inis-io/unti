package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/unti-io/go-utils/utils"
)

func Index(ctx *gin.Context) {

	ctx.Header("Content-Type", "text/html; charset=utf-8")

	ctx.HTML(200, "index.html", gin.H{
		"TITLE": "欢迎使用",
		"INIS":  utils.Json.Encode(map[string]any{
			"TEST": "这是服务端设置的数据",
		}),
	})
}