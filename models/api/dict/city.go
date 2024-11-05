package dictapimodels

import dbmodels "hr-tools-backend/models/db"

type CityData struct {
	Address string `json:"address"`
}

type CityView struct {
	CityData
	ID string `json:"id"`
}

func CityConvert(rec dbmodels.City) CityView {
	return CityView{
		CityData: CityData{
			Address: rec.Address,
		},
		ID: rec.ID,
	}
}
