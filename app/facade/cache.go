package facade

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/fsnotify/fsnotify"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
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
	// CacheModeRedis - Redis缓存
	CacheModeRedis = "redis"
	// CacheModeFile  - 文件缓存
	CacheModeFile  = "file"
	// CacheModeRAM   - 内存缓存
	CacheModeRAM   = "ram"
)

// NewCache - 创建Cache实例
/**
 * @param mode 驱动模式
 * @return CacheInterface
 * @example：
 * 1. cache := facade.NewCache("redis")
 * 2. cache := facade.NewCache(facade.CacheModeRedis)
 */
func NewCache(mode any) CacheInterface {
	switch strings.ToLower(cast.ToString(mode)) {
	case CacheModeRedis:
		Cache = Redis
	case CacheModeFile:
		Cache = FileCache
	case CacheModeRAM:
		Cache = BigCache
	default:
		Cache = FileCache
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
		Content: utils.Replace(TempCache, map[string]any{
			"${open}":           "false",
			"${default}":        "file",
			"${local.expire}":   300,
			"${redis.host}":     "localhost",
			"${redis.port}":     "6379",
			"${redis.password}": "",
			"${redis.expire}":   "2 * 60 * 60",
			"${redis.prefix}":   "unti:",
			"${redis.database}": 0,
			"${file.expire}"   : "2 * 60 * 60",
			"${file.path}":      "runtime/cache",
			"${file.prefix}":    "unti_",
			"${ram.expire}":     "2 * 60 * 60",
		}),
	}).Read()

	if item.Error != nil {
		Log.Error(map[string]any{
			"error":     item.Error,
			"func_name": utils.Caller().FuncName,
			"file_name": utils.Caller().FileName,
			"file_line": utils.Caller().Line,
		}, "缓存配置初始化错误")
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
	redisExpire := time.Duration(cast.ToInt(utils.Calc(CacheToml.Get("redis.expire", 7200)))) * time.Second

	// Redis 缓存
	Redis = &RedisCacheStruct{
		Client: redisClient,
		Prefix: redisPrefix,
		Expire: redisExpire,
	}

	// 文件缓存
	FileClient, _ := utils.NewFileCache(
		CacheToml.Get("file.path"),
		utils.Calc(CacheToml.Get("file.expire", 7200)),
		CacheToml.Get("file.prefix"),
	)

	// File 缓存
	FileCache = &FileCacheStruct{
		Client: FileClient,
	}

	// BigCache 缓存
	BigCache = &BigCacheStruct{
		Client: NewBigCache(utils.Calc(CacheToml.Get("file.expire", 7200))),
	}

	switch cast.ToString(CacheToml.Get("default")) {
	case CacheModeRedis:
		Cache = Redis
	case CacheModeFile:
		Cache = FileCache
	case CacheModeRAM:
		Cache = BigCache
	default:
		Cache = FileCache
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
var FileCache *FileCacheStruct
var BigCache *BigCacheStruct

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
	// DelPrefix
	/**
	 * @name 删除前缀缓存
	 * @param prefix 缓存的前缀
	 * @return bool
	 */
	DelPrefix(prefix ...any) (ok bool)
	// DelTags
	/**
	 * @name 删除标签缓存
	 * @param tag 缓存的标签
	 * @return bool
	 */
	DelTags(tag ...any) (ok bool)
	// Clear
	/**
	 * @name 清空缓存
	 * @return bool
	 */
	Clear() (ok bool)
}


// ==================== Redis 缓存 ====================


type RedisCacheStruct struct {
	Client *redis.Client
	Prefix string
	Expire time.Duration
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

func (this *RedisCacheStruct) Set(key any, value any, expire ...any) (ok bool) {

	ctx := context.Background()
	// 设置过期时间
	if len(expire) == 0 {
		expire = append(expire, this.Expire)
	}

	// 如果 exp不为时间类型，则转码为时间类型
	if reflect.ValueOf(expire[0]).Kind() != reflect.Int64 && expire[0] != 0 {
		expire[0] = time.Duration(cast.ToInt(expire[0])) * time.Second
	}

	err := this.Client.Set(ctx, this.Prefix+cast.ToString(key), utils.Json.Encode(value), cast.ToDuration(expire[0])).Err()
	return utils.Ternary[bool](err != nil, false, true)
}

func (this *RedisCacheStruct) Del(key any) (ok bool) {

	ctx := context.Background()
	err := this.Client.Del(ctx, this.Prefix+cast.ToString(key)).Err()
	return utils.Ternary[bool](err != nil, false, true)
}

func (this *RedisCacheStruct) DelPrefix(prefix ...any) (ok bool) {

	var keys []string
	var prefixes []string
	ctx := context.Background()

	if len(prefix) == 0 {
		return false
	}

	for _, value := range prefix {
		// 判断是否为切片
		if reflect.ValueOf(value).Kind() == reflect.Slice {
			for _, val := range cast.ToSlice(value) {
				prefixes = append(prefixes, this.Prefix+cast.ToString(val)+"*")
			}
		} else {
			prefixes = append(prefixes, this.Prefix+cast.ToString(value)+"*")
		}
	}

	// 获取 prefixes 所有的key
	for _, val := range prefixes {
		item, err := this.Client.Keys(ctx, val).Result()
		if err != nil {
			return false
		}
		keys = append(keys, item...)
	}

	// 去重 - 去空
	keys = cast.ToStringSlice(utils.Array.Empty(utils.ArrayUnique(keys)))

	if len(keys) > 0 {
		err := this.Client.Del(ctx, keys...).Err()
		if err != nil {
			return false
		}
	}

	return true
}

func (this *RedisCacheStruct) DelTags(tag ...any) (ok bool) {

	var keys []string
	var tags []string
	ctx := context.Background()

	if len(tag) == 0 {
		return false
	}

	for _, value := range tag {

		var item string

		// 判断是否为切片
		if reflect.ValueOf(value).Kind() == reflect.Slice {
			var tmp []string
			for _, val := range cast.ToSlice(value) {
				tmp = append(tmp, cast.ToString(val))
			}
			// 数组分割字符串
			item = strings.Join(tmp, "*")
		} else {
			item = cast.ToString(value)
		}
		tags = append(tags, fmt.Sprintf("%s*%s*", this.Prefix, item))
	}

	// 获取 prefixes 所有的key
	for _, val := range tags {
		item, err := this.Client.Keys(ctx, val).Result()
		if err != nil {
			return false
		}
		keys = append(keys, item...)
	}

	// 去重 - 去空
	keys = cast.ToStringSlice(utils.Array.Empty(utils.ArrayUnique(keys)))

	if len(keys) > 0 {
		err := this.Client.Del(ctx, keys...).Err()
		if err != nil {
			return false
		}
	}

	return true
}

func (this *RedisCacheStruct) Clear() (ok bool) {

	ctx := context.Background()
	err := this.Client.FlushDB(ctx).Err()
	return utils.Ternary[bool](err != nil, false, true)
}


// ============================ 文件缓存 ============================


type FileCacheStruct struct {
	Client *utils.FileCacheClient
}

func (this *FileCacheStruct) Has(key any) (ok bool) {
	return this.Client.Has(key)
}

func (this *FileCacheStruct) Get(key any) (value any) {
	return utils.Json.Decode(this.Client.Get(key))
}

func (this *FileCacheStruct) Set(key any, value any, expire ...any) (ok bool) {
	return this.Client.Set(key, []byte(utils.Json.Encode(value)), expire...)
}

func (this *FileCacheStruct) Del(key any) (ok bool) {
	return this.Client.Del(key)
}

func (this *FileCacheStruct) DelPrefix(prefix ...any) (ok bool) {
	return this.Client.DelPrefix(prefix...)
}

func (this *FileCacheStruct) DelTags(tag ...any) (ok bool) {
	return this.Client.DelTags(tag...)
}

func (this *FileCacheStruct) Clear() (ok bool) {
	return this.Client.Clear()
}


// ============================ 内存缓存 ============================


type BigCacheStruct struct {
	Client *BigCacheClient
}

func (this *BigCacheStruct) Has(key any) (ok bool) {
	return this.Client.Has(key)
}

func (this *BigCacheStruct) Get(key any) (value any) {
	return utils.Json.Decode(this.Client.Get(key))
}

func (this *BigCacheStruct) Set(key any, value any, expire ...any) (ok bool) {
	return this.Client.Set(key, []byte(utils.Json.Encode(value)), expire...)
}

func (this *BigCacheStruct) Del(key any) (ok bool) {
	return this.Client.Del(key)
}

func (this *BigCacheStruct) DelPrefix(prefix ...any) (ok bool) {
	return this.Client.DelPrefix(prefix...)
}

func (this *BigCacheStruct) DelTags(tag ...any) (ok bool) {
	return this.Client.DelTags(tag...)
}

func (this *BigCacheStruct) Clear() (ok bool) {
	return this.Client.Clear()
}

// BigCacheClient 缓存
type BigCacheClient struct {
	mutex      sync.Mutex   	// 互斥锁，用于保证并发安全
	prefix	   string			// 缓存文件名前缀
	expire	   int64			// 默认缓存过期时间
	items 	   map[string]*bigcache.BigCache
}

// NewBigCache 创建一个新的缓存实例
func NewBigCache(expire any, prefix ...string) *BigCacheClient {

	var cache BigCacheClient

	cache.expire = cast.ToInt64(expire)
	cache.items  = make(map[string]*bigcache.BigCache)
	cache.prefix = "cache_"
	if len(prefix) > 0 {
		cache.prefix = prefix[0]
	}

	return &cache
}

// Get 获取缓存
func (this *BigCacheClient) Get(key any) (result []byte) {
	res, err := this.GetE(key)
	return utils.Ternary(err != nil, nil, res)
}

// Has 判断缓存是否存在
func (this *BigCacheClient) Has(key any) (ok bool) {
	_, ok = this.items[this.name(key)]
	return
}

// Set 设置缓存
func (this *BigCacheClient) Set(key any, value []byte, expire ...any) (ok bool) {

	exp := this.expire

	if len(expire) > 0 {
		if !utils.Is.Empty(expire[0]) {
			// 判断 expire[0] 是否为Duration类型
			if reflect.TypeOf(expire[0]).String() == "time.Duration" {
				// 转换为int64
				exp = cast.ToInt64(cast.ToDuration(expire[0]).Seconds())
			} else {
				exp = cast.ToInt64(expire[0])
			}
		}
	}

	err := this.SetE(key, value, exp)

	return utils.Ternary(err != nil, false, true)
}

// Del 删除缓存
func (this *BigCacheClient) Del(key any) (ok bool) {

	err := this.DelE(key)

	return utils.Ternary(err != nil, false, true)
}

// Clear 清空缓存
func (this *BigCacheClient) Clear() (ok bool) {

	err := this.ClearE()

	return utils.Ternary(err != nil, false, true)
}

// DelPrefix 根据前缀删除缓存
func (this *BigCacheClient) DelPrefix(prefix ...any) (ok bool) {
	err := this.DelPrefixE(prefix...)
	return utils.Ternary(err != nil, false, true)
}

// DelTags 根据标签删除缓存
func (this *BigCacheClient) DelTags(tags ...any) (ok bool) {
	err := this.DelTagsE(tags...)
	return utils.Ternary(err != nil, false, true)
}

// GetE 获取缓存
func (this *BigCacheClient) GetE(key any) (result []byte, err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	item, ok := this.items[this.name(key)]
	if !ok {
		delete(this.items, this.name(key))
		return nil, fmt.Errorf("cache %s not exists", this.name(key))
	}

	value, err := item.Get(this.name(key))
	if err != nil {
		return nil, err
	}

	return value, nil
}

// SetE 设置缓存
func (this *BigCacheClient) SetE(key any, value []byte, expire int64) (err error) {

	// end 过期时间，expire = 0 表示永不过期
	var end time.Duration
	if expire == 0 {
		// 100年后过期
		end = time.Duration(100 * 365 * 24 * 60 * 60 * 1e9)
	} else {
		end = time.Duration(expire) * time.Second
	}

	item, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(end))

	err = item.Set(this.name(key), value)

	if err != nil {
		return err
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.items[this.name(key)] = item

	return nil
}

// DelE 删除缓存
func (this *BigCacheClient) DelE(key any) (err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	item, ok := this.items[this.name(key)]
	if !ok {
		return fmt.Errorf("cache %s not exists", this.name(key))
	}

	err = item.Delete(this.name(key))
	if err != nil {
		return err
	}

	delete(this.items, this.name(key))

	return nil
}

// ClearE 清空缓存
func (this *BigCacheClient) ClearE() (err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, item := range this.items {
		err = item.Reset()
		if err != nil {
			return err
		}
	}

	return nil
}

// DelPrefixE 删除指定前缀的缓存
func (this *BigCacheClient) DelPrefixE(prefix ...any) (err error) {

	this.mutex.Lock()
	defer this.mutex.Unlock()

	for key, item := range this.items {
		if strings.HasPrefix(key, cast.ToString(prefix)) {
			err = item.Reset()
			if err != nil {
				return err
			}
			delete(this.items, key)
		}
	}

	return nil
}

// DelTagsE 删除指定标签的缓存
func (this *BigCacheClient) DelTagsE(tag ...any) (err error) {

	var keys []string
	var tags []string

	if len(tag) == 0 {
		return nil
	}

	for _, value := range tag {

		var item string

		// 判断是否为切片
		if reflect.ValueOf(value).Kind() == reflect.Slice {
			var tmp []string
			for _, val := range cast.ToSlice(value) {
				tmp = append(tmp, cast.ToString(val))
			}
			item = strings.Join(tmp, "*")
		} else {
			item = cast.ToString(value)
		}

		tags = append(tags, fmt.Sprintf("*%s*", item))
	}

	// 获取所有缓存名称
	for key := range this.items {
		keys = append(keys, key)
	}

	// 模糊匹配
	keys = this.fuzzyMatch(keys, tags)

	for _, key := range keys {
		item, ok := this.items[key]
		if !ok {
			continue
		}
		err = item.Reset()
		if err != nil {
			return err
		}
		delete(this.items, key)
	}

	return nil
}

// name 获取缓存名称
func (this *BigCacheClient) name(key any) (result string) {
	return fmt.Sprintf("%s%s", this.prefix, cast.ToString(key))
}

// fuzzyMatch 模糊匹配
func (this *BigCacheClient) fuzzyMatch(keys []string, tags []string) (result []string) {
	for _, item := range keys {
		for _, tag := range tags {
			if matched, _ := filepath.Match(tag, item); matched {
				result = append(result, item)
				break
			}
		}
	}
	return result
}