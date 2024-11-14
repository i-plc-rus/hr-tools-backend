package apimodels

type Response struct {
	Status  string      `json:"status"`            //результат обработки fail/success
	Message string      `json:"message,omitempty"` //сообщение ошибки
	Data    interface{} `json:"data,omitempty"`    //данные ответа
}

type ScrollerResponse struct {
	Response
	RowCount int64 `json:"row_count,omitempty"` //для списков, общее кол-во записей, учитывая фильтр (если он есть)
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
	Limit int `json:"limit"` // Записей на странице
	Page  int `json:"page"`  // Страница (1,2,3..)
}

func (r Pagination) Validate() error {
	return nil
}

func (r Pagination) GetPage() (page, limit int) {
	page = 1
	limit = 10
	if r.Page > 0 {
		page = r.Page
	}
	if r.Limit > 0 {
		limit = r.Limit
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func NewScrollerResponse(data interface{}, rowCount int64) ScrollerResponse {
	return ScrollerResponse{
		Response: Response{
			Status: "success",
			Data:   data,
		},
		RowCount: rowCount,
	}
}
