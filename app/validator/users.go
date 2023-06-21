package validator

type Users struct {
	Account     string `json:"account" rule:"alphaNum,min=4,max=32"`
	Password    string `json:"password" rule:"min=6,max=32"`
	Email       string `json:"email" rule:"email"`
	Nickname    string `json:"nickname" rule:"chsDash,max=32"`
	Description string `json:"description" rule:"max:256"`
}

var UsersMessage = map[string]string{
	"account.alphaNum": "账号只能是字母和数字！",
	"account.min":      "账号最少不能少于4个字符！",
	"account.max":      "账号最多不能超过32个字符！",
	"password.min":     "密码最少不能少于6个字符！",
	"password.max":     "密码最多不能超过32个字符！",
	"email.email":      "邮箱格式不正确！",
	"nickname.chsDash": "昵称只能是汉字、字母、数字和下划线_及破折号-！",
	"nickname.max":     "昵称最多不能超过32个字符！",
	"description.max":  "个人简介最多不能超过256个字符！",
}

func (this Users) Message() map[string]string {
	return UsersMessage
}

func (this Users) Struct() any {
	return this
}
