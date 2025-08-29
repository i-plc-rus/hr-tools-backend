package hhapimodels

import "time"

type RequestToken struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code,omitempty"`
	RedirectUri  string `json:"redirect_uri,omitempty"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ResponseToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIN    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
}

type ErrorData struct {
	Error            string      `json:"error"` //400
	ErrorDescription string      `json:"error_description"`
	Errors           []ErrorItem `json:"errors"`      //400/403
	OauthError       string      `json:"oauth_error"` //403
}

func (e ErrorData) GetPublishErrorReason() string {
	if e.ErrorDescription != "" {
		return e.ErrorDescription
	}
	if len(e.Errors) != 0 {
		switch e.Errors[0].Value {
		case "not_enough_purchased_services":
			return "купленных услуг недостаточно для публикации или обновления данного типа вакансии"
		case "quota_exceeded":
			return "квота менеджера на публикацию данного типа вакансии закончилась"
		case "duplicate":
			return "аналогичная вакансия уже опубликована"
		case "replacement":
			return "вакансия существенно изменена, есть риски блокировки"
		case "creation_forbidden":
			return "публикация вакансий недоступна текущему менеджеру"
		case "unavailable_for_archived":
			return "редактирование недоступно для архивной вакансии"
		case "conflict_changes":
			return "конфликтные изменения данных вакансии"
		}
	}
	return "Ошибка публикация вакансии на HeadHunter"
}

type ErrorItem struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type TokenData struct {
	ResponseToken
	ExpiresAt time.Time
}

type MeResponse struct {
	ID                    string     `json:"id"`
	IsAdmin               bool       `json:"is_admin"`
	IsApplicant           bool       `json:"is_applicant"`
	IsApplication         bool       `json:"is_application"`
	IsEmployer            bool       `json:"is_employer"`
	IsEmployerIntegration bool       `json:"is_employer_integration"`
	AuthType              string     `json:"auth_type"`
	Email                 string     `json:"email"`
	LastName              string     `json:"last_name"`
	FirstName             string     `json:"first_name"`
	MiddleName            string     `json:"middle_name"`
	Employer              MeEmployer `json:"employer"`
}

type MeEmployer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (t TokenData) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(time.Second * time.Duration(t.ExpiresIN)))
}
