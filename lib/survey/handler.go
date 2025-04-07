package survey

import (
	"encoding/json"
	"fmt"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/applicant"
	applicanthistoryhandler "hr-tools-backend/lib/applicant-history"
	applicantstore "hr-tools-backend/lib/applicant/store"
	gpthandler "hr-tools-backend/lib/gpt"
	applicantsurveystore "hr-tools-backend/lib/survey/applicant-survey-store"
	vacancysurveystore "hr-tools-backend/lib/survey/vacancy-survey-store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	"hr-tools-backend/models"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"
	"sort"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	SaveHRSurvey(spaceID, vacancyID string, survey surveyapimodels.HRSurvey) (*surveyapimodels.HRSurveyView, error)
	GetHRSurvey(spaceID, vacancyID string) (*surveyapimodels.HRSurveyView, error)
	GetApplicantSurvey(spaceID, vacancyID, applicantID string) (*surveyapimodels.ApplicantSurveyView, error)
	GenApplicantSurvey(spaceID, vacancyID, applicantID string) (ok bool, err error)
	GetPublicApplicantSurvey(id string) (*surveyapimodels.ApplicantSurveyView, error)
	AnswerPublicApplicantSurvey(id string, answers []surveyapimodels.ApplicantSurveyAnswer) (hMsg string, err error)
	AIScore(applicantSurveyRec dbmodels.ApplicantSurvey) (ok bool, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		vSurveyStore:   vacancysurveystore.NewInstance(db.DB),
		vacancyStore:   vacancystore.NewInstance(db.DB),
		applicantStore: applicantstore.NewInstance(db.DB),
		aSurveyStore:   applicantsurveystore.NewInstance(db.DB),
	}
}

type impl struct {
	vSurveyStore   vacancysurveystore.Provider
	vacancyStore   vacancystore.Provider
	applicantStore applicantstore.Provider
	aSurveyStore   applicantsurveystore.Provider
}

func (i impl) SaveHRSurvey(spaceID, vacancyID string, survey surveyapimodels.HRSurvey) (*surveyapimodels.HRSurveyView, error) {
	rec := dbmodels.HRSurvey{
		BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
		IsFilledOut:    false,
		VacancyID:      vacancyID,
		Survey: dbmodels.HRSurveyQuestions{
			Questions: make([]dbmodels.HRSurveyQuestion, 0, len(survey.Questions)),
		},
	}
	regenerateQuestions := []dbmodels.HRSurveyQuestion{}
	selectedCount := 0
	for _, question := range survey.Questions {
		switch question.Selected {
		case "Обязательно":
			question.Weight = 30
		case "Желательно":
			question.Weight = 20
		case "Не требуется":
			question.Weight = 10
		case "Не подходит":
			regenerateQuestions = append(regenerateQuestions, question)
			question.Weight = 0
			// вопрос добавляем для перегенерации, из основного списка исключаем
			continue
		default:
			if question.QuestionType == "free_text" {
				question.Weight = 15
			} else {
				question.Weight = 0
			}
		}
		if question.Selected != "" {
			selectedCount++
		}
		rec.Survey.Questions = append(rec.Survey.Questions, question)
	}
	if len(regenerateQuestions) != 0 {
		// в анкете найдены вопросы для перегенерации
		vacancyRec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
		if err != nil {
			return nil, errors.Wrap(err, "ошибка получения вакансии")
		}
		regeneratedQuestions, err := i.regenerateSurvey(*vacancyRec, regenerateQuestions)
		if err != nil {
			return nil, err
		}
		rec.Survey.Questions = append(rec.Survey.Questions, regeneratedQuestions.Questions...)
	} else {
		if selectedCount == len(rec.Survey.Questions) {
			rec.IsFilledOut = true
		}
	}
	sort.Slice(rec.Survey.Questions, func(i, j int) bool {
		return rec.Survey.Questions[i].QuestionID < rec.Survey.Questions[j].QuestionID
	})

	_, err := i.vSurveyStore.Save(rec)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка сохранения анкеты")
	}
	result := surveyapimodels.HRSurveyView{
		IsFilledOut: rec.IsFilledOut,
		HRSurvey: surveyapimodels.HRSurvey{
			Questions: rec.Survey.Questions,
		},
	}
	return &result, nil
}

