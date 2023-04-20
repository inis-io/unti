package validator

type AuthRules struct {
	Common int `json:"common" rule:"number,min:0,max:1"`
}

var AuthRulesMessage = map[string]string{
	"common.number":    "是否在线只能是数字！",
	"common.min":       "是否在线只能是0或1！",
	"common.max":       "是否在线只能是0或1！",
}

func (this AuthRules) Message() map[string]string {
	return AuthRulesMessage
}

func (this AuthRules) Struct() any {
	return this
}