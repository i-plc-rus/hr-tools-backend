package spaceusershander

import (
	"fmt"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/smtp"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	pushsettingsstore "hr-tools-backend/lib/space/push/settings-store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	CreateUser(request spaceapimodels.CreateUser) error
	UpdateUser(userID string, request spaceapimodels.UpdateUser) error
	DeleteUser(userID string) error
	GetListUsers(spaceID string, page, limit int) (usersList []spaceapimodels.SpaceUser, err error)
	GetByID(userID string) (user spaceapimodels.SpaceUser, err error)
	UpdateUserProfile(userID string, request spaceapimodels.SpaceUserProfileData) error
	GetProfileByID(userID string) (user spaceapimodels.SpaceUserProfileView, err error)
	СhangePassword(userID string, payload spaceapimodels.PasswordChange) (nMsg string, err error)
	GetPushSettings(spaceID, userID string) (spaceapimodels.PushSettings, error)
	UpdatePushSettings(spaceID, userID string, payload spaceapimodels.PushSettingData) error
	UpdatePushEnable(userID string, enabled bool) error
	CreatePushSettings(tx *gorm.DB, spaceID, userID string) error
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceUserStore:    spaceusersstore.NewInstance(db.DB),
		pushSettingsStore: pushsettingsstore.NewInstance(db.DB),
	}
}

type impl struct {
	spaceUserStore    spaceusersstore.Provider
	pushSettingsStore pushsettingsstore.Provider
}

func (i impl) GetByID(userID string) (user spaceapimodels.SpaceUser, err error) {
	userDB, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		log.
			WithField("user_id", userID).
			WithError(err).
			Error("ошибка поиска пользователя")
		return spaceapimodels.SpaceUser{}, err
	}
	if userDB == nil {
		return spaceapimodels.SpaceUser{}, errors.New("пользователь не найден")
	}
	return userDB.ToModel(), nil
}

func (i impl) CreateUser(request spaceapimodels.CreateUser) error {
	userExist, err := i.spaceUserStore.ExistByEmail(request.Email)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("ошибка проверки уже существующего пользователя space")
		return err
	}
	if userExist {
		return errors.New("пользователь с такой почтой уже существует")
	}
	rec := dbmodels.SpaceUser{
		Password:        authutils.GetMD5Hash(request.Password),
		FirstName:       request.FirstName,
		LastName:        request.LastName,
		Email:           request.Email,
		IsActive:        true,
		PhoneNumber:     request.PhoneNumber,
		SpaceID:         request.SpaceID,
		TextSign:        request.TextSign,
		IsEmailVerified: !smtp.Instance.IsConfigured(),
	}
	if request.JobTitleID != "" {
		rec.JobTitleID = &request.JobTitleID
	}
	if request.IsAdmin {
		rec.Role = models.SpaceAdminRole
	} else {
		rec.Role = models.SpaceUserRole
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		spaceUserStore := spaceusersstore.NewInstance(db.DB)
		id, err := spaceUserStore.Create(rec)
		if err != nil {
			return errors.Wrap(err, "ошибка создания пользователя")
		}
		err = i.CreatePushSettings(tx, rec.SpaceID, id)
		if err != nil {
			return errors.Wrap(err, "ошибка создания списка настроек пушей для пользователя")
		}
		return nil
	})
	return nil
}

