package middleware

import (
	"github.com/gin-gonic/gin"
)

func App(ctx *gin.Context) {
	ctx.Next()
}
