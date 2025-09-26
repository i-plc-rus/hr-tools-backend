package dbmodels

type QuestionHistory struct {
	BaseModel
	VacancyID    string `gorm:"type:varchar(36)" comment:"Идентификатор вакансии"`
	JobTitleName string `gorm:"type:varchar(255)"`
	VacancyName  string `gorm:"type:varchar(255)"`
	Text         string `json:"text"`    // Текст вопроса
	Comment      string `json:"comment"` // Комментарий к вопросу
}
