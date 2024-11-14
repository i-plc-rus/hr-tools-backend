package spacehandler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"hr-tools-backend/db"
	companystructload "hr-tools-backend/lib/company-struct-load"
	filestorage "hr-tools-backend/lib/file-storage"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spacestore "hr-tools-backend/lib/space/store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	CreateOrganizationSpace(request spaceapimodels.CreateOrganization) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{}
}

type impl struct {
}

func (i impl) CreateOrganizationSpace(request spaceapimodels.CreateOrganization) error {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// создаем пространство для организации
		spaceID, err := i.createSpace(tx, request)
		if err != nil {
			return err
		}
		// создаем главного админа для пространства
		err = i.createAdmin(tx, spaceID, request.AdminData)
		if err != nil {
			return err
		}
		// создаем настройки по-умолчанию для простраства
		err = i.createDefaultSpaceSettings(tx, spaceID)
		if err != nil {
			return err
		}
		// подгружаем справочники доступные по-умолчанию
		err = i.preloadDicts(tx, spaceID)
		if err != nil {
			return err
		}
		// создаем отдельный бакет в S3
		err = i.makeS3Bucket(context.Background(), spaceID)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("Ошибка создания организации в space")
		return err
	}

	return nil
}

func (i impl) createSpace(tx *gorm.DB, request spaceapimodels.CreateOrganization) (spaceID string, err error) {
	space := dbmodels.Space{
		TypeBilling:      "",
		IsActive:         true,
		OrganizationName: request.OrganizationName,
		Inn:              request.Inn,
		Kpp:              request.Kpp,
		OGRN:             request.OGRN,
		FullName:         request.FullName,
		DirectorName:     request.DirectorName,
	}
	spaceID, err = spacestore.NewInstance(tx).CreateSpace(space)
	if err != nil {
		return "", errors.Wrap(err, "Ошибка создания организации в space")
	}
	return spaceID, nil
}

func (i impl) createAdmin(tx *gorm.DB, spaceID string, adminData spaceapimodels.CreateUser) error {
	admin := dbmodels.SpaceUser{
		Password:    authutils.GetMD5Hash(adminData.Password),
		FirstName:   adminData.FirstName,
		LastName:    adminData.LastName,
		Role:        models.SpaceAdminRole,
		Email:       adminData.Email,
		IsActive:    true,
		PhoneNumber: adminData.PhoneNumber,
		SpaceID:     spaceID,
	}
	err := spaceusersstore.NewInstance(tx).Create(admin)
	if err != nil {
		return errors.Wrap(err, "Ошибка создания администратора в space")
	}
	return nil
}

func (i impl) createDefaultSpaceSettings(tx *gorm.DB, spaceID string) error {
	yaGPTSetting := dbmodels.SpaceSetting{
		SpaceID: spaceID,
		Name:    "Инструкции для YandexGPT",
		Code:    models.YandexGPTPromtSetting,
		Value:   "",
	}
	err := spacesettingsstore.NewInstance(tx).Create(yaGPTSetting)
	if err != nil {
		return errors.Wrap(err, "ошибка добавления настройки YandexGPT")
	}
	return nil
}

func (i impl) preloadDicts(tx *gorm.DB, spaceID string) error {
	return companystructload.PreloadCompanyStruct(tx, spaceID)
}

func (i impl) makeS3Bucket(ctx context.Context, spaceID string) error {
	err := filestorage.Instance.MakeSpaceBucket(ctx, spaceID)
	if err != nil {
		return errors.Wrap(err, "ошибка создания бакета для space")
	}
	return nil
}
