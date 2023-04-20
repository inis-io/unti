package facade

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sort"
	"time"
)

// LogToml - 日志配置文件
var LogToml *utils.ViperResponse

// initLogToml - 初始化缓存配置文件
func initLogToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "log",
		Content: `# ======== 日志配置 ========

# 是否启用日志
on		   = true
# 日志存储路径
path       = "runtime/"
# 单个日志文件大小（MB）
size	   = 2
# 日志文件保存天数
age		   = 7
# 日志文件最大保存数量
backups	   = 20
`,
	}).Read()

	if item.Error != nil {
		// Log().Error("日志配置初始化错误", map[string]any{
		// 	"error": item.Error,
		// 	"func_name": utils.Caller().FuncName,
		// 	"file_name": utils.Caller().FileName,
		// 	"file_line": utils.Caller().Line,
		// })
		fmt.Println("日志配置初始化错误", item.Error)
		return
	}

	LogToml = &item
}

func init() {
	// 初始化配置文件
	initLogToml()
	// 初始化缓存
	initLog()

	// 监听配置文件变化
	CacheToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	CacheToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initLog()
	})
}

// logLevel - 创建日志通道
func logLevel(Level string) *zap.Logger {

	group := "logs"
	path  := fmt.Sprintf("%s%s/%s/%s.log", LogToml.Get("path"), group, time.Now().Format("2006-01-02"), Level)

	write := zapcore.AddSync(&lumberjack.Logger{
		Filename  : path,
		MaxAge    : cast.ToInt(LogToml.Get("age")),
		MaxSize   : cast.ToInt(LogToml.Get("size")),
		MaxBackups: cast.ToInt(LogToml.Get("backups")),
	})

	// 编码器
	encoder := func() zapcore.Encoder {

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey        = "time"
		encoderConfig.EncodeTime	 = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		encoderConfig.EncodeCaller   = zapcore.ShortCallerEncoder

		return zapcore.NewJSONEncoder(encoderConfig)
	}

	level := new(zapcore.Level)
	err   := level.UnmarshalText([]byte(Level))
	if err != nil {
		fmt.Println("日志配置初始化错误", err)
	}
	core := zapcore.NewCore(encoder(), write, level)

	return zap.New(core)
}

// initLog - 初始化日志
func initLog() {
	LogInfo  = logLevel("info")
	LogWarn  = logLevel("warn")
	LogError = logLevel("error")
	LogDebug = logLevel("debug")

	Log = &log{
		Info:  Info,
		Warn:  Warn,
		Error: Error,
		Debug: Debug,
	}
}

// LogInfo - info日志通道
var LogInfo *zap.Logger
// LogWarn - warn日志通道
var LogWarn *zap.Logger
// LogError - error日志通道
var LogError *zap.Logger
// LogDebug - debug日志通道
var LogDebug *zap.Logger

// log - 日志结构体
type log struct{
	Info  func(msg string, data map[string]any)
	Warn  func(msg string, data map[string]any)
	Error func(msg string, data map[string]any)
	Debug func(msg string, data map[string]any)
}

// Log - 日志
var Log *log

func Info(msg string, data map[string]any) {


	if len(data) > 0 {

		// ========== 此处解决 map 无序问题 - 开始 ==========
		keys := make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
		// 排序 keys
		sort.Strings(keys)
		// ========== 此处解决 map 无序问题 - 开始 ==========

		slice := make([]zap.Field, 0)

		for key := range keys {
			slice = append(slice, zap.Any(keys[key], data[keys[key]]))
		}

		LogInfo.Info(msg, slice...)
	}
}
func Warn(msg string, data map[string]any) {

	if len(data) > 0 {

		// ========== 此处解决 map 无序问题 - 开始 ==========
		keys := make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
		// 排序 keys
		sort.Strings(keys)
		// ========== 此处解决 map 无序问题 - 开始 ==========

		slice := make([]zap.Field, 0)

		for key := range keys {
			slice = append(slice, zap.Any(keys[key], data[keys[key]]))
		}

		LogWarn.Warn(msg, slice...)
	}
}
func Error(msg string, data map[string]any) {

	if len(data) > 0 {

		// ========== 此处解决 map 无序问题 - 开始 ==========
		keys := make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
		// 排序 keys
		sort.Strings(keys)
		// ========== 此处解决 map 无序问题 - 开始 ==========

		slice := make([]zap.Field, 0)

		for key := range keys {
			slice = append(slice, zap.Any(keys[key], data[keys[key]]))
		}

		LogError.Error(msg, slice...)
	}
}
func Debug(msg string, data map[string]any) {

	if len(data) > 0 {

		// ========== 此处解决 map 无序问题 - 开始 ==========
		keys := make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
		// 排序 keys
		sort.Strings(keys)
		// ========== 此处解决 map 无序问题 - 开始 ==========

		slice := make([]zap.Field, 0)

		for key := range keys {
			slice = append(slice, zap.Any(keys[key], data[keys[key]]))
		}

		LogDebug.Debug(msg, slice...)
	}
}