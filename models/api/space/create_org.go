package spaceapimodels

type CreateOrganization struct {
	OrganizationName string     `json:"organization_name"`
	Inn              string     `json:"inn"`
	Kpp              string     `json:"kpp"`
	OGRN             string     `json:"ogrn"`
	FullName         string     `json:"full_name"`
	DirectorName     string     `json:"director_name"`
	AdminData        CreateUser `json:"admin_data"`
}

func (r CreateOrganization) Validate() error {
	//TODO заглушка
	return nil
}
