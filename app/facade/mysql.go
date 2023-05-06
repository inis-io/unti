package facade

import (
	"fmt"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"time"
)

var MySQL *MySqlStruct

type MySqlStruct struct {
	// DB 数据库实例
	Conn *gorm.DB
}

// 初始化数据库
func initMySql() {

	hostname := cast.ToString(DBToml.Get("mysql.hostname", "localhost"))
	hostport := cast.ToString(DBToml.Get("mysql.hostport", "3306"))
	username := cast.ToString(DBToml.Get("mysql.username", ""))
	database := cast.ToString(DBToml.Get("mysql.database", ""))
	password := cast.ToString(DBToml.Get("mysql.password", ""))
	charset := cast.ToString(DBToml.Get("mysql.charset", "utf8mb4"))
	prefix := cast.ToString(DBToml.Get("mysql.prefix", "unti_"))

	conn, _ := gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", username, password, hostname, hostport, database, charset),
		// string 类型字段的默认长度
		DefaultStringSize: 256,
		// 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DisableDatetimePrecision: true,
		// 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameIndex: true,
		// 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		DontSupportRenameColumn: true,
		// 根据当前 MySQL 版本自动配置
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			// 表名前缀，`User` 的表名应该是 `t_users`
			TablePrefix: prefix,
			// 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
			SingularTable: true,
		},
		// 关闭终端显示查询信息
		Logger: logger.Default.LogMode(logger.Silent),
	})

	sqlDB, _ := conn.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)

	MySQL = &MySqlStruct{
		Conn: conn,
	}
}

func (this *MySqlStruct) Drive() *gorm.DB {
	return this.Conn
}

func (this *MySqlStruct) Model(model any) *ModelStruct {
	return &ModelStruct{
		dest:              model,
		model:             this.Conn.Model(model),
		softDelete:        "delete_time",
		defaultSoftDelete: 0,
	}
}

func (this *ModelStruct) Dest(dest any) *ModelStruct {
	this.dest = dest
	return this
}

func (this *ModelStruct) Scan(dest any) *ModelStruct {
	this.model.Scan(dest)
	return this
}

func (this *ModelStruct) Debug(yes ...any) *ModelStruct {

	if len(yes) == 0 {
		yes = append(yes, true)
	}

	if cast.ToBool(yes[0]) {
		this.model = this.model.Debug()
	}

	return this
}

// Where - 条件
func (this *ModelStruct) Where(args ...any) *ModelStruct {

	// fmt.Println("where", args, len(args))

	if len(args) >= 3 {

		query := fmt.Sprintf("%v %v ?", args[0], args[1])
		this.model.Where(query, args[2])

	} else if len(args) == 2 {

		query := fmt.Sprintf("%v = ?", args[0])
		this.model.Where(query, args[1])

	} else if len(args) == 1 {

		// 判断是否为数组
		if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
			// 情况二：二维数组
			if reflect.TypeOf(args[0].([]any)[0]).Kind() == reflect.Slice {
				for _, val := range args[0].([]any) {
					this.Where(val)
				}
			} else {
				// 情况一：一维数组
				this.Where(args[0].([]any)...)
			}

		} else {

			// 情况三：字符串 - 必须空格分隔且长度为3 - 否则不处理
			if reflect.TypeOf(args[0]).Kind() == reflect.String {
				str := strings.Split(args[0].(string), " ")
				if len(str) == 3 {
					query := fmt.Sprintf("%v %v ?", str[0], str[1])
					this.model.Where(query, str[2])
				}
			} else {
				this.model.Where(args[0])
			}
		}

	}
	// else {
	//
	// 	// 情况三：字符串 - 必须空格分隔且长度为3 - 否则不处理
	// 	if reflect.TypeOf(args[0]).Kind() == reflect.String {
	// 		str := strings.Split(args[0].(string), " ")
	// 		if len(str) == 3 {
	// 			query := fmt.Sprintf("%v %v ?", str[0], str[1])
	// 			this.model.Where(query, str[2])
	// 		}
	// 	}
	// }

	return this
}

