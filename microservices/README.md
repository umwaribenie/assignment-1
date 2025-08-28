# Microservices Architecture with Go

This project demonstrates a microservices architecture using Go, featuring:

- **User Service**: Manages user data with Redis caching
- **Notification Service**: Handles notifications with Kafka event processing
- **Redis**: For caching and session management
- **Kafka**: For asynchronous communication between services
- **PostgreSQL**: Database for both services

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Service  │    │ Notification    │    │   PostgreSQL    │
│   (Port: 8081)  │    │ Service         │    │   Database      │
└─────────────────┘    │ (Port: 8082)    │    └─────────────────┘
         │              └─────────────────┘              │
         │                       │                       │
         │              ┌─────────────────┐              │
         └──────────────│      Kafka      │──────────────┘
                        │   (Port: 9092)  │
                        └─────────────────┘
                                │
                        ┌─────────────────┐
                        │      Redis      │
                        │   (Port: 6379)  │
                        └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24.2 or later
- Make (optional, for convenience)

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
   docker-compose logs -f user-service
   docker-compose logs -f notification-service
   ```

### Running Locally

1. **Start infrastructure services:**
   ```bash
   docker-compose up -d postgres redis kafka zookeeper
   ```

2. **Set up environment variables:**
   ```bash
   # User Service
   cp user-service/.env.example user-service/.env
   
   # Notification Service
   cp notification-service/.env.example notification-service/.env
   ```

3. **Run services:**
   ```bash
   # Terminal 1 - User Service
   cd user-service
   go mod tidy
   go run cmd/main.go
   
   # Terminal 2 - Notification Service
   cd notification-service
   go mod tidy
   go run cmd/main.go
   ```

## 📊 Service Endpoints

### User Service (Port 8081)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/users` | Get all users (with pagination) |
| POST | `/api/v1/users` | Create a new user |
| GET | `/api/v1/users/:id` | Get user by ID |
| GET | `/api/v1/users/email` | Get user by email |
| PUT | `/api/v1/users/:id` | Update user |
| DELETE | `/api/v1/users/:id` | Delete user |
| GET | `/swagger/*` | API documentation |

### Notification Service (Port 8082)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/notifications` | Get all notifications |
| POST | `/api/v1/notifications` | Create a new notification |
| GET | `/api/v1/notifications/:id` | Get notification by ID |
| GET | `/api/v1/notifications/user/:user_id` | Get notifications by user ID |
| GET | `/swagger/*` | API documentation |

## 🔧 Configuration

### Environment Variables

#### User Service
```env
PORT=8081
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=user_service
REDIS_ADDR=localhost:6379
KAFKA_BROKERS=localhost:9092
```

#### Notification Service
```env
PORT=8082
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=notification_service
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_USER_EVENTS=user-events
KAFKA_GROUP_ID=notification-service
```

## 📝 API Examples

### Create a User

```bash
curl -X POST http://localhost:8081/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client123",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "password": "password123",
    "phone": "+1234567890",
    "username": "johndoe",
    "role": "user"
  }'
```

### Get All Users

```bash
curl "http://localhost:8081/api/v1/users?pageNumber=1&pageSize=10&role=user"
```

### Get Notifications for a User

```bash
curl "http://localhost:8082/api/v1/notifications/user/{user-id}?pageNumber=1&pageSize=10"
```

## 🔄 Event Flow

1. **User Creation:**
   - User service creates a user
   - Publishes `user.created` event to Kafka
   - Notification service consumes the event
   - Creates welcome notification

2. **User Update:**
   - User service updates a user
   - Publishes `user.updated` event to Kafka
   - Notification service creates profile update notification

3. **Status Change:**
   - User service changes user status
   - Publishes `user.status_changed` event to Kafka
   - Notification service creates status change notification

## 🗄️ Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    -- ... other fields
);
```

### Notifications Table
```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(50) NOT NULL,
    retry_count INTEGER DEFAULT 0,
    -- ... other fields
);
```

## 🧪 Testing

### Manual Testing

1. **Create a user:**
   ```bash
   curl -X POST http://localhost:8081/api/v1/users \
     -H "Content-Type: application/json" \
     -d '{"client_id":"test","email":"test@example.com","first_name":"Test","last_name":"User","password":"password","phone":"+1234567890","username":"testuser","role":"user"}'
   ```

2. **Check notifications:**
   ```bash
   curl "http://localhost:8082/api/v1/notifications"
   ```

### Monitoring

- **Kafka UI:** http://localhost:8080
- **Redis Commander:** http://localhost:8083
- **User Service Swagger:** http://localhost:8081/swagger/index.html
- **Notification Service Swagger:** http://localhost:8082/swagger/index.html

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
│   └── go.mod
├── notification-service/
│   ├── cmd/main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── service/
│   │   └── kafka/
│   ├── Dockerfile
│   └── go.mod
├── shared/
│   ├── models/
│   ├── utils/
│   └── kafka/
├── database/
│   └── init.sql
├── docker-compose.yml
└── README.md
```

## 🔍 Key Features

### User Service
- ✅ User CRUD operations
- ✅ Redis caching for performance
- ✅ Kafka event publishing
- ✅ Pagination and filtering
- ✅ Input validation
- ✅ Error handling

### Notification Service
- ✅ Kafka event consumption
- ✅ Notification creation and management
- ✅ Retry mechanism for failed notifications
- ✅ Multiple notification types (email, SMS, push)
- ✅ Event-driven architecture

### Infrastructure
- ✅ PostgreSQL with proper indexing
- ✅ Redis for caching
- ✅ Kafka for event streaming
- ✅ Docker containerization
- ✅ Health checks
- ✅ API documentation (Swagger)

## 🚀 Deployment

### Production Considerations

1. **Security:**
   - Use environment variables for secrets
   - Implement proper authentication/authorization
   - Use HTTPS in production
   - Secure database connections

2. **Monitoring:**
   - Add logging (e.g., ELK stack)
   - Implement metrics (e.g., Prometheus)
   - Set up alerting
   - Monitor Kafka lag

3. **Scaling:**
   - Use Kubernetes for orchestration
   - Implement horizontal scaling
   - Use load balancers
   - Consider database sharding

4. **Backup:**
   - Regular database backups
   - Kafka topic replication
   - Disaster recovery plan

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Troubleshooting

### Common Issues

1. **Kafka connection issues:**
   - Ensure Zookeeper is running
   - Check Kafka broker configuration
   - Verify network connectivity

2. **Database connection issues:**
   - Check PostgreSQL is running
   - Verify database credentials
   - Ensure database exists

3. **Redis connection issues:**
   - Check Redis is running
   - Verify Redis configuration
   - Check network connectivity

### Logs

View service logs:
```bash
docker-compose logs -f [service-name]
```

### Health Checks

Check service health:
```bash
curl http://localhost:8081/api/v1/health
curl http://localhost:8082/api/v1/health
```