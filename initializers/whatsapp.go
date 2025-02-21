package initializers

import (
	"context"
	"hr-tools-backend/config"
	whatsappclient "hr-tools-backend/lib/whatsup/client"
)

func InitWhatsupp(ctx context.Context) {
	err := whatsappclient.Connect(
		ctx,
		config.Conf.WhatsUpp.BaseUrl,
		config.Conf.WhatsUpp.AccessToken,
		config.Conf.WhatsUpp.APIVersion,
		config.Conf.WhatsUpp.BusinessAccountID,
	)
	if err != nil {
		panic(err.Error())
	}
}