func (i impl) UpdateUser(userID string, request spaceapimodels.UpdateUser) error {
	user, err := i.GetByID(userID)
	if err != nil {
		return err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		updMap := map[string]interface{}{
			"first_name":   request.FirstName,
			"last_name":    request.LastName,
			"is_admin":     request.IsAdmin,
			"password":     authutils.GetMD5Hash(request.Password),
			"phone_number": request.PhoneNumber,
			"text_sign":    request.TextSign,
		}
		if request.JobTitleID != "" {
			updMap["job_title_id"] = request.JobTitleID
		}
		isEmailChanged := user.Email != request.Email
		if isEmailChanged {
			if smtp.Instance.IsConfigured() {
				updMap["new_email"] = request.Email
			} else {
				updMap["email"] = request.Email
				updMap["is_email_verified"] = true
			}
		}
		spaceUserStore := spaceusersstore.NewInstance(tx)
		err := spaceUserStore.Update(userID, updMap)
		if err != nil {
			log.
				WithField("request", fmt.Sprintf("%+v", request)).
				WithError(err).
				Error("ошибка обновления пользователя space")
			return err
		}

		if isEmailChanged && smtp.Instance.IsConfigured() {
			// при смене мыла, отправляем подтверждение
			err := spaceauthhandler.Instance.SendEmailConfirmation(request.Email)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (i impl) DeleteUser(userID string) error {
	err := i.spaceUserStore.Delete(userID)
	if err != nil {
		log.
			WithField("user_id", userID).
			WithError(err).
			Error("ошибка удаления пользователя space")
		return err
	}
	return nil
}

func (i impl) GetListUsers(spaceID string, page, limit int) (usersList []spaceapimodels.SpaceUser, err error) {
	list, err := i.spaceUserStore.GetList(spaceID, page, limit)
	if err != nil {
		log.WithField("space_id", spaceID).WithError(err).Error("ошибка получения списка пользователей space")
		return nil, err
	}
	for _, user := range list {
		usersList = append(usersList, user.ToModel())
	}
	return usersList, nil
}

func (i impl) UpdateUserProfile(userID string, request spaceapimodels.SpaceUserProfileData) error {
	user, err := i.GetByID(userID)
	if err != nil {
		return err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		updMap := map[string]interface{}{
			"first_name":            request.FirstName,
			"last_name":             request.LastName,
			"phone_number":          request.PhoneNumber,
			"internal_phone_number": request.InternalPhoneNumber,
			"text_sign":             request.TextSign,
			"use_personal_sign":     request.UsePersonalSign,
		}
		isEmailChanged := user.Email != request.Email
		if isEmailChanged {
			if smtp.Instance.IsConfigured() {
				updMap["new_email"] = request.Email
			} else {
				updMap["email"] = request.Email
				updMap["is_email_verified"] = true
			}
		}
		spaceUserStore := spaceusersstore.NewInstance(tx)
		err := spaceUserStore.Update(userID, updMap)
		if err != nil {
			return err
		}

		if isEmailChanged && smtp.Instance.IsConfigured() {
			// при смене мыла, отправляем подтверждение
			err := spaceauthhandler.Instance.SendEmailConfirmation(request.Email)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (i impl) GetProfileByID(userID string) (user spaceapimodels.SpaceUserProfileView, err error) {
	userDB, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		return spaceapimodels.SpaceUserProfileView{}, err
	}
	if userDB == nil {
		return spaceapimodels.SpaceUserProfileView{}, errors.New("пользователь не найден")
	}
	return userDB.ToProfile(), nil
}

func (i impl) СhangePassword(userID string, payload spaceapimodels.PasswordChange) (nMsg string, err error) {
	userDB, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		return "", err
	}
	if userDB == nil {
		return "", errors.New("пользователь не найден")
	}
	if userDB.Password != authutils.GetMD5Hash(payload.CurrentPassword) {
		return "Текущий пароль указан не верно", nil
	}
	updMap := map[string]interface{}{
		"password": authutils.GetMD5Hash(payload.NewPassword),
	}
	err = i.spaceUserStore.Update(userID, updMap)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (i impl) GetPushSettings(spaceID, userID string) (data spaceapimodels.PushSettings, err error) {
	userDB, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		return spaceapimodels.PushSettings{}, err
	}
	if userDB == nil {
		return spaceapimodels.PushSettings{}, errors.New("пользователь не найден")
	}
	list, err := i.pushSettingsStore.List(spaceID, userID)
	if err != nil {
		return spaceapimodels.PushSettings{}, err
	}
	settingsList := make([]spaceapimodels.PushSettingView, 0, len(list))
	for _, rec := range list {
		settingsList = append(settingsList, rec.ToModelView())
	}

	data = spaceapimodels.PushSettings{
		IsActive: userDB.PushEnabled,
		Settings: settingsList,
	}

	return data, nil
}

func (i impl) UpdatePushSettings(spaceID, userID string, payload spaceapimodels.PushSettingData) error {
	updMap := map[string]interface{}{
		"system_value": payload.Value.System,
		"email_value":  payload.Value.Email,
		"tg_value":     payload.Value.Tg,
	}
	return i.pushSettingsStore.Update(spaceID, userID, payload.Code, updMap)
}

func (i impl) UpdatePushEnable(userID string, enabled bool) error {
	updMap := map[string]interface{}{
		"push_enabled": enabled,
	}
	return i.spaceUserStore.Update(userID, updMap)
}

func (i impl) CreatePushSettings(tx *gorm.DB, spaceID, userID string) error {
	store := pushsettingsstore.NewInstance(tx)
	value := false
	rec := dbmodels.SpacePushSetting{
		BaseSpaceModel: dbmodels.BaseSpaceModel{
			SpaceID: spaceID,
		},
		SpaceUserID: userID,
		SystemValue: &value,
		EmailValue:  &value,
		TgValue:     &value,
	}

	for key, _ := range models.PushCodeMap {
		rec.Code = key
		err := store.Create(rec)
		if err != nil {
			return err
		}
	}
	return nil
}
