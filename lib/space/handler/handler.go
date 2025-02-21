package spacehandler

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"hr-tools-backend/db"
	companystructload "hr-tools-backend/lib/company-struct-load"
	filestorage "hr-tools-backend/lib/file-storage"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spacestore "hr-tools-backend/lib/space/store"
	spaceusershander "hr-tools-backend/lib/space/users/hander"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	CreateOrganizationSpace(request spaceapimodels.CreateOrganization) error
	GetProfile(spaceID string) (spaceapimodels.ProfileData, error)
	UpdateProfile(spaceID string, data spaceapimodels.ProfileData) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceStore: spacestore.NewInstance(db.DB),
	}
}

type impl struct {
	spaceStore spacestore.Provider
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
		return err
	}

	return nil
}

func (i impl) GetProfile(spaceID string) (spaceapimodels.ProfileData, error) {
	rec, err := i.spaceStore.GetByID(spaceID)
	if err != nil {
		return spaceapimodels.ProfileData{}, err
	}
	if rec == nil {
		return spaceapimodels.ProfileData{}, errors.New("Профиль организации не найден")
	}

	return rec.ToModel(), nil
}

func (i impl) UpdateProfile(spaceID string, data spaceapimodels.ProfileData) error {
	updMap := map[string]interface{}{
		"organization_name": data.OrganizationName,
		"web":               data.Web,
		"time_zone":         data.TimeZone,
		"description":       data.Description,
		"director_name":     data.DirectorName,
	}

	err := i.spaceStore.UpdateSpace(spaceID, updMap)
	if err != nil {
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
	id, err := spaceusersstore.NewInstance(tx).Create(admin)
	if err != nil {
		return errors.Wrap(err, "Ошибка создания администратора в space")
	}
	err = spaceusershander.Instance.CreatePushSettings(tx, admin.SpaceID, id)
	if err != nil {
		return errors.Wrap(err, "ошибка создания списка настроек пушей для пользователя")
	}
	return nil
}

func (i impl) createDefaultSpaceSettings(tx *gorm.DB, spaceID string) error {
	store := spacesettingsstore.NewInstance(tx)
	for code, spaceSettingData := range dbmodels.DefaultSettinsMap {
		spaceSettingData.SpaceID = spaceID
		err := store.Create(spaceSettingData)
		if err != nil {
			return errors.Wrapf(err, "ошибка добавления настройки %v", code)
		}
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
