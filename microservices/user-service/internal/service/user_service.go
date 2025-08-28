package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"user-service/internal/models"
	"user-service/internal/repository"
	"user-service/pkg/kafka"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService represents the user service
type UserService struct {
	userRepo *repository.UserRepository
	producer *kafka.Producer
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, producer *kafka.Producer) *UserService {
	return &UserService{
		userRepo: userRepo,
		producer: producer,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate slug
	slug := s.generateSlug(req.FirstName, req.LastName)

	// Create user
	user := &models.User{
		ClientID:              req.ClientID,
		Email:                 req.Email,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		NationalID:            req.NationalID,
		PassportNumber:        req.PassportNumber,
		Password:              string(hashedPassword),
		Phone:                 req.Phone,
		ProfilePicture:        req.ProfilePicture,
		Username:              req.Username,
		Role:                  req.Role,
		Status:                "active",
		Slug:                  slug,
		ReferralCode:          req.ReferralCode,
		HasActiveSubscription: false,
		IsActive:              true,
		OTPRequired:           false,
		CreatedBy:             req.CreatedBy,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish user created event
	if err := s.producer.PublishUserCreated(ctx, user.ID, user.Email, user.FirstName, user.LastName, user.Phone, user.Role); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish user created event: %v\n", err)
	}

	return s.toUserResponse(user), nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.UserResponse, error) {
	// Get current user
	currentUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})
	
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.NationalID != nil {
		updates["national_id"] = *req.NationalID
	}
	if req.PassportNumber != nil {
		updates["passport_number"] = *req.PassportNumber
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.ProfilePicture != nil {
		updates["profile_picture"] = *req.ProfilePicture
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.ReferralCode != nil {
		updates["referral_code"] = *req.ReferralCode
	}

	// Handle status change
	if req.Status != nil && *req.Status != currentUser.Status {
		updates["status"] = *req.Status
		
		// Publish status changed event
		if err := s.producer.PublishUserStatusChanged(ctx, currentUser.ID, currentUser.Email, currentUser.Status, *req.Status, "system", ""); err != nil {
			fmt.Printf("Failed to publish user status changed event: %v\n", err)
		}
	}

	if err := s.userRepo.Update(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user
	updatedUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Publish user updated event
	if err := s.producer.PublishUserUpdated(ctx, updatedUser.ID, updatedUser.Email, updatedUser.FirstName, updatedUser.LastName, updatedUser.Phone, updatedUser.Role, updatedUser.Status); err != nil {
		fmt.Printf("Failed to publish user updated event: %v\n", err)
	}

	return s.toUserResponse(updatedUser), nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Get user before deletion for event publishing
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish user deleted event (you can add this to the kafka package)
	// For now, we'll just log it
	fmt.Printf("User deleted: %s (%s)\n", user.Email, user.ID)

	return nil
}

// GetAllUsers retrieves all users with filtering and pagination
func (s *UserService) GetAllUsers(ctx context.Context, filter models.UserFilter) (*models.PaginatedResponse, error) {
	response, err := s.userRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Convert users to responses
	if users, ok := response.Data.([]models.User); ok {
		var userResponses []models.UserResponse
		for _, user := range users {
			userResponses = append(userResponses, *s.toUserResponse(&user))
		}
		response.Data = userResponses
	}

	return response, nil
}

// toUserResponse converts a User to UserResponse
func (s *UserService) toUserResponse(user *models.User) *models.UserResponse {
	return &models.UserResponse{
		ID:                    user.ID,
		ClientID:              user.ClientID,
		Email:                 user.Email,
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		NationalID:            user.NationalID,
		PassportNumber:        user.PassportNumber,
		Phone:                 user.Phone,
		ProfilePicture:        user.ProfilePicture,
		Username:              user.Username,
		Role:                  user.Role,
		Status:                user.Status,
		Slug:                  user.Slug,
		ReferralCode:          user.ReferralCode,
		HasActiveSubscription: user.HasActiveSubscription,
		IsActive:              user.IsActive,
		OTPRequired:           user.OTPRequired,
		CreatedBy:             user.CreatedBy,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}
}

// generateSlug generates a unique slug for a user
func (s *UserService) generateSlug(firstName, lastName string) string {
	baseSlug := strings.ToLower(strings.ReplaceAll(firstName+"-"+lastName, " ", "-"))
	// Remove special characters
	baseSlug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, baseSlug)
	
	// Add timestamp for uniqueness
	return fmt.Sprintf("%s-%d", baseSlug, time.Now().Unix())
}