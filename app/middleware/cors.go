package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ctx.Header("Access-Control-Max-Age", "1800")
		ctx.Header("Access-Control-Allow-Origin", "*")
		// ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Content-Type", "application/json; charset=utf-8")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS, PATCH")
		ctx.Header("Access-Control-Allow-Headers", "Token, Authorization, i-api-key, Content-Type, If-Match, If-Modified-Since, If-None-Match, If-Unmodified-Since, X-CSRF-TOKEN, X-Requested-With")
		ctx.Header("Access-Control-Expose-Headers", "Content-Type, Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")

		// 放行所有OPTIONS方法
		if strings.ToUpper(ctx.Request.Method) == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}

		ctx.Next()
	}
}
