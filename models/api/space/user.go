package spaceapimodels

type CreateUser struct {
	Password string `json:"password"`
	SpaceUserCommonData
}

type UpdateUser struct {
	Password string `json:"password"`
	SpaceUserCommonData
}

type SpaceUser struct {
	ID string `json:"id"`
	SpaceUserCommonData
}

type SpaceUserCommonData struct {
	SpaceID     string `json:"space_id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	IsAdmin     bool   `json:"is_admin"`
	Role        string `json:"role"`
}

func (r SpaceUserCommonData) Validate() error {
	//TODO: add data validators
	return nil
}
