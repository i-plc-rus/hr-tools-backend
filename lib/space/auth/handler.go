package spaceauthhandler

import (
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	emailverify "hr-tools-backend/lib/email-verify"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
)

type Provider interface {
	SendEmailConfirmation(email string) error
	VerifyEmail(code string) error
	CheckEmail(email string) (bool, error)
}

var Instance Provider

func NewHandler() {
	Instance = impl{
		emailVerify:     emailverify.NewInstance(config.Conf.Smtp.EmailSendVerification),
		spaceUsersStore: spaceusersstore.NewInstance(db.DB),
	}
}

type impl struct {
	emailVerify     emailverify.Provider
	spaceUsersStore spaceusersstore.Provider
}

func (i impl) CheckEmail(email string) (passed bool, err error) {
	exist, err := i.spaceUsersStore.ExistByEmail(email)
	if err != nil {
		return false, err
	}
	return !exist, nil
}

func (i impl) VerifyEmail(code string) error {
	err := i.emailVerify.VerifyCode(code)
	if err != nil {
		return err
	}
	return nil
}

func (i impl) SendEmailConfirmation(email string) error {
	err := i.emailVerify.SendVerifyCode(email)
	if err != nil {
		return err
	}
	return nil
}
