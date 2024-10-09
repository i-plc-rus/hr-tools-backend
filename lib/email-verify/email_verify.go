package emailverify

import (
	"fmt"
	"github.com/pkg/errors"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	emailverifystore "hr-tools-backend/lib/email-verify/store"
	"hr-tools-backend/lib/smtp"
	dbmodels "hr-tools-backend/models/db"
	"math/rand"
	"strings"
	"time"
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
	message := fmt.Sprintf("Ссылка для подтверждения почты: %s/api/v1/auth/verify-email?code=%s", config.Conf.Smtp.DomainForVerifyLink, verifyData.Code)
	err = smtp.Instance.SendConfirmEMail(i.emailFrom, email, message, "EMail Confirm")
	if err != nil {
		return err
	}
	return nil
}

func (i impl) VerifyCode(code string) error {
	verifyData, err := i.verifyStore.GetByCode(code)
	if err != nil {
		return err
	}
	if verifyData == nil {
		return errors.New("указанный код не найден")
	}
	if !verifyData.DateUsed.IsZero() {
		return errors.New("указанный код уже использован")
	}
	if verifyData.DateExpires.After(time.Now()) {
		return errors.New("срок указанного кода истек")
	}
	updMap := map[string]interface{}{
		"date_used": time.Now(),
	}
	return i.verifyStore.UpdateByCode(code, updMap)
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
