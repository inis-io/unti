### 简介

> 这是一个上手简单的Go框架，基于 [Gin](https://gin-gonic.com/zh-cn/) 二次开发，数据库基于 [Gorm](https://gorm.io/) ，设计风格参考了 [ThinkPHP 6](https://www.kancloud.cn/manual/thinkphp6_0/1037479)

### 配置文件
> 配置文件在 `config` 目录下

### 运行
> 运行前请先安装 [Go](https://golang.org/dl/) ，然后在项目根目录下执行 `go run main.go` 即可   
> 如需后台运行，可以使用 `go build -ldflags -H=windowsgui` 命令编译，[bee](https://github.com/beego/bee) 工具也可以使用 `bee pack -ba="-ldflags -H=windowsgui"` 命令打包

### 部署
> 部署前请先安装 [Go](https://golang.org/dl/) ，然后在项目根目录下执行 `go build` 即可，编译完成后会生成一个的可执行文件，将其放到服务器上即可

### 门面模式 facade
1. Cache：Redis缓存、文件缓存、内存缓存
2. DB：MySQL数据库驱动
3. i18n：国际化多语言模块
4. Log：多日志通道（info、warn、error、debug），日志按日期和大小分包
5. SMS：阿里云短信、腾讯云短信、邮件推送
6. Storage：本地存储、阿里云OSS存储、腾讯云COS存储、七牛云KODO存储

### 多应用
1. API 资源路由（遵循RESTful API规范）

### 其他
1. 数据库模型
2. 参数校验器
3. utils工具包
4. public静态资源目录
5. 图片缩略图

### 目录结构
```
├─app                应用目录
│  ├─api             API应用
│  │  ├─controller   控制器
│  │  ├─middleware   局部中间件
│  │  └─route        路由
│  ├─dev             开发应用
│  │  ├─controller   控制器
│  │  └─route        路由
│  ├─facade          门面模式
│  ├─index           前台应用
│  │  ├─controller   控制器
│  │  └─route        路由
│  ├─middleware      全局中间件
│  ├─model           数据库模型
│  ├─socket          Socket应用
│  │  ├─controller
│  │  ├─middleware
│  │  └─route
│  └─validator       参数校验器
├─config             配置文件
│  └─i18n            多语言
├─public             静态资源（对外，其他目录均受保护）
│  ├─assets
│  │  └─images 
│  │      ├─avatar
│  │      └─gif
│  └─storage
│      └─rand
│          └─avatar
└─runtime            运行时目录
    ├─cache          缓存
    └─logs           日志
        └─2023-06-21 分片日志
```

### 项目地址
GitHub: [Unti：github.com/unti-io/unti](https://github.com/unti-io/unti)

### 联系方式
QQ：97783391
微信：v-inis
QQ群：563867338