# Microservices Architecture - User Management System

This project demonstrates a microservices architecture with two main services: **User Service** and **Notification Service**, built with Go, using Redis for caching, Kafka for event-driven communication, and PostgreSQL for data persistence.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   User Service  │    │ Notification     │    │   PostgreSQL    │
│   (Port: 8081)  │    │ Service          │    │   Database      │
│                 │    │ (Port: 8082)     │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│      Redis      │    │      Kafka       │    │   Kafka UI      │
│   (Port: 6379)  │    │   (Port: 9092)   │    │  (Port: 8080)   │
│     Cache       │    │   Event Bus      │    │   Monitoring    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development)

### Running with Docker Compose

1. **Clone and navigate to the project:**
   ```bash
   cd microservices
   ```

2. **Start all services:**
   ```bash
   docker-compose up -d
   ```

3. **Check service status:**
   ```bash
   docker-compose ps
   ```

4. **View logs:**
   ```bash
   # All services
   docker-compose logs -f
   
   # Specific service
   docker-compose logs -f user-service
   docker-compose logs -f notification-service
   ```

### Access Points

- **User Service API**: http://localhost:8081
- **Notification Service API**: http://localhost:8082
- **Kafka UI**: http://localhost:8080
- **Redis Commander**: http://localhost:8083
- **PostgreSQL**: localhost:5432

## 📋 Service Endpoints

### User Service (Port 8081)

#### User Management
- `POST /api/v1/users/` - Create a new user
- `GET /api/v1/users/` - Get all users (with pagination and filtering)
- `GET /api/v1/users/{id}` - Get user by ID
- `GET /api/v1/users/email?email={email}` - Get user by email
- `GET /api/v1/users/username?username={username}` - Get user by username
- `GET /api/v1/users/slug/{slug}` - Get user by slug
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user
- `PUT /api/v1/users/{id}/password` - Update user password

#### Admin Operations
- `POST /api/v1/users/admin` - Create user by admin

#### Health Check
- `GET /api/v1/health` - Service health status

### Notification Service (Port 8082)

#### Notification Management
- `POST /api/v1/notifications/` - Create a new notification
- `GET /api/v1/notifications/` - Get all notifications (with pagination and filtering)
- `GET /api/v1/notifications/{id}` - Get notification by ID
- `GET /api/v1/notifications/user/{userId}` - Get notifications by user ID
- `POST /api/v1/notifications/retry` - Retry failed notifications

#### Health Check
- `GET /api/v1/health` - Service health status

## 🔧 Configuration

### Environment Variables

#### User Service (.env)
```env
PORT=8081
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=12345
DB_NAME=user_service
REDIS_ADDR=localhost:6379
KAFKA_BROKERS=localhost:9092
JWT_SECRET_KEY=your_jwt_secret
```

#### Notification Service (.env)
```env
PORT=8082
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=12345
DB_NAME=notification_service
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_USER_EVENTS=user-events
KAFKA_GROUP_ID=notification-service
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
```

## 📊 API Examples

### Create a User
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

### Get All Users
```bash
curl "http://localhost:8081/api/v1/users/?page=1&size=10&role=user&status=active"
```

### Create a Notification
```bash
curl -X POST http://localhost:8082/api/v1/notifications/ \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-id-here",
    "type": "email",
    "subject": "Welcome!",
    "message": "Welcome to our platform!",
    "recipient": "user@example.com",
    "template": "welcome_email"
  }'
```

## 🔄 Event Flow

1. **User Creation Event:**
   - User Service creates a user
   - Publishes `user.created` event to Kafka
   - Notification Service consumes the event
   - Creates welcome email and SMS notifications

2. **User Update Event:**
   - User Service updates a user
   - Publishes `user.updated` event to Kafka
   - Notification Service creates profile update notification

3. **User Status Change Event:**
   - User Service changes user status
   - Publishes `user.status_changed` event to Kafka
   - Notification Service creates status change notification

## 🗄️ Database Schema

### User Service Database
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    national_id VARCHAR(255) UNIQUE,
    passport_number VARCHAR(255) UNIQUE,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(255) UNIQUE NOT NULL,
    profile_picture TEXT,
    username VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

### Notification Service Database
```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    subject TEXT,
    message TEXT NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    template VARCHAR(100),
    variables JSONB,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    sent_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    error_msg TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

## 🧪 Testing

### Manual Testing
1. Start the services using Docker Compose
2. Use the API endpoints with tools like curl, Postman, or the Swagger UI
3. Monitor logs: `docker-compose logs -f`

### Integration Testing
```bash
# Test User Service
curl http://localhost:8081/api/v1/health

