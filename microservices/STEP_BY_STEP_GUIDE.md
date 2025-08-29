# Step-by-Step Guide: Monolithic to Microservices Transformation

This guide walks you through transforming your existing monolithic Go application into a microservices architecture.

## 🎯 Overview

We'll break down your monolithic application into two main microservices:
1. **User Service** - Handles user management with Redis caching
2. **Notification Service** - Handles notifications with Kafka event processing

## 📋 Prerequisites

- Go 1.24+
- Docker and Docker Compose
- Basic understanding of Go, Redis, Kafka, and PostgreSQL

## 🚀 Step 1: Project Structure Setup

### 1.1 Create Directory Structure
```bash
mkdir -p microservices/{user-service,notification-service,shared}
mkdir -p microservices/user-service/{cmd,internal/{handlers,models,repository,service,middleware,cache},pkg/{kafka,utils}}
mkdir -p microservices/notification-service/{cmd,internal/{handlers,models,repository,service,kafka},pkg/utils}
mkdir -p microservices/shared/{models,utils,kafka}
```

### 1.2 Initialize Go Modules
```bash
cd microservices/user-service
go mod init user-service

cd ../notification-service
go mod init notification-service
```

## 🔧 Step 2: Shared Components

### 2.1 Shared Models (`shared/models/models.go`)
Create common data structures used across services:

```go
package models

import "time"

type UserRole string
const (
    RoleUser  UserRole = "user"
    RoleAdmin UserRole = "admin"
)

type UserStatus string
const (
    ActiveStatus   UserStatus = "active"
    InactiveStatus UserStatus = "inactive"
    DeletedStatus  UserStatus = "deleted"
)

type User struct {
    ID             string     `json:"id"`
    ClientID       string     `json:"clientId"`
    Email          string     `json:"email"`
    FirstName      string     `json:"firstName"`
    LastName       string     `json:"lastName"`
    Phone          string     `json:"phone"`
    Username       string     `json:"username"`
    Role           UserRole   `json:"role"`
    Status         UserStatus `json:"status"`
    CreatedAt      time.Time  `json:"createdAt"`
    UpdatedAt      time.Time  `json:"updatedAt"`
}
```

### 2.2 Shared Kafka Events (`shared/kafka/events.go`)
Define event structures for inter-service communication:

```go
package kafka

import (
    "time"
    "github.com/google/uuid"
)

type EventType string
const (
    EventUserCreated     EventType = "user.created"
    EventUserUpdated     EventType = "user.updated"
    EventUserStatusChanged EventType = "user.status_changed"
    EventUserDeleted     EventType = "user.deleted"
)

type BaseEvent struct {
    ID        string    `json:"id"`
    Type      EventType `json:"type"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`
}

type UserCreatedEvent struct {
    BaseEvent
    Data struct {
        UserID    string `json:"user_id"`
        Email     string `json:"email"`
        Username  string `json:"username"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Phone     string `json:"phone"`
    } `json:"data"`
}
```

## 🏗️ Step 3: User Service Implementation

### 3.1 Database Models (`user-service/internal/models/models.go`)
```go
package models

import (
    "time"
    "gorm.io/gorm"
    sharedModels "microservices/shared/models"
)

