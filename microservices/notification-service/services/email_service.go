package services

import (
	"bytes"
	"fmt"
	"log"
	"notification-service/config"
	"notification-service/models"
	"text/template"
	"time"

	"gopkg.in/gomail.v2"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendEmail sends an email
func (es *EmailService) SendEmail(to, subject, content string, isHTML bool) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", config.AppConfig.FromName, config.AppConfig.FromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	if isHTML {
		m.SetBody("text/html", content)
	} else {
		m.SetBody("text/plain", content)
	}

	d := gomail.NewDialer(
		config.AppConfig.SMTPHost,
		config.AppConfig.SMTPPort,
		config.AppConfig.SMTPUsername,
		config.AppConfig.SMTPPassword,
	)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

// SendTemplateEmail sends an email using a template
func (es *EmailService) SendTemplateEmail(to, templateID string, variables map[string]string) error {
	emailTemplate, exists := models.EmailTemplates[templateID]
	if !exists {
		return fmt.Errorf("template %s not found", templateID)
	}

	// Parse and execute subject template
	subjectTmpl, err := template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return fmt.Errorf("failed to parse subject template: %v", err)
	}

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, variables); err != nil {
		return fmt.Errorf("failed to execute subject template: %v", err)
	}

	// Parse and execute HTML body template
	htmlTmpl, err := template.New("html").Parse(emailTemplate.HTMLBody)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %v", err)
	}

	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, variables); err != nil {
		return fmt.Errorf("failed to execute HTML template: %v", err)
	}

	// Parse and execute text body template
	textTmpl, err := template.New("text").Parse(emailTemplate.TextBody)
	if err != nil {
		return fmt.Errorf("failed to parse text template: %v", err)
	}

	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, variables); err != nil {
		return fmt.Errorf("failed to execute text template: %v", err)
	}

	// Create and send email
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", config.AppConfig.FromName, config.AppConfig.FromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subjectBuf.String())
	m.SetBody("text/plain", textBuf.String())
	m.AddAlternative("text/html", htmlBuf.String())

	d := gomail.NewDialer(
		config.AppConfig.SMTPHost,
		config.AppConfig.SMTPPort,
		config.AppConfig.SMTPUsername,
		config.AppConfig.SMTPPassword,
	)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send template email to %s: %v", to, err)
		return err
	}

	log.Printf("Template email (%s) sent successfully to %s", templateID, to)
	return nil
}

// SendWelcomeEmail sends a welcome email to a new user
func (es *EmailService) SendWelcomeEmail(to, firstName string) error {
	variables := map[string]string{
		"FirstName": firstName,
		"AppName":   "Your Application",
	}

	return es.SendTemplateEmail(to, "welcome", variables)
}

// SendPasswordResetEmail sends a password reset email with OTP
func (es *EmailService) SendPasswordResetEmail(to, otp string) error {
	variables := map[string]string{
		"OTP":     otp,
		"AppName": "Your Application",
	}

	return es.SendTemplateEmail(to, "password_reset", variables)
}

// SendOTPEmail sends an OTP verification email
func (es *EmailService) SendOTPEmail(to, otp string) error {
	variables := map[string]string{
		"OTP":     otp,
		"AppName": "Your Application",
	}

	return es.SendTemplateEmail(to, "otp_verification", variables)
}