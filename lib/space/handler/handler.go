package spacehandler

import (
	"context"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	companystructload "hr-tools-backend/lib/company-struct-load"
	filestorage "hr-tools-backend/lib/file-storage"
	licensestore "hr-tools-backend/lib/licence/store"
	messagetemplate "hr-tools-backend/lib/message-template"
	"hr-tools-backend/lib/smtp"
	spacesettingsstore "hr-tools-backend/lib/space/settings/store"
	spacestore "hr-tools-backend/lib/space/store"
	spaceusershander "hr-tools-backend/lib/space/users/hander"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Provider interface {
	CreateOrganizationSpace(request spaceapimodels.CreateOrganization) error
	GetProfile(spaceID string) (spaceapimodels.ProfileData, error)
	UpdateProfile(spaceID string, data spaceapimodels.ProfileData) error
	SendLicenseRequest(spaceID, userID, text string) (hMsg string, err error)
}

var Instance Provider

func NewHandler(salesEmail string) {
	Instance = impl{
		spaceStore:     spacestore.NewInstance(db.DB),
		spaceUserStore: spaceusersstore.NewInstance(db.DB),
		salesEmail:     salesEmail,
	}
}

type impl struct {
	spaceStore     spacestore.Provider
	spaceUserStore spaceusersstore.Provider
	salesEmail     string
}

func (i impl) CreateOrganizationSpace(request spaceapimodels.CreateOrganization) error {
	var spaceID string
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		// создаем пространство для организации
		spaceID, err = i.createSpace(tx, request)
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
		// создаем лицензию
		err = i.addLicense(tx, spaceID)
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

func (i impl) SendLicenseRequest(spaceID, userID, text string) (hMsg string, err error) {
	spaceRec, err := i.spaceStore.GetByID(spaceID)
	if err != nil {
		return "", err
	}
	if spaceRec == nil {
		return "Профиль организации не найден", nil
	}
	userRec, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		return "", err
	}
	if userRec == nil {
		return "Профиль пользователя не найден", nil
	}
	email, err := messagetemplate.Instance.GetSenderEmail(spaceID)
	if err != nil {
		return "", err
	}
	if email == "" {
		return "в настройках пространства не указана почта для отправки", nil
	}
	msg, err := messagetemplate.BuildLicenceRenewMsg(text, *userRec, *spaceRec)
	if err != nil {
		return "", err
	}
	title := messagetemplate.GetLicenceRenewTitle()
	err = smtp.Instance.SendHtmlEMail(email, i.salesEmail, msg, title, nil)
	if err != nil {
		return "", errors.Wrap(err, "ошибка отправки почты в отдел продаж")
	}
	return "", nil
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
	setting := dbmodels.SpaceSetting{
		SpaceID: spaceID,
		Name:    "Инструкции для YandexGPT",
		Code:    models.YandexGPTPromtSetting,
		Value:   "",
	}
	err := store.Create(setting)
	if err != nil {
		return errors.Wrap(err, "ошибка добавления настройки YandexGPT")
	}

	for code, spaceSettingData := range dbmodels.DefaultSettinsMap {
		spaceSettingData.SpaceID = spaceID
		err = store.Create(spaceSettingData)
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

func (i impl) addLicense(tx *gorm.DB, spaceID string) error {
	now := time.Now()
	endAt := now.Add(time.Hour * 24 * 7)
	plan := config.Conf.Sales.DefaultPlan
	rec := dbmodels.License{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: spaceID,
			},
			Status:    models.LicenseStatusActive,
			StartsAt:  &now,
			EndsAt:    &endAt,
			Plan:      plan,
			AutoRenew: false,
		}
	_, err := licensestore.NewInstance(tx).Create(rec)
	if err != nil {
		return errors.Wrap(err, "Ошибка добавления лицензии для организации")
	}
	return nil
}
