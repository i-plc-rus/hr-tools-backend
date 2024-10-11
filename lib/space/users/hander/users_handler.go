package spaceusershander

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	spaceapimodels "hr-tools-backend/models/api/space"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	CreateUser(request spaceapimodels.CreateUser) error
	UpdateUser(userID string, request spaceapimodels.UpdateUser) error
	DeleteUser(userID string) error
	GetListUsers(ctx *fiber.Ctx, page, limit int) (usersList []spaceapimodels.SpaceUser, err error)
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
		return spaceapimodels.SpaceUser{}, nil
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
		Password:    authutils.GetMD5Hash(request.Password),
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		IsAdmin:     request.IsAdmin,
		Email:       request.Email,
		IsActive:    true,
		PhoneNumber: request.PhoneNumber,
		SpaceID:     request.SpaceID,
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
	updMap := map[string]interface{}{
		"email":        request.Email,
		"first_name":   request.FirstName,
		"last_name":    request.LastName,
		"is_admin":     request.IsAdmin,
		"password":     authutils.GetMD5Hash(request.Password),
		"phone_number": request.PhoneNumber,
	}
	err := i.spaceUserStore.Update(userID, updMap)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("ошибка обновления пользователя space")
		return err
	}
	return nil
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

func (i impl) GetListUsers(ctx *fiber.Ctx, page, limit int) (usersList []spaceapimodels.SpaceUser, err error) {
	//TODO get space id from ctx (from jwt)
	spaceID := "ctx"
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
