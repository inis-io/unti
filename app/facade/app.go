package facade

import (
	"github.com/fsnotify/fsnotify"
	"github.com/unti-io/go-utils/utils"
)

type H map[string]any

// AppToml - App配置文件
var AppToml *utils.ViperResponse

// initAppToml - 初始化App配置文件
func initAppToml() {

	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "app",
		Content: utils.Replace(TempApp, nil),
	}).Read()

	if item.Error != nil {
		Log.Error(map[string]any{
			"error":     item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "Crypt配置初始化错误")
		return
	}

	AppToml = &item
}

func init() {
	// 初始化配置文件
	initAppToml()
	// 初始化缓存
	initApp()

	// 监听配置文件变化
	AppToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	AppToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initApp()
	})
}

// 初始化加密配置
func initApp() {

}