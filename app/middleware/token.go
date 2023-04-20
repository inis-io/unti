package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
)

// Token 简单 token 验证
func Token() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 检查请求头的 Authorization 字段是否存在
		auth := ctx.Request.Header.Get("Authorization")
		// 从请求参数中获取 token
		if utils.Is.Empty(auth) {
			auth = ctx.Query("token")
		}

		if utils.Is.Empty(auth) {
			ctx.JSON(200, gin.H{"data": nil, "code": 401, "msg": "未授权"})
			ctx.Abort()
			return
		}

		// token := cast.ToString(utils.Env().Get("app.token", "0147."))
		token := cast.ToString("0147.")

		// 检查请求头的 Authorization 字段是否正确
		if auth != token {
			ctx.JSON(200, gin.H{"data": nil, "code": 403, "msg": "无权限"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