func (i impl) GetHRSurvey(spaceID, vacancyID string) (*surveyapimodels.HRSurveyView, error) {
	rec, err := i.vSurveyStore.GetByVacancyID(spaceID, vacancyID)
	if err != nil {
		return nil, err
	}
	if rec != nil {
		result := surveyapimodels.HRSurveyView{
			IsFilledOut: rec.IsFilledOut,
			HRSurvey: surveyapimodels.HRSurvey{
				Questions: rec.Survey.Questions,
			},
		}
		return &result, nil
	}
	// анкета не найдена, генерируем
	vacancyRec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения вакансии")
	}
	surveyData, err := i.generateSurvey(*vacancyRec)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка генерации анкеты")
	}
	return i.SaveHRSurvey(spaceID, vacancyID, *surveyData)
}

func (i impl) GetApplicantSurvey(spaceID, vacancyID, applicantID string) (*surveyapimodels.ApplicantSurveyView, error) {
	applicant, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения данных кандидата")
	}
	if applicant == nil {
		return nil, errors.New("кандидат не найден")
	}
	if applicant.ApplicantSurvey != nil {
		result := surveyapimodels.ApplicantSurveyView{
			ApplicantSurvey: surveyapimodels.ApplicantSurvey{
				Questions: applicant.ApplicantSurvey.Survey.Questions,
			},
			IsFilledOut: applicant.ApplicantSurvey.IsFilledOut,
		}
		return &result, nil
	}
	return nil, errors.New("опрос отсутсвует")
}

func (i impl) GenApplicantSurvey(spaceID, vacancyID, applicantID string) (ok bool, err error) {
	applicant, err := i.applicantStore.GetByID(spaceID, applicantID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения данных кандидата")
	}
	if applicant == nil {
		return false, nil
	}
	if applicant.ApplicantSurvey != nil {
		return false, nil
	}

	vacancyRec, err := i.vacancyStore.GetByID(spaceID, vacancyID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancyRec == nil || vacancyRec.HRSurvey == nil {
		return false, nil
	}

	surveyData, err := i.generateApplicantSurvey(*vacancyRec, applicant.Applicant, *vacancyRec.HRSurvey)
	if err != nil {
		return false, errors.Wrap(err, "ошибка генерации анкеты")
	}
	rec := dbmodels.ApplicantSurvey{
		BaseSpaceModel:  dbmodels.BaseSpaceModel{SpaceID: spaceID},
		VacancySurveyID: vacancyRec.HRSurvey.ID,
		ApplicantID:     applicantID,
		Survey:          dbmodels.ApplicantSurveyQuestions{Questions: surveyData.Questions},
		IsFilledOut:     false,
		HrThreshold:     vacancyRec.HRSurvey.Survey.GetThreshold(),
	}
	_, err = i.aSurveyStore.Save(rec)
	if err != nil {
		return false, errors.Wrap(err, "ошибка сохранения анкеты")
	}
	return true, nil
}

func (i impl) GetPublicApplicantSurvey(id string) (*surveyapimodels.ApplicantSurveyView, error) {
	rec, err := i.aSurveyStore.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return nil, errors.New("анкета не найдена")
	}

	result := surveyapimodels.ApplicantSurveyView{
		ApplicantSurvey: surveyapimodels.ApplicantSurvey{
			Questions: rec.Survey.Questions,
		},
		IsFilledOut: rec.IsFilledOut,
	}
	return &result, nil
}

func (i impl) AnswerPublicApplicantSurvey(id string, answers []surveyapimodels.ApplicantSurveyAnswer) (hMsg string, err error) {
	rec, err := i.aSurveyStore.GetByID(id)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения анкеты кандидата")
	}
	if rec == nil {
		return "анкета не найдена", nil
	}
	if rec.IsFilledOut {
		return "анкета уже заполнена", nil
	}

	answersMap := map[string]string{}
	for _, answer := range answers {
		answersMap[answer.QuestionID] = answer.Answer
	}
	totalAnswers := 0
	for k, question := range rec.Survey.Questions {
		selected := answersMap[question.QuestionID]
		if selected == "" {
			continue
		}
		if question.QuestionType == "single_choice" {
			found := false
			for _, answer := range question.Answers {
				if selected == answer {
					found = true
					break
				}
			}
			if !found {
				return fmt.Sprintf("для вопроса {%v} необходимо указать ответ из списка", question.QuestionText), nil
			}
		}
		totalAnswers++
		rec.Survey.Questions[k].Selected = answersMap[question.QuestionID]
	}
	if totalAnswers == len(rec.Survey.Questions) {
		rec.IsFilledOut = true
	}
	_, err = i.aSurveyStore.Save(*rec)
	if err != nil {
		return "", errors.Wrap(err, "ошибка сохранения ответов в анкету кандидата")
	}
	return "", nil
}