# Test Notification Service
curl http://localhost:8082/api/v1/health

# Test Kafka connectivity
# Access Kafka UI at http://localhost:8080
```

## 📈 Monitoring

### Kafka UI
- Access: http://localhost:8080
- Monitor topics, messages, and consumer groups
- View message content and metadata

### Redis Commander
- Access: http://localhost:8083
- Monitor Redis cache operations
- View cached data and keys

### Service Health Checks
- User Service: http://localhost:8081/api/v1/health
- Notification Service: http://localhost:8082/api/v1/health

## 🏗️ Project Structure

```
microservices/
├── user-service/
│   ├── cmd/main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── models/
│   │   ├── repository/
│   │   └── service/
│   ├── pkg/
│   │   ├── kafka/
│   │   └── utils/
│   ├── Dockerfile
│   └── .env
├── notification-service/
│   ├── cmd/main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── kafka/
│   │   ├── models/
│   │   ├── repository/
│   │   └── service/
│   ├── Dockerfile
│   └── .env
├── shared/
│   ├── models/
│   └── kafka/
├── database/
│   └── init.sql
├── docker-compose.yml
└── README.md
```

## 🔑 Key Features

### User Service
- ✅ User CRUD operations
- ✅ Redis caching for performance
- ✅ Kafka event publishing
- ✅ JWT authentication (ready for implementation)
- ✅ Pagination and filtering
- ✅ Swagger API documentation

### Notification Service
- ✅ Email notifications (SMTP)
- ✅ SMS notifications (simulated)
- ✅ Push notifications (simulated)
- ✅ Kafka event consumption
- ✅ Retry mechanism for failed notifications
- ✅ Template support
- ✅ Swagger API documentation

### Infrastructure
- ✅ PostgreSQL with separate databases
- ✅ Redis for caching
- ✅ Kafka for event-driven communication
- ✅ Docker containerization
- ✅ Health checks
- ✅ Monitoring tools

## 🚀 Deployment

### Production Considerations
1. **Security:**
   - Use strong JWT secrets
   - Enable HTTPS
   - Implement proper authentication/authorization
   - Secure database connections

2. **Scalability:**
   - Use connection pooling
   - Implement rate limiting
   - Consider horizontal scaling
   - Use load balancers

3. **Monitoring:**
   - Implement proper logging
   - Add metrics collection
   - Set up alerting
   - Use APM tools

4. **Backup:**
   - Regular database backups
   - Configuration backups
   - Disaster recovery plan

## 🛠️ Development

### Local Development
1. **Clone the repository**
2. **Install dependencies:**
   ```bash
   cd user-service && go mod download
   cd ../notification-service && go mod download
   ```

3. **Set up local environment:**
   - Install PostgreSQL, Redis, and Kafka locally
   - Or use Docker for infrastructure services

4. **Run services:**
   ```bash
   # User Service
   cd user-service && go run cmd/main.go
   
   # Notification Service
   cd notification-service && go run cmd/main.go
   ```

### Adding New Features
1. **User Service:**
   - Add new models in `internal/models/`
   - Implement repository methods in `internal/repository/`
   - Add business logic in `internal/service/`
   - Create handlers in `internal/handlers/`
   - Update routes in `cmd/main.go`

2. **Notification Service:**
   - Follow the same pattern as User Service
   - Add new Kafka event handlers in `internal/kafka/`

## 📚 Learning Resources

### Microservices
- [Microservices.io](https://microservices.io/)
- [Martin Fowler's Blog](https://martinfowler.com/articles/microservices.html)

### Go
- [Go Documentation](https://golang.org/doc/)
- [Go Web Examples](https://gowebexamples.com/)

### Kafka
- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Kafka Tutorial](https://kafka.apache.org/quickstart)

### Redis
- [Redis Documentation](https://redis.io/documentation)
- [Redis Tutorial](https://redis.io/topics/tutorial)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Troubleshooting

### Common Issues

1. **Services not starting:**
   ```bash
   docker-compose logs -f
   # Check for dependency issues
   ```

2. **Database connection errors:**
   ```bash
   # Check if PostgreSQL is running
   docker-compose ps postgres
   ```

3. **Kafka connection issues:**
   ```bash
   # Check Kafka UI at http://localhost:8080
   # Verify topics are created
   ```

4. **Redis connection issues:**
   ```bash
   # Check Redis Commander at http://localhost:8083
   # Verify Redis is accessible
   ```

### Reset Everything
```bash
# Stop and remove all containers
docker-compose down -v

# Remove all images
docker-compose down --rmi all

# Start fresh
docker-compose up -d
```