type User struct {
    ID             string                `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
    ClientID       string                `gorm:"unique" json:"clientId"`
    Email          string                `gorm:"uniqueIndex" json:"email"`
    FirstName      string                `json:"firstName"`
    LastName       string                `json:"lastName"`
    Password       string                `json:"-"`
    Phone          string                `gorm:"unique" json:"phone"`
    Username       string                `gorm:"uniqueIndex" json:"username"`
    Slug           string                `gorm:"uniqueIndex" json:"slug"`
    Role           sharedModels.UserRole `gorm:"type:varchar(50);default:'user'" json:"role"`
    Status         sharedModels.UserStatus `gorm:"type:varchar(50);default:'active'" json:"status"`
    CreatedAt      time.Time             `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt      time.Time             `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt      gorm.DeletedAt        `gorm:"index" json:"-"`
}
```

### 3.2 Repository Layer (`user-service/internal/repository/user_repository.go`)
```go
package repository

import (
    "context"
    "time"
    "microservices/user-service/internal/models"
    "microservices/user-service/pkg/utils"
    "gorm.io/gorm"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, id string, user *models.User) error
    Delete(ctx context.Context, id string) error
    GetAll(ctx context.Context, filter models.UserFilter) ([]models.User, int64, error)
}

type userRepository struct {
    db    *gorm.DB
    cache *utils.CacheClient
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
    // Try cache first
    cacheKey := utils.GenerateUserCacheKey(id)
    var user models.User
    
    if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil
    }
    
    // Get from database
    if err := r.db.First(&user, "id = ?", id).Error; err != nil {
        return nil, err
    }
    
    // Store in cache
    r.cache.Set(ctx, cacheKey, user, 30*time.Minute)
    return &user, nil
}
```

### 3.3 Service Layer (`user-service/internal/service/user_service.go`)
```go
package service

import (
    "context"
    "microservices/user-service/internal/models"
    "microservices/user-service/internal/repository"
    "microservices/user-service/pkg/kafka"
    "golang.org/x/crypto/bcrypt"
)

type UserService interface {
    CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error)
    GetUserByID(ctx context.Context, id string) (*models.UserResponse, error)
    UpdateUser(ctx context.Context, id string, req models.UpdateUserRequest) (*models.UserResponse, error)
    DeleteUser(ctx context.Context, id string) error
}

type userService struct {
    userRepo      repository.UserRepository
    kafkaProducer *kafka.Producer
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    
    // Create user
    user := &models.User{
        ClientID:  req.ClientID,
        Email:     req.Email,
        FirstName: req.FirstName,
        LastName:  req.LastName,
        Password:  string(hashedPassword),
        Phone:     req.Phone,
        Username:  req.Username,
        Role:      "user",
        Status:    "active",
    }
    
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Publish event
    s.kafkaProducer.PublishUserCreated(ctx, user.ID, user.Email, user.Username, user.FirstName, user.LastName, user.Phone)
    
    return s.toUserResponse(user), nil
}
```

### 3.4 HTTP Handlers (`user-service/internal/handlers/user_handlers.go`)
```go
package handlers

import (
    "net/http"
    "microservices/user-service/internal/models"
    "microservices/user-service/internal/service"
    "github.com/gin-gonic/gin"
)

type UserHandler struct {
    userService service.UserService
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req models.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
        return
    }
    
    user, err := h.userService.CreateUser(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, user)
}
```

### 3.5 Main Application (`user-service/cmd/main.go`)
```go
package main

import (
    "log"
    "microservices/user-service/internal/handlers"
    "microservices/user-service/internal/repository"
    "microservices/user-service/internal/service"
    "microservices/user-service/pkg/kafka"
    "microservices/user-service/pkg/utils"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    godotenv.Load()
    
    // Initialize dependencies
    db := initDatabase()
    cache := utils.NewCacheClient()
    kafkaProducer := kafka.NewProducer()
    
    // Initialize layers
    userRepo := repository.NewUserRepository(db, cache)
    userService := service.NewUserService(userRepo, cache, kafkaProducer)
    userHandler := handlers.NewUserHandler(userService)
    
    // Setup router
    router := gin.Default()
    setupRoutes(router, userHandler)
    
    // Start server
    log.Fatal(router.Run(":8081"))
}
```

## 📧 Step 4: Notification Service Implementation

### 4.1 Database Models (`notification-service/internal/models/models.go`)
```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Notification struct {
    ID         string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
    UserID     string         `json:"userId"`
    Type       string         `json:"type"` // email, sms, push
    Subject    string         `json:"subject"`
    Message    string         `json:"message"`
    Recipient  string         `json:"recipient"`
    Status     string         `json:"status"` // pending, sent, failed
    Template   string         `json:"template"`
    Variables  map[string]interface{} `gorm:"type:jsonb" json:"variables"`
    RetryCount int            `json:"retryCount"`
    MaxRetries int            `json:"maxRetries"`
    SentAt     *time.Time     `json:"sentAt"`
    FailedAt   *time.Time     `json:"failedAt"`
    ErrorMsg   string         `json:"errorMsg"`
    CreatedAt  time.Time      `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 4.2 Kafka Consumer (`notification-service/internal/kafka/consumer.go`)
```go
package kafka

import (
    "context"
    "encoding/json"
    "log"
    "microservices/shared/kafka"
    "microservices/notification-service/internal/service"
    "github.com/segmentio/kafka-go"
)

type Consumer struct {
    reader *kafka.Reader
    notificationService service.NotificationService
}

func (c *Consumer) Start(ctx context.Context) {
    for {
        message, err := c.reader.ReadMessage(ctx)
        if err != nil {
            log.Printf("Error reading message: %v", err)
            continue
        }
        
        c.handleMessage(ctx, message)
    }
}

func (c *Consumer) handleMessage(ctx context.Context, message kafka.Message) {
    var baseEvent kafka.BaseEvent
    if err := json.Unmarshal(message.Value, &baseEvent); err != nil {
        log.Printf("Error unmarshaling event: %v", err)
        return
    }
    
    switch baseEvent.Type {
    case kafka.EventUserCreated:
        c.handleUserCreated(ctx, message.Value)
    case kafka.EventUserUpdated:
        c.handleUserUpdated(ctx, message.Value)
    case kafka.EventUserStatusChanged:
        c.handleUserStatusChanged(ctx, message.Value)
    }
}

func (c *Consumer) handleUserCreated(ctx context.Context, data []byte) {
    var event kafka.UserCreatedEvent
    if err := json.Unmarshal(data, &event); err != nil {
        log.Printf("Error unmarshaling user created event: %v", err)
        return
    }
    
    // Create welcome notifications
    c.notificationService.CreateWelcomeNotifications(ctx, event.Data.UserID, event.Data.Email, event.Data.Phone)
}
```

## 🐳 Step 5: Docker Configuration

### 5.1 User Service Dockerfile (`user-service/Dockerfile`)
```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8081
CMD ["./main"]
```

### 5.2 Docker Compose (`docker-compose.yml`)
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: user_service
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: 12345
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/init.sql:/docker-entrypoint-initdb.d/init.sql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    ports:
      - "9092:9092"

  user-service:
    build: ./user-service
    ports:
      - "8081:8081"
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - postgres
      - redis
      - kafka

  notification-service:
    build: ./notification-service
    ports:
      - "8082:8082"
    environment:
      - DB_HOST=postgres
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - postgres
      - kafka

volumes:
  postgres_data:
```

## 🚀 Step 6: Testing the Implementation

### 6.1 Start Services
```bash
cd microservices
docker-compose up -d
```

### 6.2 Test User Creation
```bash
curl -X POST http://localhost:8081/api/v1/users/ \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "CLIENT001",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "password": "password123",
    "phone": "+1234567890",
    "username": "johndoe"
  }'
```

### 6.3 Check Notifications
```bash
curl http://localhost:8082/api/v1/notifications/
```

## 📚 Key Learning Points

### 1. **Service Separation**
- Each service has its own database
- Services communicate via events
- Clear boundaries and responsibilities

### 2. **Event-Driven Architecture**
- Services publish events for state changes
- Other services react to events
- Loose coupling between services

### 3. **Caching Strategy**
- Redis for frequently accessed data
- Cache invalidation on data changes
- Improved performance

### 4. **Database Design**
- Database per service pattern
- Shared models for consistency
- Proper indexing and constraints

## 🔄 Next Steps

1. **Add Authentication/Authorization**
2. **Implement API Gateway**
3. **Add Monitoring and Logging**
4. **Set up CI/CD Pipeline**
5. **Add Unit and Integration Tests**
6. **Implement Circuit Breakers**
7. **Add Rate Limiting**

## 📖 Additional Resources

- [Microservices.io](https://microservices.io/)
- [Martin Fowler's Blog](https://martinfowler.com/articles/microservices.html)
- [Go Documentation](https://golang.org/doc/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)
- [Redis Documentation](https://redis.io/documentation)

This transformation demonstrates the core principles of microservices architecture while maintaining the functionality of your original monolithic application.