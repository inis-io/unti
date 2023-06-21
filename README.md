### 简介

> 这是一个上手简单的Go框架，基于 [Gin](https://gin-gonic.com/zh-cn/) 二次开发，数据库基于 [Gorm](https://gorm.io/) ，设计风格参考了 [ThinkPHP 6](https://www.kancloud.cn/manual/thinkphp6_0/1037479)

### 配置文件
> 配置文件在 `config` 目录下

### 运行
> 运行前请先安装 [Go](https://golang.org/dl/) ，然后在项目根目录下执行 `go run main.go` 即可   
> 如需后台运行，可以使用 `go build -ldflags -H=windowsgui` 命令编译，[bee](https://github.com/beego/bee) 工具也可以使用 `bee pack -ba="-ldflags -H=windowsgui"` 命令打包

### 部署
> 部署前请先安装 [Go](https://golang.org/dl/) ，然后在项目根目录下执行 `go build` 即可，编译完成后会生成一个的可执行文件，将其放到服务器上即可

### 作者
> [兔子](https://inis.cn)，QQ：[97783391](https://wpa.qq.com/msgrd?v=3&uin=97783391&site=qq&menu=yes)，微信：`v-inis`