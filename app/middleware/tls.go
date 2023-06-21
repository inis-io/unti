package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unrolled/secure"
	"github.com/unti-io/go-utils/utils"
	"strings"
)

// Tls https 处理, 即变为 wss://*
func Tls() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 获取配置文件中的启动端口
		port := func() string {
			item := utils.Env().Get("app.port", ":8642")
			result := cast.ToString(item)

			// 判断 result 是否包含 :
			if !strings.Contains(result, ":") {
				result = ":" + result
			}
			return result
		}

		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     port(),
		})
		err := secureMiddleware.Process(ctx.Writer, ctx.Request)

		if err != nil {
			return
		}

		ctx.Next()
	}
}
