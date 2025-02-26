package survey

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	SaveSurvey(spaceID, vacancyID string, survay surveyapimodels.Survey) (*surveyapimodels.SurveyView, error)
	GetSurvey(spaceID, vacancyID string) (*surveyapimodels.SurveyView, error)
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

func (i impl) SaveSurvey(spaceID, vacancyID string, survay surveyapimodels.Survey) (*surveyapimodels.SurveyView, error) {
	rec := dbmodels.VacancySurvey{
		BaseSpaceModel: dbmodels.BaseSpaceModel{SpaceID: spaceID},
		IsFilledOut:    false,
		VacancyID:      vacancyID,
		Survey: dbmodels.SurveyQuestions{
			Questions: make([]dbmodels.SurveyQuestion, 0, len(survay.Questions)),
		},
	}
	regenerateQuestions := []dbmodels.SurveyQuestion{}
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
	result := surveyapimodels.SurveyView{
		IsFilledOut: rec.IsFilledOut,
		Survey: surveyapimodels.Survey{
			Questions: rec.Survey.Questions,
		},
	}
	return &result, nil
}

func (i impl) GetSurvey(spaceID, vacancyID string) (*surveyapimodels.SurveyView, error) {
	rec, err := i.vSurvayStore.GetByVacancyID(spaceID, vacancyID)
	if err != nil {
		return nil, err
	}
	if rec != nil {
		result := surveyapimodels.SurveyView{
			IsFilledOut: rec.IsFilledOut,
			Survey: surveyapimodels.Survey{
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
	return i.SaveSurvey(spaceID, vacancyID, *surveyData)
}

func (i impl) generateSurvey(vacancyRec dbmodels.Vacancy) (*surveyapimodels.Survey, error) {
	content := buildGeneratePromt(vacancyRec)
	surveyResp, err := gpthandler.Instance.Gen(vacancyRec.SpaceID, content) //TODO новый метод
	if err != nil {
		return nil, errors.Wrap(err, "ошибка вызова ИИ при геренации анкеты")
	}
	surveyData := new(surveyapimodels.Survey)
	err = json.Unmarshal([]byte(surveyResp.Description), surveyData)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка декодирования json в структуру вопросов для анкеты")
	}
	return surveyData, nil
}

func (i impl) regenerateSurvey(vacancyRec dbmodels.Vacancy, questions []dbmodels.SurveyQuestion) (*surveyapimodels.Survey, error) {
	content, err := buildReGeneratePromt(vacancyRec, questions)
	if err != nil {
		return nil, err
	}
	surveyResp, err := gpthandler.Instance.ReGen(vacancyRec.SpaceID, content) //TODO новый метод
	if err != nil {
		return nil, errors.Wrap(err, "ошибка вызова ИИ при перегеренации анкеты")
	}
	surveyData := new(surveyapimodels.Survey)
	err = json.Unmarshal([]byte(surveyResp.Description), surveyData)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка декодирования json в структуру вопросов для анкеты")
	}
	// берем только необходимое кол-во вопросов и проставляем им те же идентификаторы
	// если вернулось меньше вопросов чем запросили, возвращаем старые в место отсутсвующих
	result := surveyapimodels.Survey{Questions: make([]dbmodels.SurveyQuestion, 0, len(questions))}
	newQuestionsCount := len(surveyData.Questions)
	for j, question := range questions {
		if newQuestionsCount < j+1 {
			result.Questions = append(result.Questions, question)
		} else {
			newQuestion := surveyData.Questions[j]
			resultQuestion := dbmodels.SurveyQuestion{
				SurveyQuestionGenerated: dbmodels.SurveyQuestionGenerated{
					QuestionID:   question.QuestionID,
					QuestionText: newQuestion.QuestionText,
					QuestionType: newQuestion.QuestionType,
					Answers:      newQuestion.Answers,
					Comment:      newQuestion.Comment,
				},
			}
			result.Questions = append(result.Questions, resultQuestion)
		}
	}
	return &result, nil
}

func buildGeneratePromt(rec dbmodels.Vacancy) string {
	// формируем текст:
	// «Менеджер по продажам в IT. Основные обязанности: холодные b2b-продажи, ведение переговоров, заключение сделок. Требования: опыт более 2 лет, умение работать в CRM. Желательна специализация в IT-сфере.»\nНужно:\n
	// 1. Сгенерировать 5 вопросов с вариантами ответов, чтобы интервьюер мог уточнить важные аспекты вакансии.\n
	return fmt.Sprintf("\"%v\" %v\"\n1. Сгенерировать 5 вопросов с вариантами ответов, чтобы интервьюер мог уточнить важные аспекты вакансии.\n", rec.VacancyName, rec.Requirements)
}

func buildReGeneratePromt(rec dbmodels.Vacancy, questions []dbmodels.SurveyQuestion) (string, error) {
	// формируем текст:
	// «Менеджер по продажам в IT. Основные обязанности: холодные b2b-продажи, ведение переговоров, заключение сделок. Требования: опыт более 2 лет, умение работать в CRM. Желательна специализация в IT-сфере.»\n
	// 1. Вопрос "Какие навыки продаж вы считаете ключевыми?" не подошёл. Сгенерируй новый вопрос с типом "free_text", добавь комментарий и опцию "Не подходит".\n
	// если несколько вопросов:
	// 1. Вопросы questions": [...] не подошли. Сгенерируй новые вопросы с аналогичными типами, добавь комментарий и опцию "Не подходит".\n
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("\"%v\" %v\"\n", rec.VacancyName, rec.Requirements))
	if len(questions) == 1 {
		question := questions[0]
		buffer.WriteString(fmt.Sprintf("1. Вопрос \"%v\" не подошёл. Сгенерируй новый вопрос с типом \"%v\", добавь комментарий и опцию \"Не подходит\".\n",
			question.QuestionText, question.QuestionType))
	} else {
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
		buffer.WriteString(fmt.Sprintf("1. Вопросы\n%v\nне подошли. Сгенерируй новые вопросы с аналогичными типами. \n", string(bQuestions)))
	}
	return buffer.String(), nil
}
