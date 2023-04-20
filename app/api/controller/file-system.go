package controller

import (
	"github.com/gin-gonic/gin"
	"inis/app/facade"
	"strings"
)

type FileSystem struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *FileSystem) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{

	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *FileSystem) IPOST(ctx *gin.Context) {

	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"upload": this.upload,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPUT - PUT请求本体
func (this *FileSystem) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{

	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *FileSystem) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{

	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *FileSystem) INDEX(ctx *gin.Context) {
	this.json(ctx, nil, "没什么用！", 200)
}

func (this *FileSystem) upload(ctx *gin.Context)  {

	params := this.params(ctx, map[string]any{
		"path": "/runtime/",
	})

	// 上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 保存文件
	err = ctx.SaveUploadedFile(file, file.Filename)
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	this.json(ctx, params, "上传成功！", 200)
}