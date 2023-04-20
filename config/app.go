package config

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"image"
	"inis/app/middleware"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Engine - gin引擎
var Engine *gin.Engine
// AppToml - App配置文件
var AppToml *utils.ViperResponse
// SSLToml - SSL配置文件
var SSLToml *utils.ViperResponse

var Server *http.Server

func init() {

	// 初始化配置文件
	initAppToml()
	initSSLToml()
	// 初始化
	InitApp()
	initSSL()

	// 监听配置文件变化
	SSLToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	SSLToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initSSL()
	})
}

// initAppToml - 初始化APP配置文件
func initAppToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "app",
		Content: `# ======== 基础服务配置 - 修改此文件建议重启服务 ========

# 应用配置
[app]
# 项目运行端口
port        = 1000
# 调试模式
debug       = false
# 版本号
version     = "2.0.0"
# 登录token名称（别乱改，别作死）
token_name  = "unti_login_token"

# JWT加密配置
[jwt]
# jwt密钥
secret   = "unti_api_key"
# 过期时间(秒)
expire   = 604800

# API限流器配置
[qps]
# 单个接口每秒最大请求数
point    = 10
# 全局接口每秒最大请求数
global   = 50
`,
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

	engine := gin.Default()
	// 处理资源路由 和 404路由
	notRoute(engine)
	// 打印版本信息
	console()

	// 全局日志处理
	engine.Use(middleware.GinLogger(), middleware.GinRecovery(true))

	// 设置引擎
	Engine = engine
}

// initSSLToml - 初始化SSL配置文件
func initSSLToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "ssl",
		Content: `# ======== SSL 配置 ========

# 是否启用SSL
on   	 = false
# SSL域名
host     = "localhost"
# 真实端口
port     = 1000
# 证书文件
cert     = "./config/ssl/localhost.crt"
# 私钥文件
key      = "./config/ssl/localhost.key"
`,
	}).Read()

	if item.Error != nil {
		fmt.Println("发生错误", item.Error)
		return
	}

	SSLToml = &item
}

// 初始化SSL
func initSSL() {

}

// Use 注册配置
func Use(args ...func(*gin.Engine)) {
	var opts []func(*gin.Engine)
	opts = append(opts, args...)
	for _, val := range args {
		val(Engine)
	}
}

// 路由不存在
func notRoute(engine *gin.Engine) {
	engine.NoRoute(func(ctx *gin.Context) {

		ctx.Status(200)

		// 获取请求的路径
		path := ctx.Request.URL.Path
		// 页面资源
		page := []any{"/", "/index.htm", "/index.html", "/index.php", "/index.jsp"}
		imgs := []any{"jpg", "jpeg", "png", "gif", "tif", "tiff", "bmp"}

		// path 以 / 分隔，取最后一个到末尾
		prefix   := path[:strings.LastIndex(path, "/")]
		// 文件名
		fileName := path[strings.LastIndex(path, "/"):]
		// 文件后缀 - 转小写
		ext      := strings.ToLower(fileName[strings.LastIndex(fileName, ".") + 1:])

		// 判断文件是否存在
		IsExist := func(path string) (check bool) {
			exist := utils.File().IsExist(path)
			if !exist {
				ctx.JSON(200, gin.H{"code": 400,"msg":  "资源不存在！","data": nil})
			}
			return exist
		}
		// 输出错误图片
		WriteByte := func(path string) {
			_, err := ctx.Writer.Write(utils.File().Byte("public/assets/images/gif/" + path).Byte)
			if err != nil {
				ctx.JSON(200, gin.H{"code": 400,"msg": "资源不存在！","data": nil})
			}
		}

		switch {
		// 页面文件
		case utils.In.Array(fileName, page):
			if check := IsExist("public/" + prefix + "/index.html"); check {
				ctx.Header("Content-Type", "text/html; charset=utf-8")
				_, err := ctx.Writer.Write(utils.File().Byte("public" + prefix + "/index.html").Byte)
				if err != nil {
					ctx.JSON(200, gin.H{"code": 400,"msg": "资源不存在！","data": nil})
					break
				}
			}
		// 图片文件 - 条件压缩处理
		case utils.In.Array(ext, imgs):

			// 设置文件类型
			ctx.Header("Content-Type", utils.Mime.Type(ext)+"; charset=utf-8")
			exist := utils.File().IsExist("public" + path)
			if !exist {
				WriteByte("404.gif")
				break
			}

			// 正则表达式，匹配图片尺寸
			reg   := regexp.MustCompile(`^(\d+)\D+(\d+)$`)
			// 从 query 的 size 中获取图片尺寸
			match := reg.FindStringSubmatch(ctx.Query("size"))

			if match != nil {

				width  := cast.ToInt(match[1])
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
					ctx.JSON(200, gin.H{"code": 400,"msg": "文件读取失败！","data": err.Error()})
					break
				}
			}
		// 路由未定义
		default:
			ctx.JSON(200, gin.H{"code": 400,"msg": "路由未定义！","data": nil})
		}
	})
}

