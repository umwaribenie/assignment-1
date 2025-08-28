# Microservices Architecture - User Management System

This project demonstrates a microservices architecture built with Go, featuring user management and notification services with Redis caching and Kafka messaging.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐
│   User Service  │    │Notification Svc │
│     :8081       │    │     :8082       │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          ├──────────────────────┤
          │                      │
┌─────────▼───────┐    ┌─────────▼───────┐
│   PostgreSQL    │    │     Redis       │
│     :5432       │    │     :6379       │
└─────────────────┘    └─────────────────┘
          │                      │
          └──────────┬───────────┘
                     │
           ┌─────────▼───────┐
           │     Kafka       │
           │     :9092       │
           └─────────────────┘
```

## 🚀 Services

### 1. User Service (Port 8081)
- **Responsibilities:**
  - User registration and authentication
  - JWT token management
  - User profile management
  - Password reset workflow
  - User data caching with Redis

- **Key Features:**
  - Redis caching for improved performance
  - JWT-based authentication
  - Password hashing with bcrypt
  - Event publishing to Kafka
  - Rate limiting protection

### 2. Notification Service (Port 8082)
- **Responsibilities:**
  - Email sending (SMTP)
  - SMS sending (Twilio integration)
  - Template management
  - Event processing from Kafka
  - Notification rate limiting

- **Key Features:**
  - Template-based email/SMS
  - Kafka event consumption
  - Rate limiting per recipient
  - Redis caching for templates
  - Async notification processing

## 🔧 Technology Stack

- **Language:** Go 1.21+
- **Framework:** Gin (HTTP router)
- **Database:** PostgreSQL
- **Cache:** Redis
- **Message Broker:** Apache Kafka
- **Containerization:** Docker & Docker Compose
- **Email:** SMTP (Gmail/Custom)
- **SMS:** Twilio (configurable)

## 📦 Setup Instructions

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- Git

### 1. Clone and Setup
```bash
git clone <your-repo>
cd microservices

# Update environment variables
cp user-service/.env.example user-service/.env
cp notification-service/.env.example notification-service/.env
```

### 2. Configure Environment Variables

#### User Service (.env)
```env
USER_SERVICE_PORT=8081
DATABASE_URL=postgres://postgres:password@localhost:5432/userdb?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRY_HOURS=24
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
KAFKA_BROKERS=localhost:9092
```

#### Notification Service (.env)
```env
NOTIFICATION_SERVICE_PORT=8082
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=1
KAFKA_BROKERS=localhost:9092

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=Your Application

# SMS Configuration (Optional)
TWILIO_ACCOUNT_SID=your-twilio-account-sid
TWILIO_AUTH_TOKEN=your-twilio-auth-token
TWILIO_FROM_NUMBER=+1234567890
```

### 3. Start Services with Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Check service health
docker-compose ps
```

### 4. Verify Setup
```bash
# Check User Service
curl http://localhost:8081/health

# Check Notification Service
curl http://localhost:8082/health

# Access Kafka UI (optional)
open http://localhost:8080
```

## 📚 API Documentation

### User Service API (Port 8081)

#### Authentication
```bash
# Register User
POST /api/v1/users
Content-Type: application/json
{
  "client_id": "client123",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "password": "securepassword",
  "phone": "+1234567890"
}

# Login
POST /api/v1/auth/login
Content-Type: application/json
{
  "username": "user@example.com",
  "password": "securepassword"
}

# Request Password Reset
POST /api/v1/auth/password-reset/request
Content-Type: application/json
{
  "username": "user@example.com",
  "client_id": "client123"
}
```

#### User Management (Requires JWT)
```bash
# Get All Users
GET /api/v1/users
Authorization: Bearer <jwt-token>

# Get User by ID
GET /api/v1/users/{id}
Authorization: Bearer <jwt-token>

# Logout
POST /api/v1/users/logout
Authorization: Bearer <jwt-token>
```

### Notification Service API (Port 8082)

#### Send Notifications
```bash
# Send Email
POST /api/v1/notifications/email
Content-Type: application/json
{
  "to": "recipient@example.com",
  "subject": "Test Email",
  "content": "Hello World!",
  "is_html": false
}

# Send Template Email
POST /api/v1/notifications/email/template
Content-Type: application/json
{
  "to": "recipient@example.com",
  "template_id": "welcome",
  "variables": {
    "FirstName": "John",
    "AppName": "Your App"
  }
}

# Send SMS
POST /api/v1/notifications/sms
Content-Type: application/json
{
  "to": "+1234567890",
  "content": "Hello via SMS!"
}
```

#### Template Management
```bash
# Get Email Templates
GET /api/v1/notifications/templates/email

# Get SMS Templates
GET /api/v1/notifications/templates/sms
```

## 🔄 Kafka Events

### Event Flow
1. **User Registration** → `user-events` topic → Welcome email sent
2. **Password Reset** → `notification-events` topic → Reset email sent
3. **OTP Request** → `notification-events` topic → OTP email/SMS sent

### Topics
- `user-events`: User lifecycle events
- `notification-events`: Notification requests

## 🔍 Monitoring and Debugging

### Health Checks
```bash
# User Service
curl http://localhost:8081/health

# Notification Service
curl http://localhost:8082/health
```

### Logs
```bash
# View all logs
docker-compose logs -f

# Service-specific logs
docker-compose logs -f user-service
docker-compose logs -f notification-service
```

### Kafka Monitoring
- Kafka UI: http://localhost:8080
- Monitor topics, messages, and consumer groups

### Redis Monitoring
```bash
# Connect to Redis CLI
docker exec -it redis_cache redis-cli

# View cached keys
127.0.0.1:6379> KEYS *
```

## 🧪 Testing

### Manual Testing
```bash
# Create a user and verify welcome email
curl -X POST http://localhost:8081/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test123",
    "email": "test@example.com",
    "first_name": "Test",
    "last_name": "User",
    "password": "password123",
    "phone": "+1234567890"
  }'
```

### Load Testing
```bash
# Install wrk
brew install wrk  # macOS
apt-get install wrk  # Ubuntu

# Test user service
wrk -t12 -c400 -d30s http://localhost:8081/health
```

## 🔒 Security Considerations

### Production Checklist
- [ ] Change default JWT secret
- [ ] Use strong database passwords
- [ ] Enable Redis authentication
- [ ] Configure SSL/TLS certificates
- [ ] Set up firewall rules
- [ ] Enable audit logging
- [ ] Implement rate limiting
- [ ] Regular security updates

### Environment Variables
Never commit sensitive data:
- JWT secrets
- Database passwords
- SMTP credentials
- API keys

## 🔧 Development

### Local Development
```bash
# Start dependencies only
docker-compose up postgres redis kafka zookeeper -d

# Run services locally
cd user-service && go run main.go
cd notification-service && go run main.go
```

### Hot Reload (using air)
```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
cd user-service && air
cd notification-service && air
```

## 🚀 Deployment

### Production Deployment
1. **Container Registry**
   - Build and push images to registry
   - Update docker-compose with image references

2. **Environment Configuration**
   - Use Docker secrets or external secret management
   - Configure production databases

3. **Monitoring**
   - Add Prometheus metrics
   - Set up logging aggregation
   - Configure alerting

### Scaling
```bash
# Scale specific services
docker-compose up -d --scale user-service=3
docker-compose up -d --scale notification-service=2
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

For questions and support:
- Create an issue in this repository
- Check the troubleshooting section
- Review the API documentation

---

**Happy Coding! 🎉**