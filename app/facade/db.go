package facade

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"gorm.io/gorm"
	"strings"
)

const (
	// DBModeMySql - MySQL数据库
	DBModeMySql = "mysql"
)

// NewDB - 创建DB实例
/**
 * @param mode 驱动模式
 * @return DBInterface
 * @example：
 * 1. db := facade.NewDB("mysql")
 * 2. db := facade.NewDB(facade.DBModeMySql)
 */
func NewDB(mode any) DBInterface {
	switch strings.ToLower(cast.ToString(mode)) {
	case DBModeMySql:
		DB = MySQL
	default:
		DB = MySQL
	}
	return DB
}

// DB - DB实例
/**
 * @return DBInterface
 * @example：
 * db := facade.DB.Model(&model.Users{}).Find()
 */
var DB DBInterface

type DBInterface interface {
	// Model 模型
	Model(model any) *ModelStruct
	// Drive - 获取数据库连接
	Drive() *gorm.DB
}

type ModelStruct struct {
	dest              any      // 目标表结构体
	model             *gorm.DB // 模型
	order             any      // 排序
	limit             any      // 限制
	page              any      // 分页
	softDelete        string   // 软删除 - 字段
	defaultSoftDelete any      // 默认软删除 - 值
	field 		      []string // 查询字段范围
	withoutField	  []string // 排除查询字段
}

type ModelInterface interface {
	// Debug - 是否开启调试模式
	Debug(yes ...any) *ModelStruct
	// Where - 排序
	Where(args ...any) *ModelStruct
	// IWhere - 断言条件
	IWhere(where any) *ModelStruct
	// WhereIn - IN查询
	WhereIn(args ...any) *ModelStruct
	// IWhereIn - 断言条件
	IWhereIn(where any) *ModelStruct
	// Not - 条件
	Not(args ...any) *ModelStruct
	// INot - 断言条件
	INot(where any) *ModelStruct
	// Or - 条件
	Or(args ...any) *ModelStruct
	// IOr - 断言条件
	IOr(where any) *ModelStruct
	// Like - 条件
	Like(args ...any) *ModelStruct
	// ILike - 断言条件
	ILike(where any) *ModelStruct
	// Null - 条件
	Null(args ...any) *ModelStruct
	// INull - 断言条件
	INull(where any) *ModelStruct
	// NotNull - 条件
	NotNull(args ...any) *ModelStruct
	// INotNull - 断言条件
	INotNull(where any) *ModelStruct
	// WithTrashed - 软删除 - 包含软删除
	WithTrashed(yes ...any) *ModelStruct
	// OnlyTrashed - 软删除 - 只包含软删除
	OnlyTrashed(yes ...any) *ModelStruct
	// Order - 排序
	Order(args ...any) *ModelStruct
	// Limit - 限制
	Limit(args ...any) *ModelStruct
	// Page - 分页
	Page(args ...any) *ModelStruct
	// Field - 查询字段范围
	Field(args ...any) *ModelStruct
	// WithoutField - 排除查询字段
	WithoutField(args ...any) *ModelStruct
	// Select - 查询多条
	Select(args ...any) (result []map[string]any)
	// Find - 查询单条
	Find(args ...any) (result map[string]any)
	// Exist - 是否存在
	Exist(args ...any) (ok bool)
	// FindOrEmpty - 是否不存在
	FindOrEmpty(args ...any) (ok bool)
	// Count - 统计
	Count() (result int64)
	// Column - 列
	Column(args ...any) (result any)
	// Sum - 求和
	Sum(field string) (result int64)
	// Max - 最大值
	Max(field string) (result int64)
	// Min - 最小值
	Min(field string) (result int64)
	// Update - 更新
	Update(data ...any) (tx *gorm.DB)
	// Force - 真实删除
	Force() *ModelStruct
	// Delete - 删除
	Delete(args ...any) (tx *gorm.DB)
	// Destroy - 销毁
	Destroy(args ...any) (tx *gorm.DB)
	// Restore - 恢复
	Restore(args ...any) (tx *gorm.DB)
	// Create - 创建
	Create(data ...any) (tx *gorm.DB)
	// Save - 保存
	Save(data ...any) (tx *gorm.DB)
	// Inc - 自增
	Inc(column any, step ...int) *ModelStruct
	// Dec - 自减
	Dec(column any, step ...int) *ModelStruct
	// UpdateColumn - 更新单个字段
	UpdateColumn(column any, value any) (tx *gorm.DB)
}

// WatchDB - 初始化数据库 - 顺便监听配置文件变化
func WatchDB(change ...bool) {

	if len(change) == 0 {
		change = append(change, false)
	}

	// 初始化配置文件
	initDBToml()
	// 初始化数据库
	InitDB()

	if change[0] {
		// 监听配置文件变化
		DBToml.Viper.WatchConfig()
		// 配置文件变化时，重新初始化配置文件
		DBToml.Viper.OnConfigChange(func(event fsnotify.Event) {
			InitDB()
		})
	}
}

// DBToml - 数据库配置文件
var DBToml *utils.ViperResponse

// initDBToml - 初始化数据库配置文件
func initDBToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "database",
		Content: utils.Replace(TempDatabase, map[string]any{
			"${mysql.hostname}": "localhost",
			"${mysql.hostport}": 3306,
			"${mysql.username}": "",
			"${mysql.database}": "",
			"${mysql.password}": "",
			"${mysql.charset}" : "utf8mb4",
			"${mysql.migrate}" : "true",
		}),
	}).Read()

	if item.Error != nil {
		Log.Error(map[string]any{
			"error":     item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "数据库配置初始化错误")
		return
	}

	DBToml = &item
}

// InitDB - 初始化数据库
func InitDB() {

	InitMySQL()

	switch cast.ToString(DBToml.Get("default")) {
	case "mysql":
		DB = MySQL
	default:
		DB = MySQL
	}
}