func (i impl) AIScore(applicantSurveyRec dbmodels.ApplicantSurvey) (ok bool, err error) {
	applicantRec, err := i.applicantStore.GetByID(applicantSurveyRec.SpaceID, applicantSurveyRec.ApplicantID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения данных кандидата")
	}
	if applicantRec == nil {
		return false, nil
	}
	if applicantRec.ApplicantSurvey == nil {
		return false, nil
	}

	vacancy, err := i.vacancyStore.GetByID(applicantSurveyRec.SpaceID, applicantRec.VacancyID)
	if err != nil {
		return false, errors.Wrap(err, "ошибка получения вакансии")
	}
	if vacancy == nil || vacancy.HRSurvey == nil {
		return false, nil
	}

	vacancyInfo, err := surveyapimodels.GetVacancyDataContent(*vacancy)
	if err != nil {
		return false, err
	}

	applicantInfo, err := surveyapimodels.GetApplicantDataContent(applicantRec.Applicant)
	if err != nil {
		return false, err
	}

	hrSurvey, err := surveyapimodels.GetHRDataContent(*vacancy.HRSurvey)
	if err != nil {
		return false, err
	}
	applicantAnswers, err := surveyapimodels.GetApplicantAnswersContent(applicantSurveyRec)
	surveyResp, err := gpthandler.Instance.ScoreApplicant(vacancy.SpaceID, vacancy.ID, vacancyInfo, applicantInfo, hrSurvey, applicantAnswers)
	if err != nil {
		return false, errors.Wrap(err, "ошибка вызова ИИ при оценке кандидата")
	}
	applicantScore := dbmodels.ScoreAI{}
	err = json.Unmarshal([]byte(surveyResp.Description), &applicantScore)
	if err != nil {
		return false, errors.Wrapf(err, "ошибка декодирования json в структуру оценки кандидата, json: %v", surveyResp.Description)
	}

	for _, score := range applicantScore.Details {
		applicantScore.Score += score.Score
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		aSurveyStore := applicantsurveystore.NewInstance(tx)
		applicantSurveyRec.ScoreAI = applicantScore
		applicantSurveyRec.IsScored = true
		_, err = aSurveyStore.Save(applicantSurveyRec)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения анкеты с оценкой")
		}
		applicantHistory := applicanthistoryhandler.NewTxHandler(tx)
		changes := dbmodels.ApplicantChanges{
			Description: fmt.Sprintf("Произведена оценка кандидата системой ИИ, оценка  %v/%v", applicantScore.Score, applicantSurveyRec.HrThreshold),
		}
		applicantHistory.Save(applicantRec.SpaceID, applicantRec.ID, vacancy.ID, "", dbmodels.HistoryAIScore, changes)

		var hMsg string
		//для ошибки изменения статуса, ошибку только логируем, чтоб оценка сохранилась и повторно не выполнялась
		if applicantScore.Score < applicantSurveyRec.HrThreshold {
			hMsg, err = applicant.Instance.UpdateStatus(applicantRec.SpaceID, applicantRec.ID, "", models.NegotiationStatusRejected)
		} else {
			hMsg, err = applicant.Instance.UpdateStatus(applicantRec.SpaceID, applicantRec.ID, "", models.NegotiationStatusAccepted)
		}
		if err != nil {
			log.
				WithError(err).
				WithField("space_id", applicantRec.SpaceID).
				WithField("applicant_id", applicantRec.ID).
				Error("ошибка перевода кандидата после оценкии системой ИИ")
		}
		if hMsg != "" {
			log.
				WithError(errors.New(hMsg)).
				WithField("space_id", applicantRec.SpaceID).
				WithField("applicant_id", applicantRec.ID).
				Error("ошибка перевода кандидата после оценкии системой ИИ")
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (i impl) generateSurvey(vacancyRec dbmodels.Vacancy) (*surveyapimodels.HRSurvey, error) {
	content, err := surveyapimodels.GetVacancyDataContent(vacancyRec)
	if err != nil {
		return nil, err
	}
	surveyResp, err := gpthandler.Instance.GenerateHRSurvey(vacancyRec.SpaceID, vacancyRec.ID, content)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка вызова ИИ при геренации анкеты")
	}
	surveyData := new(surveyapimodels.HRSurvey)
	err = json.Unmarshal([]byte(surveyResp.Description), surveyData)
	if err != nil {
		return nil, errors.Wrapf(err, "ошибка декодирования json в структуру вопросов для анкеты, json: %v", surveyResp.Description)
	}
	return surveyData, nil
}

func (i impl) regenerateSurvey(vacancyRec dbmodels.Vacancy, questions []dbmodels.HRSurveyQuestion) (*surveyapimodels.HRSurvey, error) {
	content, err := surveyapimodels.GetVacancyDataContent(vacancyRec)
	if err != nil {
		return nil, err
	}
	questionsContent, err := getQuestionPromt(questions)
	if err != nil {
		return nil, err
	}
	surveyResp, err := gpthandler.Instance.ReGenerateHRSurvey(vacancyRec.SpaceID, vacancyRec.ID, content, questionsContent)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка вызова ИИ при перегеренации анкеты")
	}
	surveyData := new(surveyapimodels.HRSurvey)
	err = json.Unmarshal([]byte(surveyResp.Description), surveyData)
	if err != nil {
		return nil, errors.Wrapf(err, "ошибка декодирования json в структуру вопросов для анкеты, json: %v", surveyResp.Description)
	}
	// берем только необходимое кол-во вопросов и проставляем им те же идентификаторы
	// если вернулось меньше вопросов чем запросили, возвращаем старые в место отсутсвующих
	result := surveyapimodels.HRSurvey{Questions: make([]dbmodels.HRSurveyQuestion, 0, len(questions))}
	newQuestions := surveyData.Questions
	for _, oldQuestion := range questions {
		var regeneratedQuestion dbmodels.HRSurveyQuestion
		regeneratedQuestion, newQuestions = getQuestion(newQuestions, oldQuestion)
		result.Questions = append(result.Questions, regeneratedQuestion)
	}
	return &result, nil
}

