package spaceusershander

import (
	"errors"
	"fmt"
	"hr-tools-backend/db"
	"hr-tools-backend/lib/smtp"
	spaceauthhandler "hr-tools-backend/lib/space/auth"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	"hr-tools-backend/models"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Provider interface {
	CreateUser(request spaceapimodels.CreateUser) error
	UpdateUser(userID string, request spaceapimodels.UpdateUser) error
	DeleteUser(userID string) error
	GetListUsers(spaceID string, page, limit int) (usersList []spaceapimodels.SpaceUser, err error)
	GetByID(userID string) (user spaceapimodels.SpaceUser, err error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		spaceUserStore: spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	spaceUserStore spaceusersstore.Provider
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
	if request.IsAdmin {
		rec.Role = models.SpaceAdminRole
	} else {
		rec.Role = models.SpaceUserRole
	}
	err = i.spaceUserStore.Create(rec)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("ошибка создания пользователя space")
		return err
	}
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
