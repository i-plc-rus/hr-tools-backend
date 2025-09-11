package gpthandler

import (
	"encoding/json"
	"fmt"
	surveyapimodels "hr-tools-backend/models/api/survey"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const(
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
)
func (i impl) VkStep1(spaceID, vacancyID, vacancyInfo, applicantInfo, questions, applicantAnswers string) (resp surveyapimodels.VkStep1, err error) {
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
	userPromt := fmt.Sprintf(VkStep1Template, vacancyInfo, applicantInfo, questions, applicantAnswers)
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
