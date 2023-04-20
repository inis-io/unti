package validator

import (
	"errors"
	"github.com/unti-io/go-utils/utils"
)

type Valid interface {
	Message() map[string]string
	Struct() any
}

func NewValid(table string, params map[string]any) (err error) {

	var item Valid

	switch table {
	case "users":
		item = &Users{}
	case "auth-group":
		item = &AuthGroup{}
	case "auth-rules":
		item = &AuthRules{}
	default:
		return errors.New("未知的验证器！")
	}

	return utils.Validate(item.Struct()).Message(item.Message()).Check(params)
}

// 使用方式 1：(推荐) - 接口方式 - 默认结构体和错误提示用这种
// err := validator.NewValid("users", params)
// 使用方式 2：(自定义) - 自定义结构体和错误提示用这种
// err := utils.Validate(validator.Users{}).Message(validator.UsersMessage).Check(params)