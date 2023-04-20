package config

import (
	"fmt"
	"github.com/spf13/cast"
	"net/http"
)

func Run(callback ...func()) {

	// 启动服务
	for _, val := range callback {
		val()
	}

	port := ":" + cast.ToString(AppToml.Get("app.port", 8080))

	Server = &http.Server{
		Addr:    port,
		Handler: Engine,
	}

	go func() {
		if err := Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("服务启动失败", err)
		}
	}()

	// 保持主线程不退出
	select {}
}