func getQuestion(newQuestions []dbmodels.HRSurveyQuestion, oldQuestion dbmodels.HRSurveyQuestion) (dbmodels.HRSurveyQuestion, []dbmodels.HRSurveyQuestion) {
	remainingNewQuestions := []dbmodels.HRSurveyQuestion{}
	var selectedQuestion *dbmodels.HRSurveyQuestion
	for _, newQuestion := range newQuestions {
		if selectedQuestion == nil && newQuestion.QuestionType == oldQuestion.QuestionType {
			newQuestion.QuestionID = oldQuestion.QuestionID
			selectedQuestion = &newQuestion
		}
		remainingNewQuestions = append(remainingNewQuestions, newQuestion)
	}
	if selectedQuestion == nil {
		return oldQuestion, newQuestions
	}
	return *selectedQuestion, remainingNewQuestions
}

func getQuestionPromt(questions []dbmodels.HRSurveyQuestion) (string, error) {
	regenData := surveyapimodels.ReGenerateSurvey{
		Questions: make([]surveyapimodels.NotSuitableQuestion, 0, len(questions)),
	}
	for _, question := range questions {
		regenData.Questions = append(regenData.Questions, surveyapimodels.NotSuitableQuestion{
			QuestionID:   question.QuestionID,
			QuestionText: question.QuestionText,
			QuestionType: question.QuestionType,
		})
	}
	bQuestions, err := json.Marshal(regenData)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации вопросов для перегенерации")
	}
	return string(bQuestions), nil
}

func (i impl) generateApplicantSurvey(vacancyRec dbmodels.Vacancy, applicantRec dbmodels.Applicant, hrSurveyRec dbmodels.HRSurvey) (*surveyapimodels.ApplicantSurvey, error) {
	vacancyInfo, err := surveyapimodels.GetVacancyDataContent(vacancyRec)
	if err != nil {
		return nil, err
	}

	applicantInfo, err := surveyapimodels.GetApplicantDataContent(applicantRec)
	if err != nil {
		return nil, err
	}

	hrSurvey, err := surveyapimodels.GetHRDataContent(hrSurveyRec)
	if err != nil {
		return nil, err
	}

	surveyResp, err := gpthandler.Instance.GenerateApplicantSurvey(vacancyRec.SpaceID, vacancyRec.ID, vacancyInfo, applicantInfo, hrSurvey)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка вызова ИИ при геренации анкеты для кандидата")
	}
	applicantSurvey := new(surveyapimodels.ApplicantSurvey)
	err = json.Unmarshal([]byte(surveyResp.Description), applicantSurvey)
	if err != nil {
		return nil, errors.Wrapf(err, "ошибка декодирования json в структуру вопросов для анкеты, json: %v", surveyResp.Description)
	}
	return applicantSurvey, nil
}
