package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"microservices/user-service/internal/models"
	"microservices/user-service/internal/repository"
	"microservices/user-service/pkg/kafka"
	"microservices/user-service/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

// UserService interface defines the business logic methods
type UserService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*models.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*models.UserResponse, error)
	GetUserByUsername(ctx context.Context, username string) (*models.UserResponse, error)
	GetUserBySlug(ctx context.Context, slug string) (*models.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	GetAllUsers(ctx context.Context, filter models.UserFilter) (*models.PaginatedResponse, error)
	UpdatePassword(ctx context.Context, id string, newPassword string) error
	CreateUserByAdmin(ctx context.Context, req models.CreateUserByAdminRequest) (*models.UserResponse, error)
}

// userService implements UserService
type userService struct {
	userRepo      repository.UserRepository
	cache         *utils.CacheClient
	kafkaProducer *kafka.Producer
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, cache *utils.CacheClient, kafkaProducer *kafka.Producer) UserService {
	return &userService{
		userRepo:      userRepo,
		cache:         cache,
		kafkaProducer: kafkaProducer,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {
	// Check if user already exists
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("user with this username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate slug
	slug := s.generateSlug(req.FirstName + " " + req.LastName)

	// Create user
	user := &models.User{
		ClientID:       req.ClientID,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		NationalID:     req.NationalID,
		PassportNumber: req.PassportNumber,
		Password:       string(hashedPassword),
		Phone:          req.Phone,
		ProfilePicture: req.ProfilePicture,
		Username:       req.Username,
		Slug:           slug,
		Role:           "user", // Default role
		Status:         "active", // Default status
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish user created event
	if err := s.kafkaProducer.PublishUserCreated(ctx, user.ID, user.Email, user.Username, user.FirstName, user.LastName, user.Phone); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish user created event: %v\n", err)
	}

	return s.toUserResponse(user), nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(ctx context.Context, id string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUserByUsername retrieves a user by username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUserBySlug retrieves a user by slug
func (s *userService) GetUserBySlug(ctx context.Context, slug string) (*models.UserResponse, error) {
	user, err := s.userRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, id string, req models.UpdateUserRequest) (*models.UserResponse, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Track changes for event publishing
	changes := make(map[string]interface{})

	// Update fields if provided
	if req.Email != nil && *req.Email != existingUser.Email {
		// Check if new email already exists
		if _, err := s.userRepo.GetByEmail(ctx, *req.Email); err == nil {
			return nil, errors.New("user with this email already exists")
		}
		existingUser.Email = *req.Email
		changes["email"] = *req.Email
	}

	if req.FirstName != nil {
		existingUser.FirstName = *req.FirstName
		changes["first_name"] = *req.FirstName
	}

	if req.LastName != nil {
		existingUser.LastName = *req.LastName
		changes["last_name"] = *req.LastName
	}

	if req.NationalID != nil {
		existingUser.NationalID = req.NationalID
		changes["national_id"] = req.NationalID
	}

	if req.PassportNumber != nil {
		existingUser.PassportNumber = req.PassportNumber
		changes["passport_number"] = req.PassportNumber
	}

	if req.Phone != nil && *req.Phone != existingUser.Phone {
		existingUser.Phone = *req.Phone
		changes["phone"] = *req.Phone
	}

	if req.ProfilePicture != nil {
		existingUser.ProfilePicture = req.ProfilePicture
		changes["profile_picture"] = req.ProfilePicture
	}

	if req.Role != nil {
		existingUser.Role = *req.Role
		changes["role"] = *req.Role
	}

	if req.Username != nil && *req.Username != existingUser.Username {
		// Check if new username already exists
		if _, err := s.userRepo.GetByUsername(ctx, *req.Username); err == nil {
			return nil, errors.New("user with this username already exists")
		}
		existingUser.Username = *req.Username
		changes["username"] = *req.Username
	}

	// Update user
	if err := s.userRepo.Update(ctx, id, existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Publish user updated event if there were changes
	if len(changes) > 0 {
		if err := s.kafkaProducer.PublishUserUpdated(ctx, existingUser.ID, existingUser.Email, existingUser.Username, existingUser.FirstName, existingUser.LastName, existingUser.Phone, changes); err != nil {
			fmt.Printf("Failed to publish user updated event: %v\n", err)
		}
	}

	return s.toUserResponse(existingUser), nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	// Get user before deletion for event publishing
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Publish user deleted event
	if err := s.kafkaProducer.PublishUserDeleted(ctx, user.ID, user.Email, user.Username, "system"); err != nil {
		fmt.Printf("Failed to publish user deleted event: %v\n", err)
	}

	return nil
}

// GetAllUsers retrieves all users with pagination and filtering
func (s *userService) GetAllUsers(ctx context.Context, filter models.UserFilter) (*models.PaginatedResponse, error) {
	users, total, err := s.userRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.toUserResponse(&user)
	}

	// Calculate pagination
	lastPage := int(math.Ceil(float64(total) / float64(filter.PageSize)))
	if lastPage == 0 && total > 0 {
		lastPage = 1
	}

	var nextPage, previousPage *int
	if filter.PageNumber < lastPage {
		next := filter.PageNumber + 1
		nextPage = &next
	}
	if filter.PageNumber > 1 {
		prev := filter.PageNumber - 1
		previousPage = &prev
	}

	return &models.PaginatedResponse{
		CurrentPage:  filter.PageNumber,
		LastPage:     lastPage,
		List:         userResponses,
		NextPage:     nextPage,
		PreviousPage: previousPage,
		Status:       "success",
		Total:        total,
	}, nil
}

// UpdatePassword updates a user's password
func (s *userService) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, id, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// CreateUserByAdmin creates a user with admin privileges
func (s *userService) CreateUserByAdmin(ctx context.Context, req models.CreateUserByAdminRequest) (*models.UserResponse, error) {
	// Check if user already exists
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("user with this username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate slug
	slug := s.generateSlug(req.FirstName + " " + req.LastName)

	// Create user
	user := &models.User{
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		NationalID:     req.NationalID,
		PassportNumber: req.PassportNumber,
		Password:       string(hashedPassword),
		Phone:          req.Phone,
		ProfilePicture: req.ProfilePicture,
		Username:       req.Username,
		Slug:           slug,
		Role:           req.Role,
		Status:         "active", // Default status
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Publish user created event
	if err := s.kafkaProducer.PublishUserCreated(ctx, user.ID, user.Email, user.Username, user.FirstName, user.LastName, user.Phone); err != nil {
		fmt.Printf("Failed to publish user created event: %v\n", err)
	}

	return s.toUserResponse(user), nil
}

// toUserResponse converts a User model to UserResponse
func (s *userService) toUserResponse(user *models.User) *models.UserResponse {
	return &models.UserResponse{
		ID:             user.ID,
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       user.Username,
		Slug:           user.Slug,
		Role:           user.Role,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}

// generateSlug generates a URL-friendly slug from a string
func (s *userService) generateSlug(text string) string {
	// Simple slug generation - in production, you might want a more sophisticated approach
	slug := ""
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			if char >= 'A' && char <= 'Z' {
				slug += string(char + 32) // Convert to lowercase
			} else {
				slug += string(char)
			}
		} else if char == ' ' {
			slug += "-"
		}
	}

	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", slug, timestamp)
}