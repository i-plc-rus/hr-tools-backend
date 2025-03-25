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
