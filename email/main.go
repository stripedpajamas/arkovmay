package email

import (
	"net/smtp"
)

type EmailSettings struct {
	Username string
	Password string
	Host     string
	HostPort string
}

var SMTPAuth smtp.Auth
var emailSettings *EmailSettings

func init() {
	emailSettings = &EmailSettings{}
	SMTPAuth = smtp.PlainAuth("",
		emailSettings.Username,
		emailSettings.Password,
		emailSettings.Host,
	)
}

func SendEmail(recipient string, email []byte) error {
	return smtp.SendMail(
		emailSettings.HostPort,
		SMTPAuth,
		emailSettings.Username,
		[]string{recipient},
		email)
}
