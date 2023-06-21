package route

import (
	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
	"github.com/radovskyb/watcher"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"inis/app/index/controller"
	"runtime/debug"
	"time"
)

func Route(Gin *gin.Engine) {

	// 拦截异常
	defer func() {
		if err := recover(); err != nil {
			facade.Log.Error(map[string]any{
				"error":     err,
				"stack":     string(debug.Stack()),
				"func_name": utils.Caller().FuncName,
				"file_name": utils.Caller().FileName,
				"file_line": utils.Caller().Line,
			}, "首页模板渲染发生错误")
		}
	}()

	// 开启监听
	go watch(Gin)

	// 注册路由
	Gin.GET("/", controller.Index)
}

// watch - 监听 public/index.html 文件变化
func watch(Gin *gin.Engine) {

	// 加载模板
	Gin.LoadHTMLGlob("public/index.html")

	item := watcher.New()
	defer item.Close()

	// 定时器监听文件是否存在
	timer := func() {
		err := gocron.Every(1).Seconds().Do(html(Gin))
		if err != nil {
			return
		}
	}

	// 只监听public目录下的index.html文件
	if err := item.Add("public/index.html"); err != nil {
		facade.Log.Error(map[string]any{
			"error":     err,
			"stack":     string(debug.Stack()),
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "监听 public/index.html 文件变化发生错误")
	}

	go func() {
		for {
			select {
			case event := <- item.Event:
				if event.Op == watcher.Write {
					Gin.LoadHTMLGlob("public/index.html")
				}
			case <- item.Error:
				timer()
			case <- item.Closed:
				timer()
			}
		}
	}()

	// 开始监听
	if err := item.Start(time.Second); err != nil {
		timer()
	}
}

// html - 定时器监听 public/index.html 文件是否存在
func html(Gin *gin.Engine) func()  {
	return func() {
		// 存在则加载
		if exist := utils.File().Exist("public/index.html"); exist {
			// 开启监听
			go watch(Gin)
			// 删除定时器
			go gocron.Remove(html(Gin))
		}
	}
}