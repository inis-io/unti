package controller

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"image"
	"inis/app/facade"
	"mime/multipart"
	"regexp"
	"strings"
)

type File struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *File) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"rand": this.rand,
		"to-base64": this.toBase64,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *File) IPOST(ctx *gin.Context) {

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
func (this *File) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *File) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *File) INDEX(ctx *gin.Context) {
	this.json(ctx, nil, facade.Lang(ctx, "没什么用！"), 200)
}

// upload - 简单文件上传
func (this *File) upload(ctx *gin.Context) {

	params := this.params(ctx)

	// 上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 文件数据
	Byte, err := file.Open()
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}
	defer func(bytes multipart.File) {
		err := bytes.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(Byte)

	// 文件后缀
	suffix := file.Filename[strings.LastIndex(file.Filename, "."):]
	params["suffix"] = suffix

	item := facade.Storage.Upload(facade.Storage.Path()+suffix, Byte)
	if item.Error != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	this.json(ctx, map[string]any{
		"path": item.Domain + item.Path,
	}, facade.Lang(ctx, "上传成功！"), 200)
}

// rand - 随机图
func (this *File) rand(ctx *gin.Context) {

	params := this.params(ctx)

	root := "public"
	path := root + "/storage/rand/"

	// 分别获取  目录下的文件和目录
	info := utils.File().DirInfo(path)

	if info.Error != nil {
		this.json(ctx, nil, info.Error.Error(), 400)
		return
	}

	// 获取目录下的文件
	fnDir  := func(path any) (slice []string) {
		item := utils.File().Dir(path).List()
		if item.Error != nil {
			return []string{}
		}
		for _, val := range item.Slice {
			// 替换 root 为域名
			val  = strings.Replace(cast.ToString(val), root, cast.ToString(this.get(ctx, "domain")), 1)
			slice = append(slice, cast.ToString(val))
		}
		return slice
	}
	// 读取文件内容
	fnFile := func(path any) (slice []string) {
		item := utils.File().Path(path).Byte()
		if item.Error != nil {
			return []string{}
		}
		for _, val := range strings.Split(item.Text, "\n") {
			// 过滤末尾的 /r
			slice = append(slice, strings.TrimRight(val, "\r"))
		}
		return slice
	}

	var list []string
	result := cast.ToStringMap(info.Result)

	// 读取系统内全部的图片
	fnAll := func(result map[string]any) (slice []string) {

		if !utils.Is.Empty(result["dirs"]) {
			for _, value := range cast.ToStringSlice(result["dirs"]) {
				list = append(list, fnDir(path + value)...)
			}
		}

		if !utils.Is.Empty(result["files"]) {
			for _, value := range cast.ToStringSlice(result["files"]) {
				list = append(list, fnFile(path + value)...)
			}
		}

		return slice
	}

	// 没有指定目录或文件
	if utils.Is.Empty(params["name"]) {
		list = append(list, fnAll(result)...)
	}

	// 指定目录
	if utils.InArray[string](cast.ToString(params["name"]), cast.ToStringSlice(result["dirs"])) {
		list = fnDir(path + cast.ToString(params["name"]))
	}
	// 指定文件
	if utils.InArray[string](cast.ToString(params["name"]), cast.ToStringSlice(result["files"])) {
		list = fnFile(path + cast.ToString(params["name"]))
	}

	if utils.Is.Empty(list) {
		this.json(ctx, nil, facade.Lang(ctx, "无图！"), 400)
		return
	}

	// 远程图片地址
	url := list[utils.Rand.Int(len(list))]

	if cast.ToBool(params["json"]) {
		this.json(ctx, url, facade.Lang(ctx, "随机图！"), 200)
		return
	}

	if cast.ToBool(params["redirect"]) {
		ctx.Redirect(302, url)
		return
	}

	// 读取远程图片内容
	item := utils.Curl().Url(url).Send()
	if item.Error != nil {
		this.json(ctx, nil, item.Error.Error(), 400)
		return
	}

	// 正则表达式，匹配图片尺寸
	reg := regexp.MustCompile(`^(\d+)\D+(\d+)$`)
	// 从 query 的 size 中获取图片尺寸
	match := reg.FindStringSubmatch(ctx.Query("size"))

	var err error
	var write int
	img := item.Byte

	if match != nil {

		width  := cast.ToInt(match[1])
		height := cast.ToInt(match[2])

		// 文件后缀 - 转小写
		ext := strings.ToLower(url[strings.LastIndex(url, ".")+1:])
		// 图片压缩
		img = compress(ctx, img, width, height, ext)
	}

	// 输出图片到页面上
	ctx.Writer.Header().Set("Content-Type", "image/jpeg")
	ctx.Writer.Header().Set("Content-Length", cast.ToString(len(img)))

	write, err = ctx.Writer.Write(img)
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}
	if write != len(item.Byte) {
		this.json(ctx, nil, "写入失败！", 400)
		return
	}
}

// toBase64 - 网络图片转 base64
func (this *File) toBase64(ctx *gin.Context) {

	// 请求参数
	params := this.params(ctx)

	if utils.Is.Empty(params["url"]) {
		this.json(ctx, nil, facade.Lang(ctx, "%s 不能为空！", "url"), 400)
		return
	}

	// 读取远程图片内容
	item := utils.Curl().Url(cast.ToString(params["url"])).Send()
	if item.Error != nil {
		this.json(ctx, nil, item.Error.Error(), 400)
		return
	}
	if item.StatusCode != 200 {
		this.json(ctx, nil, fmt.Sprintf("状态码：%d", item.StatusCode), 400)
		return
	}

	// 转 base64
	res := fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(item.Byte))
	this.json(ctx, res, facade.Lang(ctx, "成功！"), 200)

	// ctx.Writer.Header().Set("Content-Type", "image/jpeg")
	// ctx.Writer.Header().Set("Content-Length", cast.ToString(len(item.Byte)))
	// ctx.Writer.Write(item.Byte)
}

// compress - 图片压缩
func compress(ctx *gin.Context, byte []byte, width, height int, ext string) (result []byte) {

	// 图片处理模式
	var mode string
	mode = utils.Ternary[string](width == height, "fill", mode)
	mode = ctx.DefaultQuery("mode", mode)

	// byte 转 image
	src, _ := imaging.Decode(bytes.NewReader(byte))

	// 处理图片
	var dstImage *image.NRGBA
	switch mode {
	case "fill":
		// 填充
		dstImage = imaging.Fill(src, width, height, imaging.Center, imaging.Lanczos)
	case "resize":
		// 完全自定义大小
		dstImage = imaging.Resize(src, width, height, imaging.Lanczos)
	case "fit":
		// 等比例缩放
		dstImage = imaging.Fit(src, width, height, imaging.Lanczos)
	default:
		// 等比例缩放
		dstImage = imaging.Fit(src, width, height, imaging.Lanczos)
	}

	// 压缩的图片格式
	var format imaging.Format
	switch ext {
	case "jpg", "jpeg":
		format = imaging.JPEG
	case "png":
		format = imaging.PNG
	case "gif":
		format = imaging.GIF
	case "tif", "tiff":
		format = imaging.TIFF
	case "bmp":
		format = imaging.BMP
	default:
		format = imaging.JPEG
	}
	buffer := new(bytes.Buffer)
	// 解决GIG压缩之后不会动的问题
	_ = imaging.Encode(buffer, dstImage, format)

	return buffer.Bytes()
}