package hhhandler

import (
	"context"
	"github.com/stretchr/testify/require"
	"hr-tools-backend/lib/external-services/hh/hhclient"
	dbmodels "hr-tools-backend/models/db"
	"testing"
)

func TestGeoJSONHelpers(t *testing.T) {
	t.Run(`getArea check`, func(t *testing.T) {
		i := getInstance()
		city := dbmodels.City{}
		city.ID = "7baaac39-aa31-4712-893d-51ef80124a86"
		city.City = "Сергиев Посад"
		city.Region = "Московская"
		area, err := i.getArea(context.TODO(), &city)
		require.Nil(t, err)
		require.Equal(t, "2069", area.ID)

		city.ID = "0c5b2444-70a0-4932-980c-b4dc0d3f02b5"
		city.City = "Москва"
		city.Region = "Москва"
		area, err = i.getArea(context.TODO(), &city)
		require.Nil(t, err)
		require.Equal(t, "1", area.ID)
	})
}

func getInstance() impl {
	hhclient.NewProvider("https://a.hr-tools.pro/api/v1/oauth/callback/hh")
	return impl{
		client:  hhclient.Instance,
		cityMap: map[string]string{},
	}
}
