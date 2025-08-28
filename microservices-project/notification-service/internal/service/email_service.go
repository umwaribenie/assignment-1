package service

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"notification-service/config"
	"notification-service/internal/models"
	"path/filepath"

	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendEmail(notification models.EmailNotification) error
	SendTemplatedEmail(to []string, templateName string, data interface{}) error
}

type emailService struct {
	dialer    *gomail.Dialer
	from      string
	templates map[string]*template.Template
}

func NewEmailService(cfg *config.Config) EmailService {
	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	
	service := &emailService{
		dialer:    dialer,
		from:      cfg.SMTPFrom,
		templates: make(map[string]*template.Template),
	}
	
	// Load email templates
	service.loadTemplates(cfg.TemplateDir)
	
	return service
}

func (s *emailService) SendEmail(notification models.EmailNotification) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", notification.To...)
	m.SetHeader("Subject", notification.Subject)
	
	if notification.IsHTML {
		m.SetBody("text/html", notification.Body)
	} else {
		m.SetBody("text/plain", notification.Body)
	}
	
	// Add attachments if any
	for _, attachment := range notification.Attachments {
		m.Attach(attachment)
	}
	
	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}
	
	log.Printf("Email sent successfully to: %v", notification.To)
	return nil
}

func (s *emailService) SendTemplatedEmail(to []string, templateName string, data interface{}) error {
	tmpl, exists := s.templates[templateName]
	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}
	
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	
	// Extract subject from template (assuming first line is subject)
	bodyStr := body.String()
	lines := bytes.Split(body.Bytes(), []byte("\n"))
	subject := string(lines[0])
	if len(lines) > 1 {
		bodyStr = string(bytes.Join(lines[1:], []byte("\n")))
	}
	
	notification := models.EmailNotification{
		To:      to,
		Subject: subject,
		Body:    bodyStr,
		IsHTML:  true,
	}
	
	return s.SendEmail(notification)
}

func (s *emailService) loadTemplates(templateDir string) {
	// Define template file mappings
	templateFiles := map[string]string{
		"registration":       "registration.html",
		"password_reset":     "password_reset.html",
		"password_changed":   "password_changed.html",
		"email_verification": "email_verification.html",
		"login_notification": "login_notification.html",
		"otp":                "otp.html",
	}
	
	for name, file := range templateFiles {
		path := filepath.Join(templateDir, file)
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			log.Printf("Failed to load template %s: %v", name, err)
			continue
		}
		s.templates[name] = tmpl
	}
	
	log.Printf("Loaded %d email templates", len(s.templates))
}