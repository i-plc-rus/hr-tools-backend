package dbmodels

type City struct {
	BaseModel
	Address         string  `gorm:"index;type:varchar(255)"` // Адрес одной строкой
	PostalCode      string  `gorm:"type:varchar(50)"`        // Почтовый индекс
	Country         string  `gorm:"type:varchar(100)"`       // Страна
	FederalDistrict string  `gorm:"type:varchar(100)"`       // Федеральный округ
	RegionType      string  `gorm:"type:varchar(50)"`        // Тип региона
	Region          string  `gorm:"type:varchar(100)"`       // Регион
	AreaType        string  `gorm:"type:varchar(255)"`       // Тип района
	Area            string  `gorm:"type:varchar(255)"`       // Район
	CityType        string  `gorm:"type:varchar(255)"`       // Тип города
	City            string  `gorm:"type:varchar(255)"`       // Город
	SettlementType  string  `gorm:"type:varchar(50)"`        // Тип населенного пункта
	Settlement      string  `gorm:"type:varchar(255)"`       // Населенный пункт
	Okato           string  `gorm:"type:varchar(20)"`        // Код ОКАТО
	Oktmo           string  `gorm:"type:varchar(20)"`        // Код ОКТМО
	Timezone        string  `gorm:"type:varchar(10)"`        // Часовой пояс
	Lat             float64 // Широта
	Lon             float64 // Долгота
}
