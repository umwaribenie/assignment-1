# Microservices User Management System

A microservices-based user management system built with Go, featuring user authentication, notification services, Redis caching, and Kafka messaging.

## Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
┌──────▼──────┐
│ API Gateway │ (Nginx - Port 8080)
└──────┬──────┘
       │
┌──────┴───────────────┬──────────────┐
│                      │              │
▼                      ▼              │
┌──────────────┐  ┌──────────────┐   │
│ User Service │  │ Notification │   │
│ (Port 8001)  │  │   Service    │   │
│              │  │ (Port 8002)  │   │
│  [Redis]     │  └──────────────┘   │
│  [PostgreSQL]│                      │
└──────────────┘                      │
       │                              │
       └──────────┬───────────────────┘
                  │
           ┌──────▼──────┐
           │   Kafka     │
           └─────────────┘
```

## Services

### 1. User Service (Port 8001)
- User registration and authentication
- JWT token generation and validation
- User profile management
- Password management
- Redis caching for improved performance

### 2. Notification Service (Port 8002)
- Email notifications (SMTP)
- SMS notifications (Twilio - optional)
- Consumes events from Kafka
- Template-based notifications

### 3. API Gateway (Port 8080)
- Single entry point for all client requests
- Request routing
- Load balancing (if multiple instances)
- Currently using Nginx

## Technology Stack

- **Language**: Go 1.21
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Broker**: Apache Kafka
- **API Gateway**: Nginx
- **Containerization**: Docker & Docker Compose

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for local development)
- SMTP credentials (for email notifications)
- Twilio credentials (optional, for SMS)

## Getting Started

### 1. Clone the Repository

```bash
cd microservices-project
```

### 2. Environment Configuration

Update the environment variables in `docker-compose.yml` or create `.env` files for each service:

#### User Service (.env)
```env
PORT=8001
JWT_SECRET=your-secret-key-here
JWT_EXPIRY_HOURS=24
DB_HOST=postgres
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=userdb
REDIS_ADDR=redis:6379
KAFKA_BROKERS=kafka:29092
```

#### Notification Service (.env)
```env
PORT=8002
KAFKA_BROKERS=kafka:29092
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourcompany.com
```

### 3. Start All Services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL (Port 5432)
- Redis (Port 6379)
- Kafka & Zookeeper
- User Service (Port 8001)
- Notification Service (Port 8002)
- API Gateway (Port 8080)
- Kafka UI (Port 8090) - for monitoring
- Redis Commander (Port 8081) - Redis GUI
- pgAdmin (Port 5050) - PostgreSQL GUI

### 4. Verify Services

Check if all services are running:
```bash
docker-compose ps
```

Health check endpoints:
- API Gateway: http://localhost:8080/health
- User Service: http://localhost:8001/health
- Notification Service: http://localhost:8002/health

## API Endpoints

### Authentication

#### Register User
```bash
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "+1234567890"
}
```

#### Login
```bash
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

### User Management

#### Get User Profile
```bash
GET http://localhost:8080/api/v1/users/{user_id}
Authorization: Bearer {jwt_token}
```

#### Update User Profile
```bash
PUT http://localhost:8080/api/v1/users/{user_id}
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Updated",
  "phone_number": "+0987654321"
}
```

#### Change Password
```bash
POST http://localhost:8080/api/v1/users/{user_id}/change-password
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "old_password": "currentpassword",
  "new_password": "newpassword"
}
```

#### Get All Users (Admin only)
```bash
GET http://localhost:8080/api/v1/users?pageNumber=1&pageSize=10&role=user&status=active
Authorization: Bearer {admin_jwt_token}
```

## Monitoring Tools

### Kafka UI
Access at: http://localhost:8090
- Monitor topics
- View messages
- Check consumer groups

### Redis Commander
Access at: http://localhost:8081
- Browse Redis keys
- Monitor cache performance
- View/Edit cached data

### pgAdmin
Access at: http://localhost:5050
- Login: admin@admin.com / admin
- Manage PostgreSQL database
- Run queries

## Development

### Running Services Locally

1. **User Service**:
```bash
cd user-service
go mod download
go run cmd/main.go
```

2. **Notification Service**:
```bash
cd notification-service
go mod download
go run cmd/main.go
```

### Building Docker Images

```bash
# Build all services
docker-compose build

# Build specific service
docker-compose build user-service
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f user-service
```

## Testing

### Manual Testing with cURL

1. **Register a new user**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

2. **Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## Kafka Topics

The system uses the following Kafka topics:
- `notifications`: All notification events (user registration, password reset, etc.)

### Viewing Kafka Messages

You can use Kafka UI at http://localhost:8090 or use the command line:

```bash
# List topics
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092

# Consume messages
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic notifications \
  --from-beginning
```

## Redis Caching Strategy

The User Service implements caching for:
- User profiles (by ID, email, username)
- TTL: 15 minutes
- Cache invalidation on updates

### Cache Keys Pattern
- User by ID: `user:id:{uuid}`
- User by Email: `user:email:{email}`
- User by Username: `user:username:{username}`

## Troubleshooting

### Common Issues

1. **Services not starting**: Check logs with `docker-compose logs [service-name]`
2. **Database connection issues**: Ensure PostgreSQL is healthy before starting User Service
3. **Kafka connection issues**: Wait for Kafka to be fully initialized
4. **Email not sending**: Verify SMTP credentials in Notification Service

### Reset Everything

```bash
# Stop all services and remove volumes
docker-compose down -v

# Remove all images
docker-compose down --rmi all
```

## Production Considerations

1. **Security**:
   - Use strong JWT secrets
   - Enable SSL/TLS
   - Implement rate limiting
   - Use secure password policies

2. **Scaling**:
   - Use Kubernetes for orchestration
   - Implement service mesh (Istio/Linkerd)
   - Add more Kafka partitions
   - Use Redis Cluster

3. **Monitoring**:
   - Add Prometheus metrics
   - Implement distributed tracing (Jaeger)
   - Use ELK stack for logging
   - Set up alerts

4. **High Availability**:
   - Multiple instances of each service
   - Database replication
   - Redis Sentinel
   - Kafka cluster

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.