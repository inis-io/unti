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

		if cast.ToBool(facade.LogToml.Get("on", true)) {
			facade.Log.Info("middleware", map[string]any{
				"path"   : ctx.Request.URL.Path,
				"method" : ctx.Request.Method,
				// "headers": ctx.Request.Header,
				"query"  : ctx.Request.URL.RawQuery,
				"params" : ctx.Request.Form,
				"ip"     : ctx.ClientIP(),
				"user-agent": ctx.Request.UserAgent(),
				"errors" : ctx.Errors.ByType(gin.ErrorTypePrivate).String(),
				"cost"   : time.Since(start).String(),
			})
		}
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(debug ...bool) gin.HandlerFunc {
	if len(debug) == 0 {
		debug = append(debug, false)
	}
	if debug[0] {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
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
					facade.Log.Error("error", map[string]any{
						"path":  ctx.Request.URL.Path,
						"request": string(request),
					})
					// If the connection is dead, we can't write a status to it.
					ctx.Error(err.(error)) // nolint: errcheck
					ctx.Abort()
					return
				}

				facade.Log.Error("[Recovery from panic]", map[string]any{
					"path":  ctx.Request.URL.Path,
					"error": err,
					"stack": string(debugs.Stack()),
					"request": string(request),
				})

				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}