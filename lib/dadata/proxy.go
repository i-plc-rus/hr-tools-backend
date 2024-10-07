package dadataproxy

import (
	"context"
	dadata "github.com/ekomobile/dadata/v2"
	"github.com/ekomobile/dadata/v2/api/suggest"
)

func ProxySuggestRequest(query string) (ret []*suggest.PartySuggestion, err error) {
	api := dadata.NewSuggestApi()
	params := suggest.RequestParams{Query: query}
	return api.Party(context.Background(), &params)
}
