package smtp

import (
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	"strings"
)

var Instance Provider

type Provider interface {
	SendEMail(from, to, message, subject string) error
}

func Connect(user, password, host, port string, tlsEnabled bool) error {
	Instance = &impl{
		user:       user,
		password:   password,
		host:       host,
		port:       port,
		tlsEnabled: tlsEnabled,
	}
	return nil
}

type impl struct {
	user                  string
	password              string
	host                  string
	port                  string
	tlsEnabled            bool
	emailSendVerification string
}

func (i impl) SendEMail(from, to, message, subject string) (err error) {
	logger := log.WithField("sender", from)
	if i.user == "" || i.host == "" || i.port == "" {
		logger.Warn("Письмо для подтверждения почты не отправлено, тк не настроен smtp клиент")
		return nil
	}
	// Receiver email address.
	sendTo := []string{
		to,
	}
	// Authentication.
	auth := sasl.NewPlainClient("", i.user, i.password)
	//var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\r\n"
	body := strings.NewReader(fmt.Sprintf("Subject: HR Tools - %s\n%s\r\n Отправитель: %s\r\n %s\r\n", subject, mimeHeaders, from, message))

	// Sending email.
	if i.tlsEnabled {
		err = smtp.SendMailTLS(i.host+":"+i.port, auth, i.user, sendTo, body)
	} else {
		err = smtp.SendMail(i.host+":"+i.port, auth, i.user, sendTo, body)
	}
	if err != nil {
		log.WithError(err).Error("Ошибка отправки сообщения")
		return err
	}
	logger.Info("письмо отправлено")
	return nil
}
