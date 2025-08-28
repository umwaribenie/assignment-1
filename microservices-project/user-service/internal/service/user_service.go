package service

import (
	"context"
	"errors"
	"time"
	"user-service/config"
	"user-service/internal/kafka"
	"user-service/internal/models"
	"user-service/internal/repository"
	"user-service/pkg/utils"

	"github.com/google/uuid"
)

type UserService interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req models.LoginRequest, ip string) (*models.LoginResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req models.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req models.ResetPasswordRequest) error
	VerifyEmail(ctx context.Context, token string) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetAllUsers(ctx context.Context, filter models.UserFilter) (*models.PaginatedResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
	producer kafka.Producer
}

func NewUserService(userRepo repository.UserRepository, producer kafka.Producer) UserService {
	return &userService{
		userRepo: userRepo,
		producer: producer,
	}
}

func (s *userService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	existingUser, _ = s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, errors.New("user with this username already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:       req.Email,
		Username:    req.Username,
		Password:    hashedPassword,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Role:        models.RoleUser,
		Status:      models.StatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send notification for email verification
	notification := models.NotificationEvent{
		EventType: models.EventUserRegistered,
		UserID:    user.ID.String(),
		Email:     user.Email,
		Data: map[string]interface{}{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"username":   user.Username,
		},
		Timestamp: time.Now(),
	}

	if err := s.producer.SendNotification(config.AppConfig.KafkaNotificationTopic, notification); err != nil {
		// Log error but don't fail registration
		// You might want to implement a retry mechanism here
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, req models.LoginRequest, ip string) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		// Increment failed login count
		// You might want to implement account lockout after N failed attempts
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if user.Status != models.StatusActive {
		return nil, errors.New("account is not active")
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateJWT(*user)
	if err != nil {
		return nil, err
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		// Log error but don't fail login
	}

	// Send login notification
	notification := models.NotificationEvent{
		EventType: models.EventLoginNotification,
		UserID:    user.ID.String(),
		Email:     user.Email,
		Data: map[string]interface{}{
			"ip":       ip,
			"time":     time.Now().Format(time.RFC3339),
			"location": "Unknown", // You can implement IP geolocation
		},
		Timestamp: time.Now(),
	}

	s.producer.SendNotification(config.AppConfig.KafkaNotificationTopic, notification)

	// Convert to response
	userResponse := models.UserResponse{
		ID:               user.ID.String(),
		Email:            user.Email,
		Username:         user.Username,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		PhoneNumber:      user.PhoneNumber,
		ProfilePicture:   user.ProfilePicture,
		Role:             user.Role,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		PhoneVerified:    user.PhoneVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      userResponse,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.City != "" {
		user.City = req.City
	}
	if req.State != "" {
		user.State = req.State
	}
	if req.Country != "" {
		user.Country = req.Country
	}
	if req.ZipCode != "" {
		user.ZipCode = req.ZipCode
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, req models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return err
	}

	// Send notification
	notification := models.NotificationEvent{
		EventType: models.EventPasswordChanged,
		UserID:    user.ID.String(),
		Email:     user.Email,
		Data: map[string]interface{}{
			"changed_at": time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	}

	s.producer.SendNotification(config.AppConfig.KafkaNotificationTopic, notification)

	return nil
}

func (s *userService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate reset token
	resetToken := utils.GenerateResetToken()
	
	// Store reset token in database (you'll need to add this to the user model and repository)
	// For now, we'll just send the notification

	// Send notification
	notification := models.NotificationEvent{
		EventType: models.EventPasswordReset,
		UserID:    user.ID.String(),
		Email:     user.Email,
		Data: map[string]interface{}{
			"reset_token": resetToken,
			"expires_at":  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	}

	return s.producer.SendNotification(config.AppConfig.KafkaNotificationTopic, notification)
}

func (s *userService) ResetPassword(ctx context.Context, req models.ResetPasswordRequest) error {
	// TODO: Implement token validation and password reset
	// This would involve checking the reset token in the database
	// and updating the password if valid
	return errors.New("not implemented")
}

func (s *userService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement email verification
	// This would involve checking the verification token
	// and updating the email_verified field
	return errors.New("not implemented")
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) GetAllUsers(ctx context.Context, filter models.UserFilter) (*models.PaginatedResponse, error) {
	users, total, err := s.userRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Convert to user responses
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = models.UserResponse{
			ID:               user.ID.String(),
			Email:            user.Email,
			Username:         user.Username,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			PhoneNumber:      user.PhoneNumber,
			ProfilePicture:   user.ProfilePicture,
			Role:             user.Role,
			Status:           user.Status,
			EmailVerified:    user.EmailVerified,
			PhoneVerified:    user.PhoneVerified,
			TwoFactorEnabled: user.TwoFactorEnabled,
			CreatedAt:        user.CreatedAt,
			UpdatedAt:        user.UpdatedAt,
		}
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:         userResponses,
		TotalRecords: total,
		TotalPages:   totalPages,
		CurrentPage:  filter.PageNumber,
		PageSize:     filter.PageSize,
	}, nil
}