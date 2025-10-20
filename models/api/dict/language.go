package dictapimodels

import dbmodels "hr-tools-backend/models/db"

type LangData struct {
	Name string `json:"name"`
}

type LangView struct {
	LangData
	ID string `json:"id"`
}

func LangConvert(rec dbmodels.LanguageData) LangView {
	return LangView{
		LangData: LangData{
			Name: rec.Name,
		},
		ID: rec.ID,
	}
}
