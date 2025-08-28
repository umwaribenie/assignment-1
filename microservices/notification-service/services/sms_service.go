package services

import (
	"bytes"
	"fmt"
	"log"
	"notification-service/config"
	"notification-service/models"
	"text/template"
)

type SMSService struct{}

func NewSMSService() *SMSService {
	return &SMSService{}
}

// SendSMS sends an SMS (mock implementation - replace with actual SMS provider)
func (ss *SMSService) SendSMS(to, content string) error {
	// This is a mock implementation
	// In production, integrate with SMS providers like Twilio, AWS SNS, etc.
	
	if config.AppConfig.TwilioAccountSID == "" {
		// Mock implementation for development
		log.Printf("MOCK SMS: Sending to %s: %s", to, content)
		return nil
	}

	// TODO: Implement actual SMS sending with Twilio or other provider
	// Example with Twilio:
	// client := twilio.NewRestClient(config.AppConfig.TwilioAccountSID, config.AppConfig.TwilioAuthToken)
	// params := &api.CreateMessageParams{}
	// params.SetTo(to)
	// params.SetFrom(config.AppConfig.TwilioFromNumber)
	// params.SetBody(content)
	// 
	// _, err := client.Api.CreateMessage(params)
	// return err

	log.Printf("SMS sent successfully to %s", to)
	return nil
}

// SendTemplateSMS sends an SMS using a template
func (ss *SMSService) SendTemplateSMS(to, templateID string, variables map[string]string) error {
	smsTemplate, exists := models.SMSTemplates[templateID]
	if !exists {
		return fmt.Errorf("SMS template %s not found", templateID)
	}

	// Parse and execute template
	tmpl, err := template.New("sms").Parse(smsTemplate.Content)
	if err != nil {
		return fmt.Errorf("failed to parse SMS template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return fmt.Errorf("failed to execute SMS template: %v", err)
	}

	return ss.SendSMS(to, buf.String())
}

// SendOTPSMS sends an OTP via SMS
func (ss *SMSService) SendOTPSMS(to, otp string) error {
	variables := map[string]string{
		"OTP":     otp,
		"AppName": "Your Application",
	}

	return ss.SendTemplateSMS(to, "otp_sms", variables)
}

// SendPasswordResetSMS sends a password reset OTP via SMS
func (ss *SMSService) SendPasswordResetSMS(to, otp string) error {
	variables := map[string]string{
		"OTP":     otp,
		"AppName": "Your Application",
	}

	return ss.SendTemplateSMS(to, "password_reset_sms", variables)
}