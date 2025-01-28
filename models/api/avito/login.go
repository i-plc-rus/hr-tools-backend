package avitoapimodels

import "time"

type RequestToken struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code,omitempty"`
}

type RefreshToken struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ResponseToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIN    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

type TokenData struct {
	ResponseToken
	ExpiresAt time.Time
}

type SelfData struct {
	Email string `json:"email"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
}
