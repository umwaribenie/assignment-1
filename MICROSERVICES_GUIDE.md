# Microservices Architecture Guide for Your Go Project

## Table of Contents
1. [Understanding Microservices](#understanding-microservices)
2. [Current Architecture Analysis](#current-architecture-analysis)
3. [Proposed Microservices Architecture](#proposed-microservices-architecture)
4. [Step-by-Step Implementation](#step-by-step-implementation)
5. [Redis Caching](#redis-caching)
6. [Kafka Communication](#kafka-communication)
7. [Recommended Resources](#recommended-resources)

## Understanding Microservices

### What are Microservices?
Microservices architecture is a design pattern where a single application is developed as a suite of small, independent services. Each service:
- Runs in its own process
- Communicates via well-defined APIs (usually HTTP/REST or messaging)
- Can be deployed independently
- Has its own database (optional but recommended)
- Is organized around business capabilities

### Benefits:
1. **Scalability**: Scale individual services based on demand
2. **Flexibility**: Use different technologies for different services
3. **Fault Isolation**: If one service fails, others continue working
4. **Team Independence**: Different teams can work on different services
5. **Easier Deployment**: Deploy services independently

### Challenges:
1. **Complexity**: More moving parts to manage
2. **Network Communication**: Services need to communicate over network
3. **Data Consistency**: Managing transactions across services
4. **Testing**: Integration testing becomes more complex
5. **Monitoring**: Need robust monitoring and logging

## Current Architecture Analysis

Your current monolithic application has:
- **Controllers/Handlers**: `auth_handlers.go`, `user_handlers.go`, `upload_handlers.go`
- **Models**: User models, request/response models
- **Utilities**: JWT, OTP, Email, SMS, Password utilities
- **Database**: Single database connection
- **Routes**: All routes in one application

## Proposed Microservices Architecture

We'll split your application into:

### 1. User Service
Responsible for:
- User registration and authentication
- User profile management
- Password management
- JWT token generation/validation

### 2. Notification Service
Responsible for:
- Email notifications
- SMS notifications
- OTP generation and verification
- Notification templates and queuing

### 3. API Gateway (Optional but recommended)
- Single entry point for clients
- Request routing
- Authentication verification
- Rate limiting

### Architecture Diagram:
```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────┐
│ API Gateway │
└──────┬──────┘
       │
┌──────┴───────────────┬──────────────┐
│                      │              │
▼                      ▼              │
┌──────────────┐  ┌──────────────┐   │
│ User Service │  │ Notification │   │
│              │  │   Service    │   │
│  [Redis]     │  └──────────────┘   │
└──────────────┘                      │
       │                              │
       └──────────┬───────────────────┘
                  │
           ┌──────▼──────┐
           │   Kafka     │
           └─────────────┘
```

## Step-by-Step Implementation

### Phase 1: Project Structure Setup

Create the following directory structure:
```
microservices-project/
├── user-service/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── service/
│   │   └── middleware/
│   ├── config/
│   ├── pkg/
│   │   └── utils/
│   ├── Dockerfile
│   ├── go.mod
│   └── .env
├── notification-service/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── models/
│   │   ├── service/
│   │   └── kafka/
│   ├── config/
│   ├── templates/
│   ├── Dockerfile
│   ├── go.mod
│   └── .env
├── api-gateway/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── middleware/
│   │   └── proxy/
│   ├── Dockerfile
│   ├── go.mod
│   └── .env
├── docker-compose.yml
└── README.md
```

### Phase 2: Extracting User Service

#### Step 1: Create User Service Structure
```bash
mkdir -p microservices-project/user-service/{cmd,internal/{handlers,models,repository,service,middleware},config,pkg/utils}
```

#### Step 2: Move User-Related Code
1. Move user handlers to `user-service/internal/handlers/`
2. Move user models to `user-service/internal/models/`
3. Move auth utilities (JWT, password) to `user-service/pkg/utils/`
4. Create user repository for database operations

#### Step 3: Create User Service Main
```go
// user-service/cmd/main.go
package main

import (
    "log"
    "user-service/config"
    "user-service/internal/handlers"
    "user-service/internal/repository"
    "user-service/internal/service"
    "github.com/gin-gonic/gin"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize database
    db := database.Init(cfg.DatabaseURL)
    
    // Initialize Redis
    redisClient := redis.NewClient(&redis.Options{
        Addr: cfg.RedisAddr,
    })
    
    // Initialize repositories
    userRepo := repository.NewUserRepository(db, redisClient)
    
    // Initialize services
    userService := service.NewUserService(userRepo)
    
    // Initialize handlers
    userHandler := handlers.NewUserHandler(userService)
    
    // Setup routes
    router := gin.Default()
    api := router.Group("/api/v1")
    {
        api.POST("/register", userHandler.Register)
        api.POST("/login", userHandler.Login)
        api.GET("/users/:id", userHandler.GetUser)
        // Add more routes
    }
    
    log.Fatal(router.Run(":" + cfg.Port))
}
```

### Phase 3: Extracting Notification Service

#### Step 1: Create Notification Service Structure
```bash
mkdir -p microservices-project/notification-service/{cmd,internal/{handlers,models,service,kafka},config,templates}
```

#### Step 2: Move Notification-Related Code
1. Move email, SMS, OTP utilities
2. Create notification models
3. Set up Kafka consumer/producer

#### Step 3: Create Notification Service Main
```go
// notification-service/cmd/main.go
package main

import (
    "log"
    "notification-service/config"
    "notification-service/internal/kafka"
    "notification-service/internal/service"
)

func main() {
    cfg := config.Load()
    
    // Initialize services
    emailService := service.NewEmailService(cfg)
    smsService := service.NewSMSService(cfg)
    
    // Initialize Kafka consumer
    consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.ConsumerGroup)
    
    // Start consuming messages
    go consumer.ConsumeNotifications(emailService, smsService)
    
    // Keep the service running
    select {}
}
```

## Redis Caching

### Why Redis?
- In-memory data store for fast access
- Reduces database load
- Improves response times
- Supports various data structures

### Implementation in User Service:

```go
// user-service/internal/repository/user_repository.go
package repository

import (
    "context"
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
)

type UserRepository struct {
    db    *gorm.DB
    redis *redis.Client
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
    // Try to get from cache first
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // Check Redis
    cached, err := r.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var user models.User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }
    
    // If not in cache, get from database
    var user models.User
    if err := r.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    
    // Cache the result
    userData, _ := json.Marshal(user)
    r.redis.Set(ctx, cacheKey, userData, 15*time.Minute)
    
    return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
    // Update database
    if err := r.db.Save(user).Error; err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    r.redis.Del(ctx, cacheKey)
    
    return nil
}
```

### Caching Strategies:
1. **Cache-Aside**: Check cache first, if miss, get from DB and cache
2. **Write-Through**: Write to cache and DB simultaneously
3. **Write-Behind**: Write to cache immediately, write to DB later
4. **TTL (Time To Live)**: Set expiration for cached data

## Kafka Communication

### Why Kafka?
- Distributed messaging system
- High throughput
- Fault-tolerant
- Scalable
- Decouples services

### Setting up Kafka Communication:

#### Producer (User Service):
```go
// user-service/internal/kafka/producer.go
package kafka

import (
    "encoding/json"
    "github.com/Shopify/sarama"
)

type Producer struct {
    producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
    config := sarama.NewConfig()
    config.Producer.Return.Successes = true
    config.Producer.RequiredAcks = sarama.WaitForAll
    
    producer, err := sarama.NewSyncProducer(brokers, config)
    if err != nil {
        return nil, err
    }
    
    return &Producer{producer: producer}, nil
}

func (p *Producer) SendNotification(topic string, notification interface{}) error {
    data, err := json.Marshal(notification)
    if err != nil {
        return err
    }
    
    msg := &sarama.ProducerMessage{
        Topic: topic,
        Value: sarama.StringEncoder(data),
    }
    
    _, _, err = p.producer.SendMessage(msg)
    return err
}
```

#### Consumer (Notification Service):
```go
// notification-service/internal/kafka/consumer.go
package kafka

import (
    "context"
    "encoding/json"
    "log"
    "github.com/Shopify/sarama"
)

type Consumer struct {
    consumer sarama.ConsumerGroup
}

func NewConsumer(brokers []string, groupID string) (*Consumer, error) {
    config := sarama.NewConfig()
    config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
    
    consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
    if err != nil {
        return nil, err
    }
    
    return &Consumer{consumer: consumer}, nil
}

func (c *Consumer) ConsumeNotifications(ctx context.Context, topics []string) error {
    handler := &NotificationHandler{
        // Initialize with services
    }
    
    for {
        err := c.consumer.Consume(ctx, topics, handler)
        if err != nil {
            log.Printf("Error consuming: %v", err)
        }
    }
}
```

### Message Examples:

#### User Registration Event:
```json
{
    "event_type": "USER_REGISTERED",
    "user_id": 123,
    "email": "user@example.com",
    "name": "John Doe",
    "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Password Reset Event:
```json
{
    "event_type": "PASSWORD_RESET_REQUESTED",
    "user_id": 123,
    "email": "user@example.com",
    "reset_token": "abc123",
    "timestamp": "2024-01-15T10:35:00Z"
}
```

## Docker Configuration

### User Service Dockerfile:
```dockerfile
# user-service/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8001
CMD ["./main"]
```

### Docker Compose:
```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: userdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000

  kafka:
    image: confluentinc/cp-kafka:7.4.0
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1

  user-service:
    build: ./user-service
    ports:
      - "8001:8001"
    depends_on:
      - postgres
      - redis
      - kafka
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: user
      DB_PASSWORD: password
      DB_NAME: userdb
      REDIS_ADDR: redis:6379
      KAFKA_BROKERS: kafka:9092

  notification-service:
    build: ./notification-service
    ports:
      - "8002:8002"
    depends_on:
      - kafka
    environment:
      KAFKA_BROKERS: kafka:9092
      CONSUMER_GROUP: notification-group

  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - user-service
      - notification-service
    environment:
      USER_SERVICE_URL: http://user-service:8001
      NOTIFICATION_SERVICE_URL: http://notification-service:8002

volumes:
  postgres_data:
  redis_data:
```

## Recommended Resources

### YouTube Tutorials:
1. **"Microservices in Go" by Nic Jackson** - Comprehensive series on building microservices
2. **"Building Microservices with Go" by TechWorld with Nana** - Beginner-friendly approach
3. **"Kafka Tutorial" by Confluent** - Official Kafka tutorials
4. **"Redis Crash Course" by Traversy Media** - Quick Redis introduction

### Books:
1. **"Building Microservices" by Sam Newman** - The definitive guide
2. **"Microservices Patterns" by Chris Richardson** - Patterns and best practices
3. **"Go Microservices" by Nic Jackson** - Go-specific implementation

### Online Courses:
1. **Udemy: "Microservices with Go"** - Hands-on course
2. **Coursera: "Microservices Specialization"** - Academic approach
3. **Pluralsight: "Building Microservices"** - Enterprise-focused

### Key Concepts to Master:
1. **Service Discovery**: How services find each other
2. **Circuit Breakers**: Handling service failures
3. **API Gateway Pattern**: Single entry point
4. **Event-Driven Architecture**: Using Kafka effectively
5. **Distributed Tracing**: Monitoring requests across services
6. **Container Orchestration**: Kubernetes basics

## Next Steps:

1. **Start Small**: Begin with extracting the User Service
2. **Test Locally**: Use Docker Compose for local development
3. **Add Monitoring**: Implement logging and metrics
4. **Handle Failures**: Add retry logic and circuit breakers
5. **Security**: Implement service-to-service authentication
6. **Documentation**: Keep API documentation updated

Remember: Microservices add complexity. Make sure the benefits outweigh the costs for your use case.