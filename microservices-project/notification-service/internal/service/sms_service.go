package service

import (
	"fmt"
	"log"
	"notification-service/config"
	"notification-service/internal/models"
)

type SMSService interface {
	SendSMS(notification models.SMSNotification) error
}

type smsService struct {
	enabled     bool
	accountSID  string
	authToken   string
	fromNumber  string
}

func NewSMSService(cfg *config.Config) SMSService {
	return &smsService{
		enabled:     cfg.SMSEnabled,
		accountSID:  cfg.TwilioAccountSID,
		authToken:   cfg.TwilioAuthToken,
		fromNumber:  cfg.TwilioFromNumber,
	}
}

func (s *smsService) SendSMS(notification models.SMSNotification) error {
	if !s.enabled {
		log.Printf("SMS service is disabled, skipping SMS to %s", notification.To)
		return nil
	}
	
	// TODO: Implement Twilio SMS sending
	// For now, just log the SMS
	log.Printf("SMS would be sent to %s: %s", notification.To, notification.Body)
	
	// Example Twilio implementation (uncomment and add twilio-go dependency):
	/*
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.accountSID,
		Password: s.authToken,
	})
	
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(notification.To)
	params.SetFrom(s.fromNumber)
	params.SetBody(notification.Body)
	
	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %v", err)
	}
	
	log.Printf("SMS sent successfully: %s", resp.Sid)
	*/
	
	return nil
}

// GenerateOTP generates a 6-digit OTP
func GenerateOTP() string {
	// In production, use a cryptographically secure random number generator
	return fmt.Sprintf("%06d", 123456) // Placeholder
}