// IWhere - 断言条件
func (this *ModelStruct) IWhere(where any) *ModelStruct {

	if utils.Is.Empty(where) {
		return this
	}

	if utils.IsString(where) {

		this.Where(where)

	} else if utils.IsSlice(where) {

		var slice []any
		// ToStringSlice
		for _, val := range cast.ToSlice(where) {
			// this.Where(val)
			slice = append(slice, val)
		}

		this.Where(slice)

	} else if utils.IsMapAny(where) {

		for _, val := range cast.ToStringMap(where) {
			this.Where(val)
		}
	}

	return this
}

// Not - 条件
func (this *ModelStruct) Not(args ...any) *ModelStruct {

	if len(args) >= 3 {

		query := fmt.Sprintf("%v %v ?", args[0], args[1])
		this.model.Not(query, args[2])

	} else if len(args) == 2 {

		query := fmt.Sprintf("%v = ?", args[0])
		this.model.Not(query, args[1])

	} else if len(args) == 1 {

		// 判断是否为数组
		if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
			// 情况二：二维数组
			if reflect.TypeOf(args[0].([]any)[0]).Kind() == reflect.Slice {
				for _, val := range args[0].([]any) {
					this.Not(val)
				}
			} else {
				// 情况一：一维数组
				this.Not(args[0].([]any)...)
			}

		} else {

			// 情况三：字符串 - 必须空格分隔且长度为3 - 否则不处理
			if reflect.TypeOf(args[0]).Kind() == reflect.String {
				str := strings.Split(args[0].(string), " ")
				if len(str) == 3 {
					query := fmt.Sprintf("%v %v ?", str[0], str[1])
					this.model.Not(query, str[2])
				}
			}
		}
	} else {

		// 情况三：字符串 - 必须空格分隔且长度为3 - 否则不处理
		if reflect.TypeOf(args[0]).Kind() == reflect.String {
			str := strings.Split(args[0].(string), " ")
			if len(str) == 3 {
				query := fmt.Sprintf("%v %v ?", str[0], str[1])
				this.model.Not(query, str[2])
			}
		}
	}

	return this
}

// INot - 断言条件
func (this *ModelStruct) INot(where any) *ModelStruct {

	if utils.Is.Empty(where) {
		return this
	}

	if utils.IsString(where) {

		this.Not(where)

	} else if utils.IsSlice(where) {

		var slice []any
		for _, val := range cast.ToStringSlice(where) {
			slice = append(slice, val)
		}
		this.Not(slice)

	} else if utils.IsMapAny(where) {

		for _, val := range cast.ToStringMap(where) {
			this.Not(val)
		}
	}

	return this
}

// Or - 条件
func (this *ModelStruct) Or(args ...any) *ModelStruct {

	if len(args) >= 3 {

		query := fmt.Sprintf("%v %v ?", args[0], args[1])
		this.model.Or(query, args[2])

	} else if len(args) == 2 {

		query := fmt.Sprintf("%v = ?", args[0])
		this.model.Or(query, args[1])

	} else if len(args) == 1 {

		// 判断是否为数组
		if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
			// 情况二：二维数组
			if reflect.TypeOf(args[0].([]any)[0]).Kind() == reflect.Slice {
				for _, val := range args[0].([]any) {
					this.Or(val)
				}
			} else {
				// 情况一：一维数组
				this.Or(args[0].([]any)...)
			}

		} else {

			// 情况三：字符串 - 必须空格分隔且长度为3 - 否则不处理
			if reflect.TypeOf(args[0]).Kind() == reflect.String {
				str := strings.Split(args[0].(string), " ")
				if len(str) == 3 {
					query := fmt.Sprintf("%v %v ?", str[0], str[1])
					this.model.Or(query, str[2])
				}
			}
		}
	} else {

		// 情况三：字符串 - 必须空格分隔且长度为3 - 否则不处理
		if reflect.TypeOf(args[0]).Kind() == reflect.String {
			str := strings.Split(args[0].(string), " ")
			if len(str) == 3 {
				query := fmt.Sprintf("%v %v ?", str[0], str[1])
				this.model.Or(query, str[2])
			}
		}
	}

	return this
}

