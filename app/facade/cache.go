package facade

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/fsnotify/fsnotify"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"reflect"
	debugs "runtime/debug"
	"strings"
	"time"
)

func init() {

	// 初始化配置文件
	initCacheToml()
	// 初始化缓存
	initCache()

	// 监听配置文件变化
	CacheToml.Viper.WatchConfig()
	// 配置文件变化时，重新初始化配置文件
	CacheToml.Viper.OnConfigChange(func(event fsnotify.Event) {
		initCache()
	})
}

const (
	// CacheModeLocal - 本地缓存
	CacheModeLocal = "local"
	// CacheModeRedis - Redis缓存
	CacheModeRedis = "redis"
)

// NewCache - 创建Cache实例
/**
 * @param mode 驱动模式
 * @return CacheInterface
 * 使用方式：
 * 1. cache := facade.NewCache("redis")
 * 2. cache := facade.NewCache(facade.CacheModeRedis)
 */
func NewCache(mode any) CacheInterface {
	switch strings.ToLower(cast.ToString(mode)) {
	case "local":
		Cache = BigCache
	case "redis":
		Cache = Redis
	default:
		Cache = BigCache
	}
	return Cache
}

// CacheToml - 缓存配置文件
var CacheToml *utils.ViperResponse

// initCacheToml - 初始化缓存配置文件
func initCacheToml() {
	item := utils.Viper(utils.ViperModel{
		Path: "config",
		Mode: "toml",
		Name: "cache",
		Content: `# ======== 缓存配置 ========

# 是否开启API缓存
api		   = true
# 默认缓存驱动
default    = "local"

# 本地缓存配置
[local]
# 过期时间(秒) - 本地缓存不建议缓存过长时间
expire     = 300

# redis配置
[redis]
# redis地址
host       = "localhost"
# redis端口
port       = 6379
# redis密码
password   = ""
# 过期时间(秒) - 0为永不过期
expire     = 7200
# redis前缀
prefix     = "unti:"
# redis数据库
database   = 0
`,
	}).Read()

	if item.Error != nil {
		Log.Error("缓存配置初始化错误", map[string]any{
			"error": item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		})
		return
	}

	CacheToml = &item
}

// 初始化缓存
func initCache() {

	host := cast.ToString(CacheToml.Get("redis.host"))
	port := cast.ToString(CacheToml.Get("redis.port"))

	redisPrefix := cast.ToString(CacheToml.Get("redis.prefix"))
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		DB:       cast.ToInt(CacheToml.Get("redis.database")),
		Password: cast.ToString(CacheToml.Get("redis.password")),
	})
	redisExpire := time.Duration(cast.ToInt(CacheToml.Get("redis.expire", 7200))) * time.Second

	// Redis 缓存
	Redis = &RedisCacheStruct{
		Client: redisClient,
		Prefix: redisPrefix,
		Expire: redisExpire,
	}

	localExpire := time.Duration(cast.ToInt(CacheToml.Get("local.expire", 7200))) * time.Second
	localClient, err := bigcache.New(context.Background(), bigcache.DefaultConfig(localExpire))
	if err != nil {
		Log.Error("本地缓存初始化失败", map[string]any{
			"error": err,
			"stack": string(debugs.Stack()),
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		})
	}

	// 本地缓存
	BigCache = &LocalCacheStruct{
		Client: localClient,
		Expire: localExpire,
	}

	switch cast.ToString(CacheToml.Get("default")) {
	case "redis":
		Cache = Redis
	case "local":
		Cache = BigCache
	default:
		Cache = BigCache
	}
}

// Cache - Cache实例
/**
 * @return CacheInterface
 * @example：
 * cache := facade.Cache.Set("test", "这是测试", 5 * time.Minute)
 */
var Cache CacheInterface
var Redis *RedisCacheStruct
var BigCache *LocalCacheStruct

type CacheInterface interface {
	// Has
	/**
	 * @name 判断缓存是否存在
	 * @param key 缓存的key
	 * @return bool
	 */
	Has(key any) (ok bool)
	// Get
	/**
	 * @name 获取缓存
	 * @param key 缓存的key
	 * @return any 缓存值
	 */
	Get(key any) (value any)
	// Set
	/**
	 * @name 设置缓存
	 * @param key 缓存的key
	 * @param value 缓存的值
	 * @param expire （可选）过期时间
	 * @return bool
	 */
	Set(key any, value any, expire ...any) (ok bool)
	// Del
	/**
	 * @name 删除缓存
	 * @param key 缓存的key
	 * @return bool
	 */
	Del(key any) (ok bool)
}

type RedisCacheStruct struct {
	Client    *redis.Client
	Prefix    string
	Expire    time.Duration
}

func (this *RedisCacheStruct) Has(key any) (ok bool) {

	ctx := context.Background()

	result, err := this.Client.Exists(ctx, this.Prefix+cast.ToString(key)).Result()
	return utils.Ternary[bool](err != nil, false, result == 1)
}

func (this *RedisCacheStruct) Get(key any) (value any) {

	ctx := context.Background()

	result, err := this.Client.Get(ctx, this.Prefix+cast.ToString(key)).Result()

	return utils.Ternary[any](err != nil, nil, utils.Json.Decode(result))
}

func (this *RedisCacheStruct) Set(key any, value any, expire ...any) bool {

	ctx := context.Background()
	// 设置过期时间
	if len(expire) == 0 {
		expire = append(expire, cast.ToInt(CacheToml.Get("redis.expire", this.Expire)))
	}

	// 如果 exp不为时间类型，则转码为时间类型
	if reflect.ValueOf(expire[0]).Kind() != reflect.Int64 && expire[0] != 0 {
		expire[0] = time.Duration(cast.ToInt(expire[0])) * time.Second
	}

	err := this.Client.Set(ctx, this.Prefix + cast.ToString(key), utils.Json.Encode(value), cast.ToDuration(expire[0])).Err()
	return utils.Ternary[bool](err != nil, false, true)
}

func (this *RedisCacheStruct) Del(key any) bool {

	ctx := context.Background()
	err := this.Client.Del(ctx, this.Prefix+cast.ToString(key)).Err()
	return utils.Ternary[bool](err != nil, false, true)
}

type LocalCacheStruct struct {
	Client    *bigcache.BigCache
	Expire    time.Duration
}

func (this *LocalCacheStruct) Has(key any) (ok bool) {

	_, err := this.Client.Get(cast.ToString(key))
	return utils.Ternary[bool](err != nil, false, true)
}

func (this *LocalCacheStruct) Get(key any) (value any) {

	result, err := this.Client.Get(cast.ToString(key))
	return utils.Ternary[any](err != nil, nil, utils.Json.Decode(string(result)))
}

func (this *LocalCacheStruct) Set(key any, value any, expire ...any) bool {

	err := this.Client.Set(cast.ToString(key), []byte(utils.Json.Encode(value)))
	return utils.Ternary[bool](err != nil, false, true)
}

func (this *LocalCacheStruct) Del(key any) bool {

	err := this.Client.Delete(cast.ToString(key))
	return utils.Ternary[bool](err != nil, false, true)
}