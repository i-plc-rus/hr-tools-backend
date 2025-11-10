package botnotify

import (
	"fmt"
	"hr-tools-backend/config"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func SendAiResult(ai, spaceID, applicantID, errs string, logger *logrus.Entry) {
	payload := fmt.Sprintf(
		`{"ai":%q,"space_id":%q,"applicant_id":%q,"error":%q}`,
		ai, spaceID, applicantID, errs)
	resp, err := http.Post(config.Conf.NotifyBot.AddrAi, "application/json", strings.NewReader(payload))
	if err != nil {
		logger.WithError(err).Errorf("error sending error notification to telegram, resp %+v", resp)
	}
}
