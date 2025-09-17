package gpthandler

import (
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	ailogstore "hr-tools-backend/lib/gpt/store"
	yagptclient "hr-tools-backend/lib/gpt/yagpt-client"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	"hr-tools-backend/models"
	gptmodels "hr-tools-backend/models/api/gpt"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	GenerateVacancyDescription(spaceID, text string) (resp gptmodels.GenVacancyDescResponse, err error)
	GenerateHRSurvey(spaceID, vacancyID, vacancyInfo string) (resp gptmodels.GenVacancyDescResponse, err error)
	ReGenerateHRSurvey(spaceID, vacancyID, vacancyInfo, questions string) (resp gptmodels.GenVacancyDescResponse, err error)
	GenerateApplicantSurvey(spaceID, vacancyID, vacancyInfo, applicantInfo, hrSurvey string) (resp gptmodels.GenVacancyDescResponse, err error)
	ScoreApplicant(spaceID, vacancyID, vacancyInfo, applicantInfo, hrSurvey, applicantAnswers string) (resp gptmodels.GenVacancyDescResponse, err error)
}

type impl struct {
	spaceSettingsStore spacesettingsstore.Provider
	logStore           ailogstore.Provider
	useFakeAi          bool
}

var Instance Provider

func NewHandler(useFakeAi bool) {
	Instance = impl{
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		logStore:           ailogstore.NewInstance(db.DB),
		useFakeAi:          useFakeAi,
	}
}

func GetHandler(useFakeAi bool) *impl {
	return &impl{
		spaceSettingsStore: spacesettingsstore.NewInstance(db.DB),
		logStore:           ailogstore.NewInstance(db.DB),
		useFakeAi:          useFakeAi,
	}
}

const (
	HrSurveySysPromt        = "Ты — нейросеть, помогаешь HR-специалистам формировать опрос для оценки кандидатов."
	HrSurveyPromt1          = "У нас есть вакансия: %v \r\nНужно:"
	HrSurveyPromt2Gen       = "1. Сгенерировать 5 вопросов (3 с одиночным выбором, 2 со свободным ответом) по ключевым аспектам: опыт, навыки, soft skills."
	HrSurveyPromt2ReGen     = "1. Вопросы: %v не подошли. Сгенерируй новые вопросы с аналогичными типами."
	HrSurveyPromt3          = "2. Формат ответа – JSON со структурой: { \"questions\": [ { \"question_id\": \"qX\", \"question_text\": \"...\", \"question_type\": \"single_choice\"/\"free_text\", \"answers\": [ {\"value\": \"...\"} ], \"comment\": \"...\" } ] }."
	HrSurveyPromt4          = "3. Каждый вопрос должен сопровождаться кратким комментарием."
	HrSurveyPromt5          = "4. Варианты ответов для одиночного выбора: \"Обязательно\", \"Желательно\", \"Не требуется\" + \"Не подходит\" (для перегенерации)."
	HrSurveyPromt6          = "5. Свободные ответы включают опцию \"Не подходит\"."
	ApplicantSurveySysPromt = "Ты — нейросеть, помогаешь HR формировать опрос для кандидатов."
	ApplicantSurveyPromt1   = "Вакансия: %v"
	ApplicantSurveyPromt2   = "Кандидат: %v"
	ApplicantSurveyPromt3   = "Приоритеты HR: %v"
	ApplicantSurveyPromt4   = "Нужно:"
	ApplicantSurveyPromt5   = "1. Сгенерировать 5 вопросов для оценки соответствия."
	ApplicantSurveyPromt6   = "2. Формат: { \"questions\": [ { \"question_id\": \"\", \"question_text\": \"\", \"question_type\": \"\", \"answers\": [], \"weight\": <число>, \"comment\": \"\" } ] }."
	ApplicantSurveyPromt7   = "3. Веса соответствуют анкете HR."

	ApplicantScoreSysPromt = "Ты — нейросеть, помогаешь HR оценивать кандидатов."
	ApplicantScorePromt1   = "Вакансия: %v"
	ApplicantScorePromt2   = "Кандидат: %v"
	ApplicantScorePromt3   = "Приоритеты HR: %v"
	ApplicantScorePromt4   = "Ответы кандидата: %v"
	ApplicantScorePromt5   = `Алгоритмическая оценка:"
- c1: 30 баллов (вес 30)
- c2: 10 баллов (вес 20)
- c3: 30 баллов (вес 30)
- c4: 15 баллов (вес 15)
- c5: 15 баллов (вес 15)
Итог: 90 баллов
Нужно:
1. Сгенерировать комментарий для каждого вопроса, объясняющий оценку, с учётом приоритетов HR и данных кандидата.
2. Сгенерировать итоговый комментарий, суммирующий соответствие кандидата вакансии.
3. Формат: {"details": [ { "question_id": "", "score": <число>, "comment": "<текст>" } ], "comment": "<итоговый текст>" }`
)

