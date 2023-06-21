package config

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"image"
	"inis/app/facade"
	"inis/app/middleware"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Gin - gin引擎
var Gin *gin.Engine

// AppToml - App配置文件
var AppToml *utils.ViperResponse

// Server - 服务
var Server *http.Server

func init() {

	// 初始化配置文件
	initAppToml()
	// 初始化
	InitApp()
}

// initAppToml - 初始化APP配置文件
func initAppToml() {

	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "app",
		Content: utils.Replace(facade.TempApp, nil),
	}).Read()

	if item.Error != nil {
		fmt.Println("APP配置文件初始化发生错误", item.Error)
		return
	}

	AppToml = &item
}

// InitApp 初始化App
func InitApp() {

	debug := cast.ToBool(AppToml.Get("app.debug", false))

	// 关闭 gin 的日志
	if !debug {
		// 设置 release 模式
		gin.SetMode(gin.ReleaseMode)
		// 关闭 gin 的日志
		gin.DefaultWriter = io.Discard
	}

	Gin = gin.Default()
	// 处理资源路由 和 404路由
	notRoute(Gin)
	// 打印版本信息
	console()

	// 全局日志处理
	Gin.Use(middleware.GinLogger(), middleware.GinRecovery(true))
}

// Use 注册配置
func Use(args ...func(*gin.Engine)) {
	var opts []func(*gin.Engine)
	opts = append(opts, args...)
	for _, fn := range args {
		fn(Gin)
	}
}

// Run 启动服务
func Run(callback ...func()) {

	// 启动服务
	for _, fn := range callback {
		fn()
	}

	port := ":" + cast.ToString(AppToml.Get("app.port", 8080))

	Server = &http.Server{
		Addr:    port,
		Handler: Gin,
	}

	go func() {
		if err := Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("服务启动失败", err)
		}
	}()

	// 保持主线程不退出
	select {}
}

// 路由不存在
func notRoute(Gin *gin.Engine) {
	Gin.NoRoute(func(ctx *gin.Context) {

		// 拦截异常
		defer func() {
			if err := recover(); err != nil {
				ctx.JSON(200, gin.H{"code": 500, "msg": "服务器内部错误！", "data": nil})
			}
		}()

		ctx.Status(200)

		// 获取请求的路径
		path := ctx.Request.URL.Path
		// 页面资源
		page := []any{"/", "/index.htm", "/index.html", "/index.php", "/index.jsp"}
		imgs := []any{"jpg", "jpeg", "png", "gif", "tif", "tiff", "bmp"}

		// path 以 / 分隔，取最后一个到末尾
		prefix := path[:strings.LastIndex(path, "/")]
		// 文件名
		fileName := path[strings.LastIndex(path, "/"):]
		// 文件后缀 - 转小写
		ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])

		// 判断文件是否存在
		IsExist := func(path string) (check bool) {
			// 判断 path 是否以 public 开头
			if !strings.HasPrefix(path, "public") {
				path = "public/" + path
			}
			exist := utils.File().Exist(path)
			if !exist {
				ctx.JSON(200, gin.H{"code": 400, "msg": "资源不存在！", "data": nil})
			}
			return exist
		}
		// 输出错误图片
		WriteByte := func(path string) {
			_, err := ctx.Writer.Write(utils.File().Byte("public/assets/images/gif/" + path).Byte)
			if err != nil {
				ctx.JSON(200, gin.H{"code": 400, "msg": "资源不存在！", "data": nil})
			}
		}

		switch {
		// 页面文件
		case utils.In.Array(fileName, page):
			if check := IsExist("public/" + prefix + "/index.html"); check {
				ctx.Header("Content-Type", "text/html; charset=utf-8")
				_, err := ctx.Writer.Write(utils.File().Byte("public" + prefix + "/index.html").Byte)
				if err != nil {
					ctx.JSON(200, gin.H{"code": 400, "msg": "资源不存在！", "data": nil})
					break
				}
			}
		// 图片文件 - 条件压缩处理
		case utils.In.Array(ext, imgs):

			// 设置文件类型
			ctx.Header("Content-Type", utils.Mime.Type(ext)+"; charset=utf-8")
			exist := utils.File().Exist("public" + path)
			if !exist {
				WriteByte("404.gif")
				break
			}

			// 正则表达式，匹配图片尺寸
			reg := regexp.MustCompile(`^(\d+)\D+(\d+)$`)
			// 从 query 的 size 中获取图片尺寸
			match := reg.FindStringSubmatch(ctx.Query("size"))

			if match != nil {

				width := cast.ToInt(match[1])
				height := cast.ToInt(match[2])

				// 图片处理模式
				var mode string
				mode = utils.Ternary[string](width == height, "fill", mode)
				mode = ctx.DefaultQuery("mode", mode)

				src, err := imaging.Open("public" + path)
				if err != nil {
					WriteByte("error.gif")
					break
				}

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
				err = imaging.Encode(buffer, dstImage, format)

				if err != nil {
					WriteByte("error.gif")
					break
				}

				_, err = ctx.Writer.Write(buffer.Bytes())
				if err != nil {
					WriteByte("error.gif")
					break
				}

				break
			}

			_, err := ctx.Writer.Write(utils.File().Byte("public" + path).Byte)
			if err != nil {
				WriteByte("error.gif")
				break
			}
		// 其他文件
		case strings.Contains(fileName, "."):

			exist := IsExist("public" + path)
			if !exist {
				break
			}

			if check := IsExist(path); check {
				// 设置文件类型
				ctx.Header("Content-Type", utils.Mime.Type(ext)+"; charset=utf-8")
				_, err := ctx.Writer.Write(utils.File().Byte("public" + path).Byte)
				if err != nil {
					ctx.JSON(200, gin.H{"code": 400, "msg": "文件读取失败！", "data": err.Error()})
					break
				}
			}
		// 路由未定义
		default:
			ctx.JSON(200, gin.H{"code": 400, "msg": "路由未定义！", "data": nil})
		}
	})
}

