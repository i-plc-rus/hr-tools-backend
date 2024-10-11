package apimodels

type Response struct {
	Status  string      `json:"status"`            //результат обработки fail/success
	Message string      `json:"message,omitempty"` //сообщение ошибки
	Data    interface{} `json:"data,omitempty"`    //данные ответа
}

func NewError(message string) Response {
	return Response{
		Status:  "fail",
		Message: message,
	}
}

func NewResponse(data interface{}) Response {
	return Response{
		Status: "success",
		Data:   data,
	}
}

type Pagination struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

func (r Pagination) Validate() error {
	return nil
}
