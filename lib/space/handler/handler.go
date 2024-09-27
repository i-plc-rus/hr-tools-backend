package spacehandler

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	spacestore "hr-tools-backend/lib/space/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	CreateOrganizationSpace(request spaceapimodels.CreateOrganization) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceStore:     spacestore.NewInstance(db.DB),
		spaceUserStore: spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	spaceStore     spacestore.Provider
	spaceUserStore spaceusersstore.Provider
}

func (i impl) CreateOrganizationSpace(request spaceapimodels.CreateOrganization) error {
	space := dbmodels.Space{
		TypeBilling:      "",
		IsActive:         true,
		OrganizationType: request.OrganizationType,
		OrganizationName: request.OrganizationName,
		Inn:              request.Inn,
		Kpp:              request.Kpp,
		OGRN:             request.OGRN,
		FullName:         request.FullName,
		DirectorName:     request.DirectorName,
	}
	spaceID, err := i.spaceStore.CreateSpace(space)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("Ошибка создания организации в space")
		return err
	}
	admin := dbmodels.SpaceUser{
		Password:    authutils.GetMD5Hash(request.AdminData.Password),
		FirstName:   request.AdminData.FirstName,
		LastName:    request.AdminData.LastName,
		IsAdmin:     true,
		Email:       request.AdminData.Email,
		IsActive:    true,
		PhoneNumber: request.AdminData.PhoneNumber,
		SpaceID:     spaceID,
	}
	err = i.spaceUserStore.Create(admin)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("Ошибка создания администратора в space")
		err = i.spaceStore.DeleteSpace(spaceID)
		if err != nil {
			log.
				WithField("request", fmt.Sprintf("%+v", request)).
				WithError(err).
				Error("Ошибка очистки space после неудачного создания администратора")
		}
		return err
	}
	return nil
}