// IOr - 断言条件
func (this *ModelStruct) IOr(where any) *ModelStruct {

	if utils.Is.Empty(where) {
		return this
	}

	if utils.IsString(where) {

		this.Or(where)

	} else if utils.IsSlice(where) {

		var slice []any
		for _, val := range cast.ToStringSlice(where) {
			slice = append(slice, val)
		}
		this.Or(slice)

	} else if utils.IsMapAny(where) {

		for _, val := range cast.ToStringMap(where) {
			this.Or(val)
		}
	}

	return this
}

// Like - 条件
func (this *ModelStruct) Like(args ...any) *ModelStruct {

	if len(args) >= 2 {

		query := fmt.Sprintf("%v LIKE ?", args[0])
		this.model.Where(query, args[1])

	} else if len(args) == 1 {

		// 判断是否为数组
		if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
			// 情况二：二维数组
			if reflect.TypeOf(args[0].([]any)[0]).Kind() == reflect.Slice {
				for _, val := range cast.ToSlice(args[0]) {
					this.Like(val)
				}
			} else {
				// 情况一：一维数组
				this.Like(cast.ToSlice(args[0])...)
			}
		} else {

			// 情况三：字符串 - 必须空格分隔且长度为3 - 否则不处理
			if reflect.TypeOf(args[0]).Kind() == reflect.String {
				str := strings.Split(args[0].(string), " ")
				if len(str) == 2 {
					query := fmt.Sprintf("%v LIKE ?", str[0])
					this.model.Where(query, str[1])
				}
			}
		}
	}

	return this
}

// ILike - 断言条件
func (this *ModelStruct) ILike(where any) *ModelStruct {

	if utils.Is.Empty(where) {
		return this
	}

	if utils.IsString(where) {

		this.Like(where)

	} else if utils.IsSlice(where) {

		var slice []any
		for _, val := range cast.ToStringSlice(where) {
			slice = append(slice, val)
		}
		this.Like(slice)

	} else if utils.IsMapAny(where) {

		var sql string
		for _, val := range cast.ToStringMap(where) {
			item := cast.ToSlice(val)
			sql += fmt.Sprintf("%v LIKE '%v' OR ", item[0], item[1])
		}
		this.model.Where(strings.TrimRight(sql, "OR "))
	}

	return this
}

// Null - 条件
func (this *ModelStruct) Null(args ...any) *ModelStruct {

	for _, val := range args {

		if reflect.TypeOf(val).Kind() == reflect.String {

			// 判断是否逗号分隔
			if strings.Contains(val.(string), ",") {
				// 逗号分割 去除空格
				for _, v := range strings.Split(val.(string), ",") {
					query := fmt.Sprintf("%v IS NULL", strings.TrimSpace(v))
					this.model.Where(query)
				}
			} else {
				query := fmt.Sprintf("%v IS NULL", val)
				this.model.Where(query)
			}

		} else if reflect.TypeOf(val).Kind() == reflect.Slice {

			this.Null(val.([]any)...)
		}
	}

	return this
}

// INull - 断言条件
func (this *ModelStruct) INull(where any) *ModelStruct {

	if utils.Is.Empty(where) {
		return this
	}

	if utils.IsString(where) {

		this.Null(where)

	} else if utils.IsSlice(where) {

		var slice []any
		for _, val := range cast.ToStringSlice(where) {
			slice = append(slice, val)
		}
		this.Null(slice)

	} else if utils.IsMapAny(where) {

		for _, val := range cast.ToStringMap(where) {
			this.Null(val)
		}
	}

	return this
}

