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
	"strings"
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
		Content: utils.Replace(TempLog, map[string]any{
			"${on}": "true",
			"${size}": 2,
			"${age}": 7,
			"${backups}": 20,
		}),
	}).Read()

	if item.Error != nil {
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
	LogToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	LogToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initLog()
	})
}

const (
	// LogModeInfo - 信息日志
	LogModeInfo = "info"
	// LogModeWarn - 警告日志
	LogModeWarn = "warn"
	// LogModeError - 错误日志
	LogModeError = "error"
	// LogModeDebug - 调试日志
	LogModeDebug = "debug"
)

type LogRequest struct {
	Level string
	Msg   string
}

// Write - 写入日志
func (this *LogRequest) Write(data map[string]any, msg ...any) {

	if len(msg) == 0 {
		msg = append(msg, this.Level)
	}

	this.Msg = cast.ToString(msg[0])

	switch this.Level {
	case LogModeInfo:
		Log.Info(data, this.Msg)
	case LogModeWarn:
		Log.Warn(data, this.Msg)
	case LogModeError:
		Log.Error(data, this.Msg)
	case LogModeDebug:
		Log.Debug(data, this.Msg)
	default:
		Log.Info(data, this.Msg)
	}
}

// NewLog - 创建Log实例
/**
 * @param mode 驱动模式
 * @return *zap.Logger
 * @example：
 * 1. log := facade.NewLog("info")
 * 2. log := facade.NewLog(facade.LogModeInfo)
 */
func NewLog(mode any) *LogRequest {
	item := &LogRequest{
		Msg: "log",
	}
	switch strings.ToLower(cast.ToString(mode)) {
	case LogModeInfo:
		item.Level = LogModeInfo
	case LogModeWarn:
		item.Level = LogModeWarn
	case LogModeError:
		item.Level = LogModeError
	case LogModeDebug:
		item.Level = LogModeDebug
	default:
		item.Level = LogModeInfo
	}
	return item
}

// logLevel - 创建日志通道
func logLevel(Level string) *zap.Logger {

	path := fmt.Sprintf("runtime/logs/%s/%s.log", time.Now().Format("2006-01-02"), Level)

	write := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxAge:     cast.ToInt(LogToml.Get("age")),
		MaxSize:    cast.ToInt(LogToml.Get("size")),
		MaxBackups: cast.ToInt(LogToml.Get("backups")),
	})

	// 编码器
	encoder := func() zapcore.Encoder {

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "time"
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		return zapcore.NewJSONEncoder(encoderConfig)
	}

	level := new(zapcore.Level)
	err := level.UnmarshalText([]byte(Level))
	if err != nil {
		fmt.Println("日志配置初始化错误", err)
	}
	core := zapcore.NewCore(encoder(), write, level)

	return zap.New(core)
}

// initLog - 初始化日志
func initLog() {
	LogInfo = logLevel("info")
	LogWarn = logLevel("warn")
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
type log struct {
	Info  func(data map[string]any, msg ...any)
	Warn  func(data map[string]any, msg ...any)
	Error func(data map[string]any, msg ...any)
	Debug func(data map[string]any, msg ...any)
}

// Log - 日志
var Log *log

func Info(data map[string]any, msg ...any) {

	if len(msg) == 0 {
		msg = append(msg, "info")
	}

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

		LogInfo.Info(cast.ToString(msg[0]), slice...)
	}
}
func Warn(data map[string]any, msg ...any) {

	if len(msg) == 0 {
		msg = append(msg, "warn")
	}

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

		LogWarn.Warn(cast.ToString(msg[0]), slice...)
	}
}
func Error(data map[string]any, msg ...any) {

	if len(msg) == 0 {
		msg = append(msg, "error")
	}

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

		LogError.Error(cast.ToString(msg[0]), slice...)
	}
}
func Debug(data map[string]any, msg ...any) {

	if len(msg) == 0 {
		msg = append(msg, "debug")
	}

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

		LogDebug.Debug(cast.ToString(msg[0]), slice...)
	}
}
