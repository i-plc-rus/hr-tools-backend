package applicantdict

import (
	applicantapimodels "hr-tools-backend/models/api/applicant"
)

var hrReasons = []string{
	"Не дозвонились",
	"Отказ образование",
	"Отказ переезд",
	"Отказ опыт работы",
	"Отказ график работы",
	"Отказ пол",
	"Отказ гражданство",
	"Отказ возраст",
	"Отказ на работном сайте",
	"Плохо выполненное тестовое задание",
	"Недостаток мотивации",
	"Отсутствие качеств, необходимых для позиции / компании",
	"Недостаток опыта",
}

var headReasons = []string{
	"Отсутствие качеств, необходимых для позиции / компании",
	"Плохо выполненное тестовое задание",
	"Недостаток мотивации",
	"Недостаток опыта",
}

var applicantReasons = []string{
	"Плохое впечатление от менеджемента",
	"Контроффер",
	"Неинтересна компания/сфера",
	"Неинтересные задачи/обязанности",
}

func GetRejectReasonList() applicantapimodels.RejectReasons {
	return applicantapimodels.RejectReasons{
		HrReasons:        hrReasons,
		HeadReasons:      headReasons,
		ApplicantReasons: applicantReasons,
	}
}