func (i impl) GenerateVacancyDescription(spaceID, text string) (resp gptmodels.GenVacancyDescResponse, err error) {
	promt, err := i.spaceSettingsStore.GetValueByCode(spaceID, models.YandexGPTPromtSetting)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка получения инструкции для YandexGPT из настройки space")
		return resp, err
	}
	if promt == "" {
		log.
			WithField("space_id", spaceID).
			Warn("инструкция для YandexGPT из настройки space не должна быть пустой")
		return resp, errors.New("инструкция для YandexGPT из настройки space не должна быть пустой")
	}
	//promt := "Ты - рекрутер компании Рога и Копыта. В компании придерживаемся свободного стиля, используем эмодзи в текстах вакансии"
	userPromt := fmt.Sprintf("Сгенерируй описание для вакансии имея эти вводные данные: %s", text)
	resp.Description, err = i.getYaClient().
		GenerateByPromtAndText(promt, userPromt)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка генерации описания через YandexGPT")
		return resp, err
	}
	i.saveLog(spaceID, "", promt, userPromt, resp.Description, dbmodels.AiHRSurveyType)
	return resp, nil
}

func (i impl) GenerateHRSurvey(spaceID, vacancyID, vacancyInfo string) (resp gptmodels.GenVacancyDescResponse, err error) {
	userPromt := fmt.Sprintf("%v\r\n%v\r\n%v\r\n%v\r\n%v\r\n%v",
		fmt.Sprintf(HrSurveyPromt1, vacancyInfo),
		HrSurveyPromt2Gen,
		HrSurveyPromt3,
		HrSurveyPromt4,
		HrSurveyPromt5,
		HrSurveyPromt6,
	)

	resp.Description, err = i.getYaClient().
		GenerateByPromtAndText(HrSurveySysPromt, userPromt)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка генерации HR анкеты через YandexGPT")
		return resp, err
	}
	i.saveLog(spaceID, vacancyID, HrSurveySysPromt, userPromt, resp.Description, dbmodels.AiHRSurveyType)
	return resp, nil
}

func (i impl) ReGenerateHRSurvey(spaceID, vacancyID, vacancyInfo, questions string) (resp gptmodels.GenVacancyDescResponse, err error) {
	userPromt := fmt.Sprintf("%v\n%v\n%v\n%v\n%v\n%v",
		fmt.Sprintf(HrSurveyPromt1, vacancyInfo),
		fmt.Sprintf(HrSurveyPromt2ReGen, questions),
		HrSurveyPromt3,
		HrSurveyPromt4,
		HrSurveyPromt5,
		HrSurveyPromt6,
	)

	resp.Description, err = i.getYaClient().
		GenerateByPromtAndText(HrSurveySysPromt, userPromt)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка перегенерации вопросов для HR анкеты через YandexGPT")
		return resp, err
	}
	i.saveLog(spaceID, vacancyID, HrSurveySysPromt, userPromt, resp.Description, dbmodels.AiRegenHRSurveyType)
	return resp, nil
}

func (i impl) GenerateApplicantSurvey(spaceID, vacancyID, vacancyInfo, applicantInfo, hrSurvey string) (resp gptmodels.GenVacancyDescResponse, err error) {
	userPromt := fmt.Sprintf("%v\r\n%v\r\n%v\r\n%v\r\n%v\r\n%v\r\n%v",
		fmt.Sprintf(ApplicantSurveyPromt1, vacancyInfo),
		fmt.Sprintf(ApplicantSurveyPromt2, applicantInfo),
		fmt.Sprintf(ApplicantSurveyPromt3, hrSurvey),
		ApplicantSurveyPromt4,
		ApplicantSurveyPromt5,
		ApplicantSurveyPromt6,
		ApplicantSurveyPromt7,
	)

	resp.Description, err = i.getYaClient().
		GenerateByPromtAndText(ApplicantSurveySysPromt, userPromt)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка перегенерации вопросов для HR анкеты через YandexGPT")
		return resp, err
	}
	i.saveLog(spaceID, vacancyID, ApplicantSurveySysPromt, userPromt, resp.Description, dbmodels.AiApplicantSurveyType)
	return resp, nil
}

func (i impl) ScoreApplicant(spaceID, vacancyID, vacancyInfo, applicantInfo, hrSurvey, applicantAnswers string) (resp gptmodels.GenVacancyDescResponse, err error) {
	userPromt := fmt.Sprintf("%v\r\n%v\r\n%v\r\n%v\r\n%v",
		fmt.Sprintf(ApplicantScorePromt1, vacancyInfo),
		fmt.Sprintf(ApplicantScorePromt2, applicantInfo),
		fmt.Sprintf(ApplicantScorePromt3, hrSurvey),
		fmt.Sprintf(ApplicantScorePromt4, applicantAnswers),
		ApplicantScorePromt5,
	)

	resp.Description, err = i.getYaClient().
		GenerateByPromtAndText(ApplicantScoreSysPromt, userPromt)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка оценки вопросов для HR анкеты через YandexGPT")
		return resp, err
	}
	i.saveLog(spaceID, vacancyID, ApplicantScoreSysPromt, userPromt, resp.Description, dbmodels.AiScoreApplicantType)
	return resp, nil
}

func (i impl) getYaClient() yagptclient.Provider {
	if i.useFakeAi {
		return yagptclient.NewFakeClient("", "")
	}
	return yagptclient.
		NewClient(config.Conf.AI.YandexGPT.IAMToken, config.Conf.AI.YandexGPT.CatalogID)
}

func (i impl) saveLog(spaceID, vacancyID, sysPromt, userPromt, answer string, reqType dbmodels.AiReqestType) {
	rec := dbmodels.AiLog{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		SysPromt:   sysPromt,
		UserPromt:  userPromt,
		Answer:     answer,
		VacancyID:  vacancyID,
		ReqestType: reqType,
		AiName:     dbmodels.AiYaGptType,
	}
	_, err := i.logStore.Save(rec)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка сохранения лога ИИ")
	}
}
