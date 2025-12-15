package spaceusershander

import (
	"hr-tools-backend/db"
	"hr-tools-backend/lib/smtp"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	pushsettingsstore "hr-tools-backend/lib/space/push/settings-store"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authhelpers "hr-tools-backend/lib/utils/auth-helpers"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
	"sort"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Provider interface {
	CreateUser(request spaceapimodels.CreateUser, authorSpaceID string) (id, hMsg string, err error)
	UpdateUser(userID string, request spaceapimodels.UpdateUser) error
	UpdateUserStatus(userID string, request spaceapimodels.UpdateUserStatus) (user spaceapimodels.SpaceUser, err error)
	DeleteUser(userID string) error
	GetListUsers(spaceID string, filter spaceapimodels.SpaceUserFilter) (usersList []spaceapimodels.SpaceUser, rowCount int64, err error)
	GetByID(userID string) (user spaceapimodels.SpaceUser, err error)
	UpdateUserProfile(userID string, request spaceapimodels.SpaceUserProfileData) error
	GetProfileByID(userID string) (user spaceapimodels.SpaceUserProfileView, err error)
	ChangePassword(userID string, payload spaceapimodels.PasswordChange) (nMsg string, err error)
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
		return spaceapimodels.SpaceUser{}, err
	}
	if userDB == nil {
		return spaceapimodels.SpaceUser{}, errors.New("пользователь не найден")
	}
	return userDB.ToModel(), nil
}

func (i impl) CreateUser(request spaceapimodels.CreateUser, authorSpaceID string) (id, hMsg string, err error) {
	userExist, err := i.spaceUserStore.ExistByEmail(request.Email)
	if err != nil {
		return "", "", err
	}
	if userExist {
		return "", "пользователь с такой почтой уже существует", nil
	}
	rec := dbmodels.SpaceUser{
		Password:        authhelpers.GetMD5Hash(request.Password),
		FirstName:       request.FirstName,
		LastName:        request.LastName,
		Email:           request.Email,
		IsActive:        true,
		PhoneNumber:     request.PhoneNumber,
		SpaceID:         authorSpaceID,
		TextSign:        request.TextSign,
		IsEmailVerified: !smtp.Instance.IsConfigured(),
	}
	if request.JobTitleID != "" {
		rec.JobTitleID = &request.JobTitleID
	}

	rec.Role = models.UserRole(request.Role)

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		spaceUserStore := spaceusersstore.NewInstance(db.DB)
		id, err = spaceUserStore.Create(rec)
		if err != nil {
			return errors.Wrap(err, "ошибка создания пользователя")
		}
		err = i.CreatePushSettings(tx, rec.SpaceID, id)
		if err != nil {
			return errors.Wrap(err, "ошибка создания списка настроек пушей для пользователя")
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}
	return id, "", nil
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
			"phone_number": request.PhoneNumber,
		}
		if request.Role != nil {
			updMap["role"] = models.UserRole(*request.Role)
		}
		if request.Password != nil && *request.Password != "" {
			updMap["password"] = authhelpers.GetMD5Hash(*request.Password)
		}

		if request.JobTitleID != nil && *request.JobTitleID != "" {
			updMap["job_title_id"] = request.JobTitleID
		}
		if request.TextSign != nil {
			updMap["text_sign"] = *request.TextSign
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

func (i impl) UpdateUserStatus(userID string, request spaceapimodels.UpdateUserStatus) (user spaceapimodels.SpaceUser, err error) {
	userDB, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		return spaceapimodels.SpaceUser{}, err
	}
	if userDB == nil {
		return spaceapimodels.SpaceUser{}, errors.New("пользователь не найден")
	}

	now := time.Now()
	updMap := map[string]interface{}{
		"status":            models.UserStatus(request.Status),
		"status_changed_at": now,
		"is_active":         models.UserStatus(request.Status) != models.SpaceDismissedStatus,
	}
	if request.Comment != nil {
		updMap["status_comment"] = request.Comment
	}

	err = i.spaceUserStore.Update(userID, updMap)
	if err != nil {
		return spaceapimodels.SpaceUser{}, err
	}

	// Обновляем поля в уже полученном объекте
	userDB.Status = models.UserStatus(request.Status)
	userDB.StatusChangedAt = now
	if request.Comment != nil {
		userDB.StatusComment = request.Comment
	}
	if models.UserStatus(request.Status) == models.SpaceDismissedStatus {
		userDB.IsActive = false
	}

	return userDB.ToModel(), nil
}

func (i impl) DeleteUser(userID string) error {
	err := i.spaceUserStore.Delete(userID)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) GetListUsers(spaceID string, filter spaceapimodels.SpaceUserFilter) (usersList []spaceapimodels.SpaceUser, rowCount int64, err error) {
	rowCount, err = i.spaceUserStore.GetCountList(spaceID, filter)
	if err != nil {
		return nil, 0, err
	}

	page, limit := filter.GetPage()
	offset := (page - 1) * limit
	if int64(offset) > rowCount {
		return []spaceapimodels.SpaceUser{}, rowCount, nil
	}

	list, err := i.spaceUserStore.GetList(spaceID, filter)
	if err != nil {
		return nil, 0, err
	}
	for _, user := range list {
		usersList = append(usersList, user.ToModel())
	}
	return usersList, rowCount, nil
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

		if request.JobTitleID != nil && *request.JobTitleID != "" {
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

func (i impl) ChangePassword(userID string, payload spaceapimodels.PasswordChange) (nMsg string, err error) {
	userDB, err := i.spaceUserStore.GetByID(userID)
	if err != nil {
		return "", err
	}
	if userDB == nil {
		return "", errors.New("пользователь не найден")
	}
	if userDB.Password != authhelpers.GetMD5Hash(payload.CurrentPassword) {
		return "Текущий пароль указан не верно", nil
	}
	updMap := map[string]interface{}{
		"password": authhelpers.GetMD5Hash(payload.NewPassword),
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
	sort.Slice(settingsList, func(i, j int) bool {
		return settingsList[i].Name < settingsList[j].Name
	})

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
