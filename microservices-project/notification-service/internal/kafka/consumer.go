package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"notification-service/internal/models"
	"notification-service/internal/service"

	"github.com/Shopify/sarama"
)

type NotificationConsumer interface {
	Start(ctx context.Context) error
	Stop() error
}

type notificationConsumer struct {
	consumer     sarama.ConsumerGroup
	topics       []string
	emailService service.EmailService
	smsService   service.SMSService
}

func NewNotificationConsumer(
	brokers []string,
	groupID string,
	topics []string,
	emailService service.EmailService,
	smsService service.SMSService,
) (NotificationConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &notificationConsumer{
		consumer:     consumer,
		topics:       topics,
		emailService: emailService,
		smsService:   smsService,
	}, nil
}

func (c *notificationConsumer) Start(ctx context.Context) error {
	handler := &consumerGroupHandler{
		emailService: c.emailService,
		smsService:   c.smsService,
	}

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Consume messages
		err := c.consumer.Consume(ctx, c.topics, handler)
		if err != nil {
			log.Printf("Error consuming messages: %v", err)
			return err
		}
	}
}

func (c *notificationConsumer) Stop() error {
	return c.consumer.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	emailService service.EmailService
	smsService   service.SMSService
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Message received: topic=%s partition=%d offset=%d", message.Topic, message.Partition, message.Offset)
		
		// Process the message
		if err := h.processMessage(message.Value); err != nil {
			log.Printf("Error processing message: %v", err)
			// In production, you might want to send to a dead letter queue
		}
		
		// Mark message as processed
		session.MarkMessage(message, "")
	}
	
	return nil
}

func (h *consumerGroupHandler) processMessage(data []byte) error {
	var event models.NotificationEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	
	log.Printf("Processing event: %s for user: %s", event.EventType, event.UserID)
	
	switch event.EventType {
	case models.EventUserRegistered:
		return h.handleUserRegistered(event)
	case models.EventPasswordReset:
		return h.handlePasswordReset(event)
	case models.EventPasswordChanged:
		return h.handlePasswordChanged(event)
	case models.EventEmailVerification:
		return h.handleEmailVerification(event)
	case models.EventLoginNotification:
		return h.handleLoginNotification(event)
	case models.EventOTPGenerated:
		return h.handleOTPGenerated(event)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
	}
	
	return nil
}

func (h *consumerGroupHandler) handleUserRegistered(event models.NotificationEvent) error {
	// Send welcome email
	data := map[string]interface{}{
		"FirstName": event.Data["first_name"],
		"LastName":  event.Data["last_name"],
		"Username":  event.Data["username"],
		"Email":     event.Email,
	}
	
	// If template exists, use it; otherwise send plain email
	if err := h.emailService.SendTemplatedEmail([]string{event.Email}, "registration", data); err != nil {
		// Fallback to plain email
		notification := models.EmailNotification{
			To:      []string{event.Email},
			Subject: "Welcome to Our Platform!",
			Body:    fmt.Sprintf("Hello %s %s,\n\nWelcome to our platform! Your account has been created successfully.\n\nUsername: %s\nEmail: %s\n\nBest regards,\nThe Team", 
				event.Data["first_name"], event.Data["last_name"], event.Data["username"], event.Email),
			IsHTML: false,
		}
		return h.emailService.SendEmail(notification)
	}
	
	return nil
}

func (h *consumerGroupHandler) handlePasswordReset(event models.NotificationEvent) error {
	data := map[string]interface{}{
		"Email":      event.Email,
		"ResetToken": event.Data["reset_token"],
		"ExpiresAt":  event.Data["expires_at"],
	}
	
	// Send password reset email
	if err := h.emailService.SendTemplatedEmail([]string{event.Email}, "password_reset", data); err != nil {
		// Fallback to plain email
		notification := models.EmailNotification{
			To:      []string{event.Email},
			Subject: "Password Reset Request",
			Body:    fmt.Sprintf("Hello,\n\nYou requested a password reset. Use this token to reset your password: %s\n\nThis token expires at: %s\n\nIf you didn't request this, please ignore this email.\n\nBest regards,\nThe Team", 
				event.Data["reset_token"], event.Data["expires_at"]),
			IsHTML: false,
		}
		return h.emailService.SendEmail(notification)
	}
	
	return nil
}

func (h *consumerGroupHandler) handlePasswordChanged(event models.NotificationEvent) error {
	data := map[string]interface{}{
		"Email":     event.Email,
		"ChangedAt": event.Data["changed_at"],
	}
	
	// Send password changed notification
	if err := h.emailService.SendTemplatedEmail([]string{event.Email}, "password_changed", data); err != nil {
		notification := models.EmailNotification{
			To:      []string{event.Email},
			Subject: "Password Changed Successfully",
			Body:    fmt.Sprintf("Hello,\n\nYour password was changed successfully at %s.\n\nIf you didn't make this change, please contact support immediately.\n\nBest regards,\nThe Team", 
				event.Data["changed_at"]),
			IsHTML: false,
		}
		return h.emailService.SendEmail(notification)
	}
	
	return nil
}

func (h *consumerGroupHandler) handleEmailVerification(event models.NotificationEvent) error {
	data := map[string]interface{}{
		"Email":             event.Email,
		"VerificationToken": event.Data["verification_token"],
	}
	
	// Send email verification
	if err := h.emailService.SendTemplatedEmail([]string{event.Email}, "email_verification", data); err != nil {
		notification := models.EmailNotification{
			To:      []string{event.Email},
			Subject: "Verify Your Email",
			Body:    fmt.Sprintf("Hello,\n\nPlease verify your email using this token: %s\n\nBest regards,\nThe Team", 
				event.Data["verification_token"]),
			IsHTML: false,
		}
		return h.emailService.SendEmail(notification)
	}
	
	return nil
}

func (h *consumerGroupHandler) handleLoginNotification(event models.NotificationEvent) error {
	data := map[string]interface{}{
		"Email":    event.Email,
		"IP":       event.Data["ip"],
		"Time":     event.Data["time"],
		"Location": event.Data["location"],
	}
	
	// Send login notification
	if err := h.emailService.SendTemplatedEmail([]string{event.Email}, "login_notification", data); err != nil {
		notification := models.EmailNotification{
			To:      []string{event.Email},
			Subject: "New Login to Your Account",
			Body:    fmt.Sprintf("Hello,\n\nA new login to your account was detected:\n\nTime: %s\nIP: %s\nLocation: %s\n\nIf this wasn't you, please secure your account immediately.\n\nBest regards,\nThe Team", 
				event.Data["time"], event.Data["ip"], event.Data["location"]),
			IsHTML: false,
		}
		return h.emailService.SendEmail(notification)
	}
	
	return nil
}

func (h *consumerGroupHandler) handleOTPGenerated(event models.NotificationEvent) error {
	// Send OTP via SMS if phone number is provided
	if phone, ok := event.Data["phone"].(string); ok && phone != "" {
		smsNotification := models.SMSNotification{
			To:   phone,
			Body: fmt.Sprintf("Your OTP is: %s. Valid for 5 minutes.", event.Data["otp"]),
		}
		if err := h.smsService.SendSMS(smsNotification); err != nil {
			log.Printf("Failed to send OTP SMS: %v", err)
		}
	}
	
	// Also send via email
	notification := models.EmailNotification{
		To:      []string{event.Email},
		Subject: "Your OTP Code",
		Body:    fmt.Sprintf("Your OTP is: %s\n\nThis code is valid for 5 minutes.", event.Data["otp"]),
		IsHTML:  false,
	}
	
	return h.emailService.SendEmail(notification)
}