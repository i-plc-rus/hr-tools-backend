package survey

import (
	"encoding/json"
	"hr-tools-backend/db"
	gpthandler "hr-tools-backend/lib/gpt"
	vacancysurvaystore "hr-tools-backend/lib/survey/vacancy-survay-store"
	vacancystore "hr-tools-backend/lib/vacancy/store"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"
	"sort"

	"github.com/pkg/errors"
)

type Provider interface {
	SaveHRSurvey(spaceID, vacancyID string, survay surveyapimodels.HRSurvey) (*surveyapimodels.HRSurveyView, error)
	GetHRSurvey(spaceID, vacancyID string) (*surveyapimodels.HRSurveyView, error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		vSurvayStore: vacancysurvaystore.NewInstance(db.DB),
		vacancyStore: vacancystore.NewInstance(db.DB),
	}
}

type impl struct {
	vSurvayStore vacancysurvaystore.Provider
	vacancyStore vacancystore.Provider
}

func (i impl) SaveHRSurvey(spaceID, vacancyID string, survay surveyapimodels.HRSurvey) (*surveyapimodels.HRSurveyView, error) {
	rec := dbmodels.HRSurvey{
		BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
		IsFilledOut:    false,
		VacancyID:      vacancyID,
		Survey: dbmodels.HRSurveyQuestions{
			Questions: make([]dbmodels.HRSurveyQuestion, 0, len(survay.Questions)),
		},
	}
	regenerateQuestions := []dbmodels.HRSurveyQuestion{}
	selectedCount := 0
	for _, question := range survay.Questions {
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

	_, err := i.vSurvayStore.Save(rec)
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
	rec, err := i.vSurvayStore.GetByVacancyID(spaceID, vacancyID)
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

func (i impl) generateSurvey(vacancyRec dbmodels.Vacancy) (*surveyapimodels.HRSurvey, error) {
	content, err := surveyapimodels.GetVacancyDataContent(vacancyRec)
	if err != nil {
		return nil, err
	}
	surveyResp, err := gpthandler.Instance.GenerateHRSurvey(vacancyRec.SpaceID, content)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка вызова ИИ при геренации анкеты")
	}
	surveyData := new(surveyapimodels.HRSurvey)
	err = json.Unmarshal([]byte(surveyResp.Description), surveyData)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка декодирования json в структуру вопросов для анкеты")
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
	surveyResp, err := gpthandler.Instance.ReGenerateHRSurvey(vacancyRec.SpaceID, content, questionsContent)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка вызова ИИ при перегеренации анкеты")
	}
	surveyData := new(surveyapimodels.HRSurvey)
	err = json.Unmarshal([]byte(surveyResp.Description), surveyData)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка декодирования json в структуру вопросов для анкеты")
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
