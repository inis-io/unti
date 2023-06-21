package validator

import (
	"errors"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"strings"
)

type Valid interface {
	Message() map[string]string
	Struct() any
}

// NewValid
/**
 * @name 验证器
 * @param table 表名
 * @param params 参数
 * @return err 错误
 * @example:
 * err := validator.NewValid("users", params)
 */
func NewValid(table any, params map[string]any) (err error) {

	var item Valid

	switch strings.ToLower(cast.ToString(table)) {
	case "users":
		item = &Users{}
	default:
		return errors.New("未知的验证器！")
	}

	return utils.Validate(item.Struct()).Message(item.Message()).Check(params)
}

// 使用方式 1：(推荐) - 接口方式 - 默认结构体和错误提示用这种
// err := validator.NewValid("users", params)
// 使用方式 2：(自定义) - 自定义结构体和错误提示用这种
// err := utils.Validate(validator.Users{}).Message(validator.UsersMessage).Check(params)
