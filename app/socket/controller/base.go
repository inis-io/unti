package controller

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"reflect"
)

type base struct{}

// Call 方法调用 - 资源路由本体
func (this base) call(allow map[string]any, name string, params ...any) ([]reflect.Value, error) {

	method := reflect.ValueOf(allow[name])

	if len(params) != method.Type().NumIn() {
		return nil, errors.New("输入参数的数量不匹配！")
	}

	in := make([]reflect.Value, len(params))

	for key, val := range params {
		in[key] = reflect.ValueOf(val)
	}

	return method.Call(in), nil
}

// Json
// 格式化JSON数据
func Json(data any) map[string]any {
	var result map[string]any
	json.Unmarshal([]byte(cast.ToString(data)), &result)
	return result
}

// 生成唯一的ID
func guid() string {
	return uuid.New().String()
}
