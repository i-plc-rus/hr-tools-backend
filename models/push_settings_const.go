package models

type SpacePushSettingCode string

type PushEvent struct {
	Name string
	Msg  string
}

var PushCodeMap = map[SpacePushSettingCode]PushEvent{
	//TODO тут спискок ивентов
	Check:     {Name: "Проверка", Msg: "Сообщение для проверки"},
	// PushLoginFail: {Name: "Ошибка (напр., неверный логин/пароль)", Msg: "Неверные учетные данные"},
}

const (
	Check     SpacePushSettingCode = "Check"
	// PushLoginFail SpacePushSettingCode = "PushLoginFail"
)
