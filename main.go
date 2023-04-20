package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	api "inis/app/api/route"
	socket "inis/app/socket/route"
	app "inis/config"
)

func main() {

	// 监听服务
	watch()
	// 运行服务
	run()

	// 静默运行 - 不显示控制台
	// go build -ldflags -H=windowsgui 或 bee pack -ba="-ldflags -H=windowsgui"
}

func run() {
	// 注册路由
	app.Use(api.Route, socket.Route)
	// 运行服务
	app.Run()
}

func watch() {

	// 监听配置文件变化
	app.AppToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	app.AppToml.Viper.OnConfigChange(func(event fsnotify.Event) {

		// 关闭服务
		if app.Server != nil {
			// 释放路由
			// app.Server.Handler = nil
			// 关闭服务
			err := app.Server.Shutdown(nil)
			if err != nil {
				fmt.Println("关闭服务发生错误: ", err)
				return
			}
		}

		watch()
		// 重新初始化驱动
		app.InitApp()
		// 重新运行服务
		run()
	})
}