package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"inis/app/facade"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	debugs "runtime/debug"
	"strings"
	"time"
)

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		start := time.Now()

		ctx.Next()

		params, _ := ctx.Get("params")

		if cast.ToBool(facade.LogToml.Get("on", true)) {
			facade.Log.Info(map[string]any{
				"path":   ctx.Request.URL.Path,
				"method": ctx.Request.Method,
				// "headers": ctx.Request.Header,
				// "query":      ctx.Request.URL.RawQuery,
				// "params":     ctx.Request.Form,
				"params":     params,
				"ip":         ctx.ClientIP(),
				"user-agent": ctx.Request.UserAgent(),
				"errors":     ctx.Errors.ByType(gin.ErrorTypePrivate).String(),
				"cost":       time.Since(start).String(),
			}, "middleware")
		}
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(debug ...bool) gin.HandlerFunc {
	if len(debug) == 0 {
		debug = append(debug, false)
	}
	if !debug[0] {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				var broken bool
				if network, ok := err.(*net.OpError); ok {
					if system, ok := network.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(system.Error()), "broken pipe") || strings.Contains(strings.ToLower(system.Error()), "connection reset by peer") {
							broken = true
						}
					}
				}

				request, _ := httputil.DumpRequest(ctx.Request, false)
				if broken {
					facade.Log.Error(map[string]any{
						"path":    ctx.Request.URL.Path,
						"request": string(request),
					}, "middleware")
					ctx.Error(err.(error))
					ctx.Abort()
					return
				}

				facade.Log.Error(map[string]any{
					"path":    ctx.Request.URL.Path,
					"error":   err,
					"stack":   string(debugs.Stack()),
					"request": string(request),
				}, "[Recovery from panic]")

				var stack []string
				for i, item := range strings.Split(string(debugs.Stack()), "\n") {
					if i > 0 && len(item) > 0 {
						stack = append(stack, strings.TrimSpace(item))
					}
				}

				// 跳过首页的错误
				if ctx.Request.URL.Path == "/" {
					ctx.Next()
					return
				}

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg" : err.(error).Error(),
					"data":	stack,
				})
				ctx.Abort()
				return
			}
		}()
		ctx.Next()
	}
}