// NotNull - 条件
func (this *ModelStruct) NotNull(args ...any) *ModelStruct {

	for _, val := range args {

		if reflect.TypeOf(val).Kind() == reflect.String {

			// 判断是否逗号分隔
			if strings.Contains(val.(string), ",") {
				// 逗号分割 去除空格
				for _, v := range strings.Split(val.(string), ",") {
					query := fmt.Sprintf("%v IS NOT NULL", strings.TrimSpace(v))
					this.model.Where(query)
				}
			} else {
				query := fmt.Sprintf("%v IS NOT NULL", val)
				this.model.Where(query)
			}

		} else if reflect.TypeOf(val).Kind() == reflect.Slice {

			this.NotNull(val.([]any)...)
		}
	}

	return this
}

// INotNull - 断言条件
func (this *ModelStruct) INotNull(where any) *ModelStruct {

	if utils.Is.Empty(where) {
		return this
	}

	if utils.IsString(where) {

		this.NotNull(where)

	} else if utils.IsSlice(where) {

		var slice []any
		for _, val := range cast.ToStringSlice(where) {
			slice = append(slice, val)
		}
		this.NotNull(slice)

	} else if utils.IsMapAny(where) {

		for _, val := range cast.ToStringMap(where) {
			this.NotNull(val)
		}
	}

	return this
}

// WithTrashed - 软删除 - 包含软删除
func (this *ModelStruct) WithTrashed(yes ...any) *ModelStruct {

	if len(yes) == 0 {
		yes = append(yes, true)
	}

	if cast.ToBool(yes[0]) {
		this.model.Unscoped()
	}

	return this
}

// OnlyTrashed - 软删除 - 只包含软删除
func (this *ModelStruct) OnlyTrashed(yes ...any) *ModelStruct {

	if len(yes) == 0 {
		yes = append(yes, true)
	}

	if cast.ToBool(yes[0]) {
		this.model.Unscoped().Where(fmt.Sprintf("%v <> ?", this.softDelete), this.defaultSoftDelete)
	}

	return this
}

// Order - 排序
func (this *ModelStruct) Order(args ...any) *ModelStruct {
	if len(args) > 0 {
		if utils.Is.Empty(args[0]) {
			return this
		}
		this.order = args[0]
		this.model.Order(args[0])
	}
	return this
}

// Limit - 限制
func (this *ModelStruct) Limit(limit ...any) *ModelStruct {
	if len(limit) > 0 {
		if utils.Is.Empty(limit[0]) {
			return this
		}
		this.limit = limit[0]
		this.model.Limit(limit[0].(int))
	}
	return this
}

// Page - 分页
func (this *ModelStruct) Page(page ...any) *ModelStruct {
	if len(page) > 0 {
		if this.limit == nil {
			this.limit = 1
			this.model.Limit(1)
		}
		this.page = page[0]
		this.model.Offset((page[0].(int) - 1) * this.limit.(int))
	}
	return this
}

// Field - 查询字段范围
func (this *ModelStruct) Field(args ...any) *ModelStruct {
	if len(args) > 0 {
		this.model.Select(args[0], args[1:]...)
	}
	return this
}

// WithoutField - 排除查询字段
func (this *ModelStruct) WithoutField(args ...any) *ModelStruct {
	if len(args) > 0 {
		for _, val := range args {
			this.model.Omit(val.(string))
		}
	}
	return this
}

// Select - 查询多条
func (this *ModelStruct) Select(args ...any) (result []map[string]any) {

	if len(args) > 0 {
		// 根据主键查询
		if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
			// 根据 id 批量查询
			this.model.Where(args[0]).Find(this.dest)
		} else {
			// 根据 id 单个查询
			this.model.Where("id = ?", args[0]).Find(this.dest)
		}
	} else {
		// 查询全部
		this.model.Find(this.dest)
	}

	// any to []map[string]any
	json := utils.Json.Decode(utils.Json.Encode(this.dest))
	for _, val := range cast.ToSlice(json) {
		result = append(result, cast.ToStringMap(val))
	}

	return
}

// Find - 查询单条
func (this *ModelStruct) Find(args ...any) (result map[string]any) {

	if len(args) > 0 {
		// 根据ID查询
		this.model.Where("id = ?", args[0])
	}

	tx := this.model.First(&this.dest)

	if tx.Error != nil {
		return nil
	}

	json := utils.Json.Decode(utils.Json.Encode(this.dest))
	result = cast.ToStringMap(json)

	if utils.Is.Empty(result) {
		return nil
	}

	return result
}

