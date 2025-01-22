package wsmodels

import ()

type ServerMessage struct {
	ToUserID string `json:"-"`
	Time     string `json:"time"`  // время события
	Code     string `json:"code"`  // код события
	Title    string `json:"title"` // заголовок события
	Msg      string `json:"msg"`   // текст события
}
