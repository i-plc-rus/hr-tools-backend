package gpthandler

import (
	"encoding/json"
	"fmt"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	VkStep1SysPromt = "Ты — нейросеть, помогаешь HR-специалистам формировать сценарии для интервью."
	VkStep1Template = `{
  "vacancy":%v,
  "candidate":%v,
  "questions":%v,
  "candidate_answers":%v,
  "instruction":"Сгенерируй пул из 15 вопросов и текст сценария для интервью (intro/outro). Формат ответа: смотри в answer_format",
"answer_format":{
  "questions":[
    {"id":"q1","text":"…","type":"single_choice"}
  ],
  "script_intro":"…",
  "script_outro":"…",
  "comments":{"q1":"…"}
}
}
`

	VkStep1RegenSysPromt = "Ты — нейросеть, помогаешь HR-специалистам формировать вопросы для интервью."
	VkStep1RegenTemplate = `{
  "vacancy":%v,
  "candidate":%v,
  "questions":%v,
  "candidate_answers":%v,
  "generated_questions":%v,
  "instruction":"Есть пул из 15 вопросов смотри generated_questions. Часть вопросов не подошла, они помечены в generated_questions аттрибутом not_suitable = true. Сгенерируй новые вопросы в место тех, которые не подошли. Ответ должен содержать только новые вопросы и коментарии к ним. Формат ответа: смотри в answer_format",
"answer_format":{
  "questions":[
    {"id":"q1","text":"…","type":"single_choice"}
  ],
  "comments":{"q1":"…"}
}
}
`
)

func (i impl) VkStep1(spaceID, vacancyID string, aiData surveyapimodels.AiData) (resp surveyapimodels.VkStep1, err error) {
	/*
			Ожидаемый ответ:
		{
		    "questions": [
		        {
		            "id": "q1",
		            "text": "…",
		            "type": "single_choice"
		        },
				    …
		    ],
		    "script_intro": "…",
		    "script_outro": "…",
		    "comments": {
		        "q1": "…",…
		    }
		}
	*/
	userPromt := fmt.Sprintf(VkStep1Template, aiData.VacancyInfo, aiData.ApplicantInfo, aiData.Questions, aiData.ApplicantAnswers)
	description, err := i.getYaClient().
		GenerateByPromtAndText(VkStep1SysPromt, userPromt)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка генерация черновика скрипта через GPT")
		return resp, err
	}
	i.saveLog(spaceID, vacancyID, ApplicantScoreSysPromt, userPromt, description, dbmodels.AiScoreApplicantType)

	err = json.Unmarshal([]byte(description), &resp)
	if err != nil {
		return resp, errors.Wrapf(err, "ошибка декодирования json в структуру черновика скрипта, json: %v", description)
	}
	return resp, nil
}

func (i impl) VkStep1Regen(spaceID, vacancyID string, aiData surveyapimodels.AiData) (newQuestions []surveyapimodels.VkStep1Question, comments map[string]string, err error) {
	userPromt := fmt.Sprintf(VkStep1RegenTemplate, aiData.VacancyInfo, aiData.ApplicantInfo, aiData.Questions, aiData.ApplicantAnswers, aiData.GeneratedQuestions)
	description, err := i.getYaClient().
		GenerateByPromtAndText(VkStep1RegenSysPromt, userPromt)
	if err != nil {
		log.
			WithField("space_id", spaceID).
			WithError(err).
			Error("ошибка перегенерации вопросов для черновика скрипта через GPT")
		return nil, nil, err
	}
	i.saveLog(spaceID, vacancyID, ApplicantScoreSysPromt, userPromt, description, dbmodels.AiScoreApplicantType)

	resp := surveyapimodels.VkStep1{}
	err = json.Unmarshal([]byte(description), &resp)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "ошибка декодирования json в структуру черновика скрипта при перегенерации, json: %v", description)
	}
	return resp.Questions, resp.Comments, nil
}
