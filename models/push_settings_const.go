package models

type SpacePushSettingCode string

type PushEvent struct {
	Name string
	Msg  string
}

var PushCodeMap = map[SpacePushSettingCode]PushEvent{
	//TODO тут спискок ивентов
	// PushLogin:     {Name: "Успешная авторизация", Msg: "Вы успешно вошли в систему"},
	// PushLoginFail: {Name: "Ошибка (напр., неверный логин/пароль)", Msg: "Неверные учетные данные"},
}

const (
	// PushLogin     SpacePushSettingCode = "PushLogin"
	// PushLoginFail SpacePushSettingCode = "PushLoginFail"
)
