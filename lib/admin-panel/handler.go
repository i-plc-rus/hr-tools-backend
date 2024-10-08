package adminpanelhandler

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/db"
	adminpaneluserstore "hr-tools-backend/lib/admin-panel/store"
	authutils "hr-tools-backend/lib/utils/auth-utils"
	adminpanelapimodels "hr-tools-backend/models/api/admin-panel"
	dbmodels "hr-tools-backend/models/db"
)

type Provider interface {
	CreateUser(request adminpanelapimodels.User) error
	UpdateUser(request adminpanelapimodels.UserUpdate) error
	DeleteUser(request adminpanelapimodels.UserID) error
	GetUser(userID string) (adminpanelapimodels.UserView, error)
	List() ([]adminpanelapimodels.UserView, error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		store: adminpaneluserstore.NewInstance(db.DB),
	}
}

type impl struct {
	store adminpaneluserstore.Provider
}

func (i impl) CreateUser(request adminpanelapimodels.User) error {
	rec := dbmodels.AdminPanelUser{
		IsActive:    true,
		Role:        request.Role,
		Password:    authutils.GetMD5Hash(request.Password),
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		Email:       request.Email,
		PhoneNumber: request.PhoneNumber,
	}
	userID, err := i.store.Create(rec)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("Ошибка создания пользователя админки")
		return err
	}
	log.
		WithField("user_id", userID).
		WithField("email", rec.Email).
		Info("Создан пользователь админки")
	return nil
}

func (i impl) UpdateUser(request adminpanelapimodels.UserUpdate) error {
	updMap := map[string]interface{}{}
	if request.Role != nil {
		updMap["Role"] = *request.Role
	}
	if request.FirstName != nil {
		updMap["FirstName"] = *request.FirstName
	}
	if request.LastName != nil {
		updMap["LastName"] = *request.LastName
	}
	if request.Password != nil {
		updMap["Password"] = authutils.GetMD5Hash(*request.Password)
	}
	if request.Email != nil {
		updMap["Email"] = *request.Email
	}
	if request.PhoneNumber != nil {
		updMap["PhoneNumber"] = *request.PhoneNumber
	}
	if request.IsActive != nil {
		updMap["IsActive"] = *request.IsActive
	}
	err := i.store.Update(request.ID, updMap)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("Ошибка обновления пользователя админки")
		return err
	}
	log.
		WithField("user_id", request.ID).
		Info("Обновлен пользователь админки")
	return nil
}

func (i impl) DeleteUser(request adminpanelapimodels.UserID) error {
	err := i.store.Delete(request.ID)
	if err != nil {
		log.
			WithField("request", fmt.Sprintf("%+v", request)).
			WithError(err).
			Error("Ошибка удаления пользователя админки")
		return err
	}
	log.
		WithField("user_id", request.ID).
		Info("Удален пользователь админки")
	return nil
}

func (i impl) GetUser(userID string) (adminpanelapimodels.UserView, error) {
	rec, err := i.store.GetByID(userID)
	if err != nil {
		log.
			WithField("userID", userID).
			WithError(err).
			Error("Ошибка получения пользователя админки")
		return adminpanelapimodels.UserView{}, err
	}
	if rec == nil {
		return adminpanelapimodels.UserView{}, errors.New("пользователь не найден")
	}
	return adminpanelapimodels.UserConvert(*rec), nil
}

func (i impl) List() ([]adminpanelapimodels.UserView, error) {
	list, err := i.store.List()
	if err != nil {
		log.
			WithError(err).
			Error("Ошибка получения списка пользователей админки")
		return nil, err
	}
	result := make([]adminpanelapimodels.UserView, 0, len(list))
	for _, rec := range list {
		result = append(result, adminpanelapimodels.UserConvert(rec))
	}
	return result, nil
}
