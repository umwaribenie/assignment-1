# Migration Guide: Implementing the New Folder Structure

## 🎯 Overview

This guide will help you migrate your current user management system to the new organized folder structure step by step.

## 📝 Current Structure vs New Structure

### Current Structure
```
/workspace/
├── main.go
├── config/
│   └── config.go
├── database/
│   └── database.go
├── models/
│   └── models.go (all models in one file)
├── handlers/
│   ├── auth_handlers.go
│   ├── user_handlers.go
│   └── upload_handlers.go
├── middleware/
│   └── middleware.go (all middleware in one file)
├── routes/
│   └── routes.go
└── utils/
    └── utils.go
```

### Target Structure
```
/workspace/
├── main.go
├── .env
├── .env.example
├── Makefile
├── docker-compose.yml
├── Dockerfile
├── [all directories as shown in PROJECT_STRUCTURE.md]
```

## 🚀 Migration Steps

### Step 1: Create Directory Structure

Run these commands in your project root:

```bash
# Create all necessary directories
mkdir -p config database/migrations database/seeds
mkdir -p models handlers middleware routes
mkdir -p services repositories utils dto validators
mkdir -p uploads/avatars uploads/documents uploads/temp
mkdir -p logs tests/unit tests/integration tests/e2e
mkdir -p scripts docs/api docs/guides
mkdir -p deployments/kubernetes deployments/docker
mkdir -p .github/workflows

# Set permissions for upload and log directories
chmod 755 uploads uploads/avatars uploads/documents uploads/temp
chmod 755 logs
```

### Step 2: Create Environment Configuration

Create `.env.example`:
```env
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost
ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=user_management
DB_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# File Upload Configuration
UPLOAD_MAX_SIZE=10485760
UPLOAD_ALLOWED_TYPES=image/jpeg,image/png,image/gif,application/pdf
UPLOAD_PATH=./uploads

# Email Configuration (if needed)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
SMTP_FROM=noreply@yourdomain.com

# Redis Configuration (if using)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Logging
LOG_LEVEL=info
LOG_FILE=./logs/app.log
```

Create `.gitignore`:
```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
/generalusermanagement
/generalusermanagement_extended

# Test binary
*.test

# Output of go coverage
*.out

# Dependency directories
vendor/

# Environment files
.env
.env.local

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Logs
logs/
*.log

# Uploads
uploads/
!uploads/.gitkeep

# Temporary files
tmp/
temp/

# Build artifacts
dist/
build/
```

### Step 3: Create Makefile

Create `Makefile`:
```makefile
# Variables
APP_NAME=user-management
BUILD_DIR=build
MAIN_FILE=main.go
DOCKER_IMAGE=user-management:latest

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) -v $(MAIN_FILE)

# Run the application
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	@$(GOCMD) run $(MAIN_FILE)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	@$(GOMOD) tidy

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# Generate Swagger documentation
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g main.go -o ./docs

# Database migrations
.PHONY: migrate-up
migrate-up:
	@echo "Running migrations..."
	@migrate -path database/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path database/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

# Docker commands
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker-compose up

.PHONY: docker-down
docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

# Development setup
.PHONY: dev
dev:
	@echo "Starting development environment..."
	@docker-compose up -d db redis
	@$(MAKE) migrate-up
	@$(MAKE) run

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...

# All-in-one command
.PHONY: all
all: clean deps fmt lint test build
```

### Step 4: Create Docker Configuration

Create `Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Create necessary directories
RUN mkdir -p uploads/avatars uploads/documents uploads/temp logs

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
```

Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  app:
    build: .
    container_name: user-management-app
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - REDIS_HOST=redis
    depends_on:
      - db
      - redis
    volumes:
      - ./uploads:/root/uploads
      - ./logs:/root/logs
    networks:
      - user-management-network

  db:
    image: postgres:15-alpine
    container_name: user-management-db
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=user_management
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - user-management-network

  redis:
    image: redis:7-alpine
    container_name: user-management-redis
    ports:
      - "6379:6379"
    networks:
      - user-management-network

  swagger:
    image: swaggerapi/swagger-ui
    container_name: user-management-swagger
    ports:
      - "8081:8080"
    environment:
      - SWAGGER_JSON=/docs/swagger.json
    volumes:
      - ./docs:/docs
    networks:
      - user-management-network

volumes:
  postgres_data:

networks:
  user-management-network:
    driver: bridge
```

### Step 5: Create Service Layer

Create `services/auth_service.go`:
```go
package services

import (
    "errors"
    "time"
    
    "your-module/models"
    "your-module/repositories"
    "your-module/utils"
)

type AuthService struct {
    userRepo    *repositories.UserRepository
    sessionRepo *repositories.SessionRepository
}

func NewAuthService(userRepo *repositories.UserRepository, sessionRepo *repositories.SessionRepository) *AuthService {
    return &AuthService{
        userRepo:    userRepo,
        sessionRepo: sessionRepo,
    }
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
    // Business logic for login
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        return nil, "", err
    }
    
    // Verify password
    if !utils.CheckPasswordHash(password, user.Password) {
        return nil, "", errors.New("invalid credentials")
    }
    
    // Generate token
    token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
    if err != nil {
        return nil, "", err
    }
    
    // Create session
    session := &models.Session{
        UserID:    user.ID,
        Token:     token,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    if err := s.sessionRepo.Create(session); err != nil {
        return nil, "", err
    }
    
    return user, token, nil
}

func (s *AuthService) Logout(userID uint) error {
    return s.sessionRepo.DeleteByUserID(userID)
}

func (s *AuthService) ValidateToken(token string) (*models.User, error) {
    claims, err := utils.ValidateJWT(token)
    if err != nil {
        return nil, err
    }
    
    // Check if session exists
    session, err := s.sessionRepo.FindByToken(token)
    if err != nil {
        return nil, errors.New("session not found")
    }
    
    if session.ExpiresAt.Before(time.Now()) {
        return nil, errors.New("session expired")
    }
    
    return s.userRepo.FindByID(claims.UserID)
}
```

### Step 6: Create Repository Layer

Create `repositories/user_repository.go`:
```go
package repositories

import (
    "gorm.io/gorm"
    "your-module/models"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
    return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) List(offset, limit int) ([]models.User, error) {
    var users []models.User
    err := r.db.Offset(offset).Limit(limit).Find(&users).Error
    return users, err
}
```

### Step 7: Create DTOs

Create `dto/auth_dto.go`:
```go
package dto

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
    User  UserResponse `json:"user"`
    Token string       `json:"token"`
}

type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
}
```

### Step 8: Create Database Migrations

Create `database/migrations/001_create_users_table.up.sql`:
```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    role VARCHAR(20) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    avatar_url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

Create `database/migrations/001_create_users_table.down.sql`:
```sql
DROP TABLE IF EXISTS users;
```

## 📊 Final Checklist

- [ ] Create all directory structure
- [ ] Set up environment files (.env, .env.example)
- [ ] Create Makefile for common commands
- [ ] Set up Docker configuration
- [ ] Create service layer for business logic
- [ ] Create repository layer for data access
- [ ] Create DTOs for API contracts
- [ ] Set up database migrations
- [ ] Move existing code to appropriate locations
- [ ] Update imports in all files
- [ ] Run tests to ensure everything works
- [ ] Update documentation

## 🎉 Completion

Once you've completed all these steps, you'll have a well-organized, scalable user management system following best practices and clean architecture principles!