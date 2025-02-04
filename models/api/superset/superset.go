package supersetapimodels

type SupersetLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Provider string `json:"provider"`
	Refresh  bool   `json:"refresh"`
}

type SupersetGuestTokenReq struct {
	Resources []Resource `json:"resources"`
	RLS       []RLS      `json:"rls"`
	User      User       `json:"user"`
}

type Resource struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type RLS struct {
	Clause  string `json:"clause"`
	Dataset int    `json:"dataset"`
}

type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}
