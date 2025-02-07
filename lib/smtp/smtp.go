package smtp

import (
	"crypto/tls"
	"fmt"
	"hr-tools-backend/models"
	"io"
	"strconv"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

var Instance Provider

type Provider interface {
	SendEMail(from, to, message, subject string) error
	IsConfigured() bool
	SendHtmlEMail(from, to, message, subject string, attachment *models.File) (err error)
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
	logger := log.
		WithField("sender", from).
		WithField("to", to).
		WithField("subject", subject)
	if i.user == "" || i.host == "" || i.port == "" {
		logger.Warn("Письмо не отправлено, тк не настроен smtp клиент")
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

func (i impl) SendHtmlEMail(from, to, message, subject string, attachment *models.File) (err error) {
	logger := log.
		WithField("sender", from).
		WithField("to", to).
		WithField("subject", subject)
	if i.user == "" || i.host == "" || i.port == "" {
		logger.Warn("Письмо не отправлено, тк не настроен smtp клиент")
		return nil
	}
	email := gomail.NewMessage()
	email.SetHeader("From", from)
	email.SetHeader("To", to) //TODO
	email.SetHeader("Subject", subject)
	email.SetBody("text/html", message)
	if attachment != nil && len(attachment.Body) != 0 {
		email.Attach(
			fmt.Sprint(attachment.FileName),
			gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(attachment.Body)
				return err
			}),
			gomail.SetHeader(map[string][]string{"Content-Type": {attachment.ContentType}}),
		)
	}
	port, err := strconv.Atoi(i.port)
	if err != nil {
		return errors.Wrap(err, "порт указан некорректно")
	}
	d := gomail.NewDialer(i.host, port, i.user, i.password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err = d.DialAndSend(email)
	if err != nil {
		return err
	}
	logger.Info("письмо отправлено")
	return nil
}

func (i impl) IsConfigured() bool {
	return i.user != "" && i.host != "" && i.port != ""
}
