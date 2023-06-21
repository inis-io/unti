package model

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
)

func task() {
	// 如果存在安装锁，表示还没进行初始化安装，不进行自动迁移
	if !utils.File().Exist("install.lock") {
		// 结束任务
		gocron.Remove(task)
		// 初始化数据库
		facade.WatchDB(true)
		// 检查是否开启自动迁移
		if cast.ToBool(facade.NewToml(facade.TomlDb).Get("mysql.migrate")) {
			go InitTable()
		}
	}
}

// InitTable - 初始化数据库表 - 自动迁移
func InitTable() {

	allow := []func(){
		InitUsers,
	}

	for _, val := range allow {
		go val()
	}
}

func init() {
	if err := gocron.Every(1).Second().Do(task); err != nil {
		return
	}
	// 启动调度器
	gocron.Start()
}

// DomainTemp1 域名模板替换
func DomainTemp1() (replace map[string]any) {
	toml := facade.NewToml(facade.TomlStorage)
	replace = make(map[string]any)
	storage := []string{"oss", "cos", "kodo"}
	// 模板变量替换
	for _, val := range storage {
		// 优先使用配置文件中的域名
		if !utils.Is.Empty(toml.Get(val + ".domain")) {
			replace["{{"+val+"}}"] = cast.ToString(toml.Get(val + ".domain"))
			continue
		}
		// 如果配置文件中没有域名，则使用默认域名
		if utils.In.Array(val, []any{"oss", "cos"}) {
			if val == "oss" {
				replace["{{"+val+"}}"] = fmt.Sprintf("https://%s.%s",
					cast.ToString(toml.Get("oss.bucket")),
					cast.ToString(toml.Get("oss.endpoint")),
				)
			}
			if val == "cos" {
				replace["{{"+val+"}}"] = fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com",
					cast.ToString(toml.Get("cos.bucket")),
					cast.ToString(toml.Get("cos.app_id")),
					cast.ToString(toml.Get("cos.region")),
				)
			}
		}
	}
	if !utils.Is.Empty(facade.Cache.Get("domain")) {
		replace["{{localhost}}"] = cast.ToString(facade.Cache.Get("domain"))
	}

	return replace
}

// DomainTemp2 域名模板替换
func DomainTemp2() (replace map[string]any) {
	toml := facade.NewToml(facade.TomlStorage)
	replace = make(map[string]any)
	storage := []string{"oss", "cos", "kodo"}
	// 拼接自定义域名
	for _, val := range storage {
		if !utils.Is.Empty(toml.Get(val + ".domain")) {
			replace[cast.ToString(toml.Get(val+".domain"))] = "{{" + val + "}}"
		}
	}
	if !utils.Is.Empty(facade.Cache.Get("domain")) {
		replace[cast.ToString(facade.Cache.Get("domain"))] = "{{localhost}}"
	}

	// 拼接 oss 域名
	oss := fmt.Sprintf("https://%s.%s",
		cast.ToString(toml.Get("oss.bucket")),
		cast.ToString(toml.Get("oss.endpoint")),
	)
	// 拼接 cos 域名
	cos := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com",
		cast.ToString(toml.Get("cos.bucket")),
		cast.ToString(toml.Get("cos.app_id")),
		cast.ToString(toml.Get("cos.region")),
	)
	replace[oss] = "{{oss}}"
	replace[cos] = "{{cos}}"

	return replace
}
