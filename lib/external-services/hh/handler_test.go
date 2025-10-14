package hhhandler

import (
	"context"
	"github.com/stretchr/testify/require"
	"hr-tools-backend/lib/external-services/hh/hhclient"
	dbmodels "hr-tools-backend/models/db"
	"testing"
)

func TestGeoJSONHelpers(t *testing.T) {
	t.Run(`getArea check`, func(t *testing.T) {
		i := getInstance()
		city := dbmodels.City{}
		city.ID = "7baaac39-aa31-4712-893d-51ef80124a86"
		city.City = "Сергиев Посад"
		city.Region = "Московская"
		area, err := i.getArea(context.TODO(), &city)
		require.Nil(t, err)
		require.Equal(t, "2069", area.ID)

		city.ID = "0c5b2444-70a0-4932-980c-b4dc0d3f02b5"
		city.City = "Москва"
		city.Region = "Москва"
		area, err = i.getArea(context.TODO(), &city)
		require.Nil(t, err)
		require.Equal(t, "1", area.ID)

		city.ID = "0c5b2444-70a0-4932-980c-b4dc0d3f02b5"
		city.City = "г москва"
		city.Region = "москва и мо"
		area, err = i.getArea(context.TODO(), &city)
		require.Equal(t, "город публикации не найден в справочнике", err.Error())
	})

	t.Run(`fillVacancyData check`, func(t *testing.T) {
		rec := dbmodels.Vacancy{
			Salary: dbmodels.Salary{
				From:     100000,
				To:       110000,
				ByResult: 110000,
				InHand:   110000,
			},
			VacancyRequestID: nil,
			VacancyRequest:   nil,
			JobTitleID: new(string),
			JobTitle: &dbmodels.JobTitle{
				DepartmentID: "8f3937ef-134a-46f3-9a30-236f4d2d6e83",
				Name:         "Программист, разработчик",
				HhRoleID:     "11",
			},
			CityID: new(string),
			City: &dbmodels.City{
				Address: "г Москва",
				Region:  "Москва",
				City:    "Москва",
			},
			// CompanyStructID:  new(string),
			// CompanyStruct:    &dbmodels.CompanyStruct{},
			VacancyName:     "PHP-разработчик (Back-end)",
			OpenedPositions: 1,
			Urgency:         "В плановом порядке",
			RequestType:     "Новая позиция",
			SelectionType:   "Индивидуальный",
			PlaceOfWork:     "Староконюшенный переулок, 6",
			ChiefFio:        "Благово Емельян Николаевич",
			Requirements:    "<p><strong>Задачи:</strong></p><ul><li>Разработка&nbsp;API&nbsp;для&nbsp;веб&nbsp;и&nbsp;мобильных&nbsp;приложений&nbsp;(Yii2&nbsp;фреймворк)</li><li>Выполнение&nbsp;рефакторинга&nbsp;кода</li><li>Оптимизация&nbsp;выполнения&nbsp;запросов,&nbsp;анализ&nbsp;и&nbsp;профилирование&nbsp;выполнения&nbsp;программного&nbsp;кода,&nbsp;повышение&nbsp;общей&nbsp;отказоустойчивости&nbsp;системы</li><li>Разработка&nbsp;API&nbsp;для&nbsp;веб&nbsp;и&nbsp;мобильных&nbsp;приложений(Yii2&nbsp;фреймворк);</li><li>Участие&nbsp;в&nbsp;проработке&nbsp;архитектуры&nbsp;TraceWay&nbsp;системы.</li></ul><p></p><p><strong>Ожидаем&nbsp;от&nbsp;Вас:</strong></p><ul><li>Опыт&nbsp;работы&nbsp;от&nbsp;3&nbsp;лет;</li><li>Знание&nbsp;PHP;</li><li>Опыт&nbsp;работы&nbsp;с&nbsp;одним&nbsp;из&nbsp;современных&nbsp;фреймворков&nbsp;(Symfony,&nbsp;Zend,&nbsp;Laravel,&nbsp;Yii2);</li><li>Понимание&nbsp;и&nbsp;опыт&nbsp;практического&nbsp;применения&nbsp;ООП&nbsp;и&nbsp;разработки&nbsp;архитектуры&nbsp;баз&nbsp;данных;</li><li>Уверенное&nbsp;знание&nbsp;SQL&nbsp;(запросы&nbsp;придется&nbsp;писать&nbsp;в&nbsp;том&nbsp;числе&nbsp;без&nbsp;ORM&nbsp;и&nbsp;билдеров);</li><li>Опыт&nbsp;использования&nbsp;системы&nbsp;контроля&nbsp;версий&nbsp;(Git).</li></ul><p></p><p><strong>Условия:</strong></p><ul><li>Работа&nbsp;в&nbsp;аккредитованной&nbsp;IT-компании;</li><li>Официальное&nbsp;трудоустройство,&nbsp;стабильная&nbsp;заработная&nbsp;плата,&nbsp;соц.&nbsp;пакет;</li><li>Перспективы&nbsp;карьерного&nbsp;и&nbsp;профессионального&nbsp;роста;</li><li>Участие&nbsp;в&nbsp;интересных&nbsp;проектах;</li><li>Работа&nbsp;в&nbsp;профессиональной&nbsp;команде,&nbsp;которая&nbsp;готова&nbsp;делиться&nbsp;своим&nbsp;опытом;</li><li>Обучение;</li><li>Корпоративные&nbsp;мероприятия,&nbsp;тимбилдинги.</li></ul>",
			Status:          "Открыта",
			Employment: "full",
			Experience: "moreThan3",
			Schedule:   "fullDay",
		}
		i := getInstance()
		i.cityMap["Москва"] = "1"
		req, hMsg := i.fillVacancyData(context.TODO(), &rec)
		require.Equal(t, "", hMsg)
		require.NotNil(t, req)
		require.Equal(t, "1", req.Area.ID)
		require.Equal(t, "free", req.BillingType.ID)
		require.NotEmpty(t, req.Description)
		require.Equal(t, "PHP-разработчик (Back-end)", req.Name)
		require.Equal(t, "open", req.Type.ID)
		require.Len(t, req.ProfessionalRoles, 1)
		require.Equal(t, "11", req.ProfessionalRoles[0].ID)

		//optional
		require.NotNil(t, req.Schedule)
		require.Equal(t, "fullDay", req.Schedule.ID)
		require.NotNil(t, req.SalaryRange)
		require.Equal(t, "RUR", req.SalaryRange.Currency)
		require.Equal(t, "MONTH", req.SalaryRange.Mode.ID)
		require.Equal(t, true, req.AllowMessages)
		require.NotNil(t, req.EmploymentFrom)
		require.Equal(t, "FULL", req.EmploymentFrom.ID)
	})
}

func getInstance() impl {
	hhclient.NewProvider("https://a.hr-tools.pro/api/v1/oauth/callback/hh")
	return impl{
		client:  hhclient.Instance,
		cityMap: map[string]string{},
	}
}