// console 控制台
func console() {
	port := AppToml.Get("app.port", 8080)
	char := "/** \n" +
		" *   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!`   `4!!!!!!!!!!~4!!!!!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!   <~:   ~!!!~   ..  4!!!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!  ~~~~~~~  '  ud$$$$$  !!!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!  ~~~~~~~~~: ?$$$$$$$$$  !!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!`     ``~!!!!!!!!!!!!!!  ~~~~~          \"*$$$$$k `!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!  $$$$$bu.  '~!~`     .  '~~~~      :~~~~          `4!!!!!!!!!!!\n" +
		" *   !!!!!!!!!  $$$$$$$$$$$c  .zW$$$$$E ~~~~      ~~~~~~~~  ~~~~~:  '!!!!!!!!!!\n" +
		" *   !!!!!!!!! d$$$$$$$$$$$$$$$$$$$$$$E ~~~~~    '~~~~~~~~    ~~~~~  !!!!!!!!!!\n" +
		" *   !!!!!!!!> 9$$$$$$$$$$$$$$$$$$$$$$$ '~~~~~~~ '~~~~~~~~     ~~~~  !!!!!!!!!!\n" +
		" *   !!!!!!!!> $$$$$$$$$$$$$$$$$$$$$$$$b   ~~~    '~~~~~~~     '~~~ '!!!!!!!!!!\n" +
		" *   !!!!!!!!> $$$$$$$$$$$$$$$$$$$$$$$$$$$cuuue$$N.   ~        ~~~  !!!!!!!!!!!\n" +
		" *   !!!!!!!!! **$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$Ne  ~~~~~~~~  `!!!!!!!!!!!\n" +
		" *   !!!!!!!!!  J$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$N  ~~~~~  zL '!!!!!!!!!!\n" +
		" *   !!!!!!!!  d$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$c     z$$$c `!!!!!!!!!\n" +
		" *   !!!!!!!> <$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$> 4!!!!!!!!\n" +
		" *   !!!!!!!  $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$  !!!!!!!!\n" +
		" *   !!!!!!! <$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$*\"   ....:!!\n" +
		" *   !!!!!!~ 9$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$e@$N '!!!!!!!\n" +
		" *   !!!!!!  9$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$  !!!!!!!\n" +
		" *   !!!!!!  $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$\"\"$$$$$$$$$$$~ ~~4!!!!\n" +
		" *   !!!!!!  9$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$    $$$$$$$Lue  :::!!!!\n" +
		" *   !!!!!!> 9$$$$$$$$$$$$\" '$$$$$$$$$$$$$$$$$$$$$$$$$$$    $$$$$$$$$$  !!!!!!!\n" +
		" *   !!!!!!! '$$*$$$$$$$$E   '$$$$$$$$$$$$$$$$$$$$$$$$$$$u.@$$$$$$$$$E '!!!!!!!\n" +
		" *   !!!!~`   .eeW$$$$$$$$   :$$$$$$$$$$$$$***$$$$$$$$$$$$$$$$$$$$u.    `~!!!!!\n" +
		" *   !!> .:!h '$$$$$$$$$$$$ed$$$$$$$$$$$$Fz$$b $$$$$$$$$$$$$$$$$$$$$F '!h.  !!!\n" +
		" *   !!!!!!!!L '$**$$$$$$$$$$$$$$$$$$$$$$ *$$$ $$$$$$$$$$$$$$$$$$$$F  !!!!!!!!!\n" +
		" *   !!!!!!!!!   d$$$$$$$$$$$$$$$$$$$$$$$$buud$$$$$$$$$$$$$$$$$$$$\"  !!!!!!!!!!\n" +
		" *   !!!!!!! .<!  #$$*\"$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$*  :!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!!:   d$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$#  :!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!~  :  '#$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$*\"    !!!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!  !!!!!:   ^\"**$$$$$$$$$$$$$$$$$$$$**#\"     .:<!!!!!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!!!!!!!!!!!:...                      .::!!!!!!!!!!!!!!!!!!!!!!!!\n" +
		" *   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n" +
		" *                          版本号：%s  端口：%d\n" +
		" *                              兔子：服务已启动\n" +
		" **/"
	fmt.Println(fmt.Sprintf(char, facade.Version, port))
}