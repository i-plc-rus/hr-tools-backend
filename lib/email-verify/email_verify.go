package emailverify

import (
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	emailverifystore "hr-tools-backend/lib/email-verify/store"
	"hr-tools-backend/lib/smtp"
	spaceusersstore "hr-tools-backend/lib/space/users/store"
	dbmodels "hr-tools-backend/models/db"
	"math/rand"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/pkg/errors"
)

const daysToExpires = 14
const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

type Provider interface {
	SendVerifyCode(email string) error
	VerifyCode(code string) error
}

var Instance Provider

func NewInstance(emailFrom string) Provider {
	return &impl{
		verifyStore: emailverifystore.NewInstance(db.DB),
		emailFrom:   emailFrom,
	}
}

type impl struct {
	verifyStore emailverifystore.Provider
	emailFrom   string
}

func (i impl) SendVerifyCode(email string) error {
	exist, err := i.verifyStore.Exist(email)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("Такая почта уже существует в системе")
	}
	verifyData := dbmodels.EmailVerify{
		Email:         email,
		Code:          i.generateCode(),
		DateGenerated: time.Now(),
		DateExpires:   time.Now().Add(time.Hour * 24 * daysToExpires),
	}
	err = i.verifyStore.Create(verifyData)
	if err != nil {
		return err
	}
	// отправляем код асинхронно, тк отправка может занять до полуминуты
	go func(code string) {
		logger := log.WithField("email", email)
		message := fmt.Sprintf("Ссылка для подтверждения почты: %s/api/v1/auth/verify-email?code=%s", config.Conf.Smtp.DomainForVerifyLink, code)
		err = smtp.Instance.SendEMail(i.emailFrom, email, message, "EMail Confirm")
		if err != nil {
			// если не удалось отправить, удаляем код для возможности повторной отправки
			logger.WithError(err).Error("Ошибка отправки ссылки для подтверждения почты")
			i.verifyStore.DeleteByCode(verifyData.Code)
		}
	}(verifyData.Code)
	return nil
}

func (i impl) VerifyCode(code string) error {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		verifyStore := emailverifystore.NewInstance(tx)
		userStore := spaceusersstore.NewInstance(tx)

		email, err := applyCode(code, verifyStore)
		if err != nil {
			return err
		}
		return updateUser(email, userStore)
	})
	return err
}

func applyCode(code string, verifyStore emailverifystore.Provider) (email string, err error) {
	verifyData, err := verifyStore.GetByCode(code)
	if err != nil {
		return "", err
	}
	if verifyData == nil {
		return "", errors.New("указанный код не найден")
	}
	if !verifyData.DateUsed.IsZero() {
		return "", errors.New("указанный код уже использован")
	}
	if verifyData.DateExpires.Before(time.Now()) {
		return "", errors.New("срок указанного кода истек")
	}
	logger := log.WithField("email", verifyData.Email)

	updMap := map[string]interface{}{
		"date_used": time.Now(),
	}
	err = verifyStore.UpdateByCode(code, updMap)
	if err != nil {
		logger.WithError(err).Error("емайл не подтвержден, ошибка обновления таблицы EmailVerify")
		return "", errors.New("ошибка применения кода")
	}
	return verifyData.Email, nil
}

func updateUser(email string, userStore spaceusersstore.Provider) error {
	user, err := userStore.FindByEmail(email, true)
	if err != nil {
		return errors.Wrap(err, "ошибка получения данных пользователя")
	}
	if user == nil {
		return errors.Wrap(err, "пользователь не найден")
	}
	updMap := map[string]interface{}{
		"is_email_verified": true,
	}
	if user.NewEmail == email {
		// подтвердили новое мыло
		updMap["email"] = user.NewEmail
		updMap["new_email"] = ""
	}
	err = userStore.Update(user.ID, updMap)
	if err != nil {
		return errors.Wrap(err, "ошибка обновления емайла пользователя space")
	}
	return nil
}

func (i impl) generateCode() string {
	sb := strings.Builder{}
	sb.Grow(24)
	for i := 0; i < 24; i++ {
		idx := rand.Int63() % int64(len(letterBytes))
		sb.WriteByte(letterBytes[idx])
	}
	return sb.String()
}
