package email

import (
	"bytes"
	"html/template"
)

type LoginEmailData struct {
	From       string
	To         string
	LoginToken string
}

const LoginEmailTemplate = `From: {{.From}}
To: {{.To}}
Subject: Arkovmay Login Link

Hello! Please click the link below to login to your Arkovmay account. The link will expire in 10 minutes!

http://localhost:8080/activate?email={{.To}}&token={{.ActivationToken}}

If you didn't request to login to Arkovmay, please ignore this email.

{{.From}}
`

func SendLoginEmail(to, activationCode string) error {
	var e bytes.Buffer
	context := &LoginEmailData{
		"Arkovmay Login Email",
		to,
		activationCode,
	}
	t := template.New("login-template")
	t, err := t.Parse(LoginEmailTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(&e, context)
	if err != nil {
		return err
	}

	// send the email
	return SendEmail(to, e.Bytes())
}
