package dadataproxy

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"hr-tools-backend/config"
	"time"
)

const dadataFindPartyApiUrl = "http://suggestions.dadata.ru/suggestions/api/4_1/rs/findById/party"

type request struct {
	Query string `json:"query"`
}

// POST http://suggestions.dadata.ru/suggestions/api/4_1/rs/findById/party { "query": "7707083893" }
func ProxySuggestRequest(query string) (ret []byte, errs []error) {
	_, ret, errs = fiber.
		Post(dadataFindPartyApiUrl).
		Add(fiber.HeaderAuthorization, fmt.Sprintf("Token %s", config.Conf.DaData.ApiKey)).
		Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON).
		Add(fiber.HeaderAccept, fiber.MIMEApplicationJSON).
		Timeout(time.Second * time.Duration(config.Conf.DaData.Timeout)).
		JSON(&request{Query: query}).
		Bytes()
	if len(errs) > 0 {
		return nil, errs
	}
	return ret, nil
}