// FindOrEmpty - 查询单条
func (this *ModelStruct) FindOrEmpty(args ...any) (ok bool) {

	if len(args) > 0 {
		// 根据ID查询
		this.model.Where("id = ?", args[0])
	}

	tx := this.model.First(&this.dest)

	if tx.Error != nil {
		return false
	}

	return utils.Ternary[bool](tx.RowsAffected > 0, true, false)
}

// Count - 统计
func (this *ModelStruct) Count() (result int64) {

	var count int64

	this.model.Count(&count)

	return count
}

// Column - 列
func (this *ModelStruct) Column(args ...any) (result any) {

	if len(args) > 0 {
		this.model.Select(args[0], args[1:]...)
	}

	if len(args) == 1 {

		var data []string
		this.model.Pluck(cast.ToString(args[0]), &data)

		return data

	} else {

		var data []map[string]any
		this.model.Scan(&data)

		return data
	}
}

// Sum - 求和
func (this *ModelStruct) Sum(field string) (result int64) {

	var sum int64
	this.model.Select("sum(" + field + ") as sum").Scan(&sum)

	return sum
}

// Max - 最大值
func (this *ModelStruct) Max(field string) (result int64) {

	var max int64
	this.model.Select("max(" + field + ") as max").Scan(&max)

	return max
}

// Min - 最小值
func (this *ModelStruct) Min(field string) (result int64) {

	var min int64
	this.model.Select("min(" + field + ") as min").Scan(&min)

	return min
}

// Create - 创建
func (this *ModelStruct) Create(data any) (tx *gorm.DB) {
	return this.model.Create(data)
}

// Update - 更新
func (this *ModelStruct) Update(data any) (tx *gorm.DB) {
	// 返回主键
	return this.model.Updates(data)
}

// Save - 保存
func (this *ModelStruct) Save(data any) (tx *gorm.DB) {

	// 查询是否存在 - 存在则更新，不存在则创建
	tx = this.model.First(&this.dest)
	if tx.Error != nil {
		return NewDB(DBModeMySql).Model(this.dest).Create(data)
	}

	// 更新
	return this.model.Updates(data)
}

// Force - 真实删除
func (this *ModelStruct) Force() *ModelStruct {
	this.model.Unscoped()
	return this
}

// Delete - 删除
func (this *ModelStruct) Delete(args ...any) (ok bool) {

	if len(args) > 0 {

		// 根据主键删除
		if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
			// 根据 id 批量删除
			this.model.Delete(nil, args[0])
		} else {
			// 根据 id 单个删除
			this.model.Where("id = ?", args[0]).Delete(nil)
		}

	} else {

		// 普通删除
		this.model.Delete(nil)
	}

	return true
}

// Destroy - 销毁
func (this *ModelStruct) Destroy(args ...any) (ok bool) {

	// 如果 args 的长度小于 2，扩容
	if len(args) < 2 {
		args = append(args, make([]any, 2-len(args))...)
	}

	// 如果 args[1] != true，设置为 false
	if args[1] != true {
		args[1] = false
	}

	if args[1] == true {
		// 真实删除
		this.model.Unscoped()
	}

	if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
		// 根据 id 批量删除
		this.model.Delete(nil, args[0])
	} else {
		// 根据 id 单个删除
		this.model.Where("id = ?", args[0]).Delete(nil)
	}

	return true
}

// Restore - 恢复
func (this *ModelStruct) Restore(args ...any) (ok bool) {

	if len(args) > 0 {
		// 根据主键查询
		if reflect.TypeOf(args[0]).Kind() == reflect.Slice {
			// 根据 id 批量查询
			this.model.Where(args[0])
		} else {
			// 根据 id 单个查询
			this.model.Where("id = ?", args[0])
		}
	}

	// 恢复
	tx := this.model.Unscoped().Update(this.softDelete, this.defaultSoftDelete)

	if tx.Error != nil {
		return false
	}

	return true
}
