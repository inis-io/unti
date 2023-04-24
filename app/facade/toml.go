package facade

import (
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"strings"
)

const (
	TomlCache   = "cache"
	TomlDb      = "db"
	TomlSMS     = "sms"
	TomlStorage = "storage"
	TomlPay     = "pay"
	TomlLog     = "log"
	TomlApp     = "app"
)

// NweToml - 获取配置文件
/**
 * @param mode 驱动模式
 * @return *utils.ViperResponse
 * @example：
 * 1. storage := facade.NweToml("cache")
 * 2. storage := facade.NweToml(facade.TomlStorage)
 */
func NweToml(mode any) *utils.ViperResponse {
	switch strings.ToLower(cast.ToString(mode)) {
	case TomlCache:
		return CacheToml
	case TomlDb:
		return DBToml
	case TomlSMS:
		return SMSToml
	case TomlStorage:
		return StorageToml
	case TomlPay:
		return PayToml
	case TomlLog:
		return LogToml
	case TomlApp:
		return AppToml
	}
	return nil
}
