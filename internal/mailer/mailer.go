package mailer

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"time"

	"gopkg.in/gomail.v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *gomail.Dialer
	sender string
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

func New(config Config) *Mailer {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	return &Mailer{
		dialer: dialer,
		sender: config.Sender,
	}
}

func (m *Mailer) Send(recipients []string, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	// Follow the same pattern to execute the "plainBody" template and store the result
	// in the plainBody variable.
	plainBody := new(bytes.Buffer)

	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}
	// And likewise with the "htmlBody" template.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	newMsg := gomail.NewMessage()

	// Set From
	from := m.sender
	newMsg.SetHeader("From", from)

	// Set To
	if len(recipients) == 0 {
		return errors.New("at least one recipient is required")
	}
	newMsg.SetHeader("To", recipients...)

	// Set Subject
	newMsg.SetHeader("Subject", subject.String())

	// Set Body - both plain and HTML if provided
	if htmlBody.String() != "" && plainBody.String() != "" {
		newMsg.SetBody("text/plain", plainBody.String())
		newMsg.AddAlternative("text/html", htmlBody.String())
	} else if htmlBody.String() != "" {
		newMsg.SetBody("text/html", htmlBody.String())
	} else if plainBody.String() != "" {
		newMsg.SetBody("text/plain", plainBody.String())
	} else {
		return fmt.Errorf("email body is required")
	}

	// Send the email
	var senderErr error
	for i := range 3 {
		senderErr = m.dialer.DialAndSend(newMsg)
		if senderErr == nil {
			return nil
		}
		if i < 2 { // Don't sleep after last attempt
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}
	return fmt.Errorf("failed to send email after 3 attempts: %w", senderErr)
}