// console 控制台
func console() {
	port := AppToml.Get("app.port", 8080)
	version := AppToml.Get("app.version", "1.0.0")
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
	fmt.Println(fmt.Sprintf(char, version, port))
	// fmt.Println(fmt.Sprintf("版本号：%s  端口：%d", version, port))
	// char := "/** \n" +
	// 	" *   MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM@@N%RYIY&$N@MMMMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMM@WEX]+!;!INMMMMMMMMMMMMMMMMMM$'...........&MMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMM%Xi,........:}@MMMMMMMMMMMMMMMMM%>..........:&@MMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMN}...........:*NMMMMMMMMMMMMMMMMN*:..........F@MMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMM$'...........~#MMMMMMMMMMMMMMMM@K`..........F@MMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMM#;...........>W@MMMMMMMMMMMMMMMW>..........X@MMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMM@*:..........,Y@MMMMMMMMMMMMMMM@}:...;!':..Y@MMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMX:...'ii~:..~E@MMMMMMMMMMMMMMMF'..`~>>,.!EMMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMN1..:;\">\"':.`*NMMMMMMMMMMMMMMME>..`~\"'`+$MMMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMM@l..:'\">!,..~RMMMMMMMMMMMMMMMNl..`~~,~&MMMMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMMMNi..:'>>~`.:*NMMMMMMMMMMMMMMM&,.:~\";,1%MMMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMMMMN}..:~>>;..!#MMMMMMMMMMMMMMM$!..`!~`..&MMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMMMMM%>..:'!;..:*@MMMMMMMMMMMMMMNIi+11\"`.'$MMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMMMMMM%}+Y$@@@M@@MMMMMMM@NN%%%%%NNNN@MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMMMMMMMMMMMMN@@%EFl+>~,`:.............:`;~>]K@@MMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMM@N@MMMMMMMMMMMMM@I'......................................lMMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMM$%MMMMMMMMMMMM$;.........................................lMMMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMNK1\"]&ENMMMMMME'...........................................+NMMMMMMMMMMMMMMMMMMM\n" +
	// 	" *   MW$R&1i&NMMMMMMM@}:...............................,/l>:........IMMMMMMMMMM@MMMMMMMM\n" +
	// 	" *   MMMMMMN&$MMMMMMMR;.........`/FFREl`.............`lF~.}EY,.......]@MMMMME*~1WMMMMMMM\n" +
	// 	" *   MMMMMM@##%MMMMMN]..........~Y/.!F#/:............:1K*i/*l~.......`FMMMMM$RE%@MMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMM#'..........:}R#FFXi:..............!XE#Fi`........`&MMMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMM@F`......::`,``\"lY]'..................::..`!/]lll]i'>$MMMMMMMMMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMMMN*';~1X&I1/iiii+/}*YY}':.:...........:,iX&]>''';;;'''\"]F*+l#@MMMMMMMM\n" +
	// 	" *   MMMMMMMMMMMMNY>!ilX]!;;;;;'''''~~>/*FY**Xl\":....;1***Il1!~~!~~~~~~';,'>>',\"&MMMMMMM\n" +
	// 	" *   MMMMMMMMMMI',~~!ii~;'~~~~~~~~~~''>1+~;~!~;\"Y&++I]!;'';;'>+\"''~~~~~~~~''\"i\";,,lNMMMM\n" +
	// 	" *   MMMMMMMM%+`'~''\"i\"''~~~~~~~~~~';~ii~;'~'~!~;+E$/,'\"~'~~';\"i\"''~~~~~~~~''>>!',,]NMMM\n" +
	// 	" *   MMMMMMMK'`;,,;'~~;,;;;;;'';;;;,,~\"~;,;;;,;;,;>I]';~;,;;;,;'!~;,;';''';;;'~';;,`+WMM\n" +
	// 	" *   MMMMMM@Y\"\"\"\"\"\"\"!!!!!!!!!\"\"!!!!!\"\"\"\"\"\"!!!!!\"\">+IY+\">>!!\"\"\"!!\"\"\"!!!!!!!!!\"\"!!\">!!]W@M\n"
	// 	fmt.Println(char)
}