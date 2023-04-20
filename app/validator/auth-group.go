package validator

type AuthGroup struct {}

var AuthGroupMessage = map[string]string{}

func (this AuthGroup) Message() map[string]string {
	return AuthGroupMessage
}

func (this AuthGroup) Struct() any {
	return this
}