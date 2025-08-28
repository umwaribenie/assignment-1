# Step-by-Step Microservices Guide

## 🎯 What You'll Learn

This guide will walk you through:
1. Understanding microservices architecture
2. Setting up Redis caching
3. Implementing Kafka event-driven communication
4. Running the complete system
5. Testing the functionality

## 📚 Understanding Microservices

### What are Microservices?

Microservices is an architectural style where an application is built as a collection of small, independent services. Each service:
- **Has its own database** (Database per Service pattern)
- **Can be developed, deployed, and scaled independently**
- **Communicates via APIs or message queues**
- **Has a single responsibility**

### Why Microservices?

**Benefits:**
- ✅ **Scalability**: Scale services independently
- ✅ **Fault Isolation**: One service failure doesn't affect others
- ✅ **Technology Diversity**: Use different tech stacks per service
- ✅ **Team Autonomy**: Teams can work independently
- ✅ **Deployment Flexibility**: Deploy services independently

**Challenges:**
- ❌ **Complexity**: More moving parts to manage
- ❌ **Network Latency**: Service-to-service communication overhead
- ❌ **Data Consistency**: Distributed data management
- ❌ **Testing**: More complex integration testing

## 🏗️ Our Architecture

### Services Overview

1. **User Service (Port 8081)**
   - Manages user data (CRUD operations)
   - Uses Redis for caching
   - Publishes events to Kafka

2. **Notification Service (Port 8082)**
   - Handles notifications (email, SMS, push)
   - Consumes events from Kafka
   - Manages notification lifecycle

### Communication Patterns

1. **Synchronous**: HTTP/REST APIs for direct communication
2. **Asynchronous**: Kafka for event-driven communication
3. **Caching**: Redis for performance optimization

## 🚀 Step 1: Understanding the Code Structure

### User Service Structure
```
user-service/
├── cmd/main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   ├── models/              # Data models
│   ├── repository/          # Database operations
│   └── service/             # Business logic
├── pkg/
│   ├── kafka/              # Kafka producer
│   └── utils/              # Cache utilities
└── go.mod                  # Dependencies
```

### Notification Service Structure
```
notification-service/
├── cmd/main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   ├── models/              # Data models
│   ├── repository/          # Database operations
│   ├── service/             # Business logic
│   └── kafka/              # Kafka consumer
└── go.mod                  # Dependencies
```

## 🔧 Step 2: Setting Up the Environment

### Prerequisites Installation

1. **Install Docker and Docker Compose:**
   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install docker.io docker-compose
   
   # macOS
   brew install docker docker-compose
   
   # Windows
   # Download Docker Desktop from https://www.docker.com/products/docker-desktop
   ```

2. **Install Go:**
   ```bash
   # Ubuntu/Debian
   sudo apt install golang-go
   
   # macOS
   brew install go
   
   # Windows
   # Download from https://golang.org/dl/
   ```

3. **Verify installations:**
   ```bash
   docker --version
   docker-compose --version
   go version
   ```

## 🐳 Step 3: Running with Docker Compose

### Quick Start (Recommended)

1. **Navigate to the microservices directory:**
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
   # View all logs
   docker-compose logs -f
   
   # View specific service logs
   docker-compose logs -f user-service
   docker-compose logs -f notification-service
   ```

### What's Running?

After running `docker-compose up -d`, you'll have:

| Service | Port | Description |
|---------|------|-------------|
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache |
| Kafka | 9092 | Message broker |
| Zookeeper | 2181 | Kafka coordination |
| User Service | 8081 | User management API |
| Notification Service | 8082 | Notification API |
| Kafka UI | 8080 | Kafka monitoring |
| Redis Commander | 8083 | Redis monitoring |

## 🧪 Step 4: Testing the System

### Test 1: Health Checks

1. **Check User Service:**
   ```bash
   curl http://localhost:8081/api/v1/health
   ```
   Expected response:
   ```json
   {
     "status": "healthy",
     "service": "user-service",
     "timestamp": 1234567890
   }
   ```

2. **Check Notification Service:**
   ```bash
   curl http://localhost:8082/api/v1/health
   ```

### Test 2: Create a User

1. **Create a user:**
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

2. **Expected response:**
   ```json
   {
     "message": "User created successfully",
     "data": {
       "id": "uuid-here",
       "email": "john.doe@example.com",
       "first_name": "John",
       "last_name": "Doe",
       "role": "user",
       "status": "active"
     }
   }
   ```

### Test 3: Check Notifications

1. **Get all notifications:**
   ```bash
   curl "http://localhost:8082/api/v1/notifications"
   ```

2. **You should see a welcome notification created automatically!**

### Test 4: Update User Status

1. **Update user status:**
   ```bash
   curl -X PUT http://localhost:8081/api/v1/users/{user-id} \
     -H "Content-Type: application/json" \
     -d '{
       "status": "suspended"
     }'
   ```

2. **Check notifications again:**
   ```bash
   curl "http://localhost:8082/api/v1/notifications"
   ```

3. **You should see a status change notification!**

## 🔍 Step 5: Understanding the Event Flow

### Event-Driven Architecture

1. **User Creation Flow:**
   ```
   User Service → Creates User → Publishes "user.created" → Kafka → Notification Service → Creates Welcome Email
   ```

2. **User Update Flow:**
   ```
   User Service → Updates User → Publishes "user.updated" → Kafka → Notification Service → Creates Update Notification
   ```

3. **Status Change Flow:**
   ```
   User Service → Changes Status → Publishes "user.status_changed" → Kafka → Notification Service → Creates Status Notification
   ```

### Monitoring the Events

1. **Kafka UI (http://localhost:8080):**
   - View topics and messages
   - Monitor consumer groups
   - Check message details

2. **Redis Commander (http://localhost:8083):**
   - View cached data
   - Monitor Redis performance
   - Check cache keys

## 📊 Step 6: Understanding Caching

### Redis Caching Strategy

1. **User Cache Keys:**
   - `user:{user-id}` - Individual user data
   - `users:list:{filter}` - User list with filters
   - `users:count` - Total user count

2. **Cache Invalidation:**
   - When user is created/updated/deleted
   - Cache is automatically invalidated
   - Ensures data consistency

3. **Cache Benefits:**
   - Faster response times
   - Reduced database load
   - Better user experience

## 🔧 Step 7: Running Locally (Development)

### Prerequisites

1. **Start infrastructure only:**
   ```bash
   docker-compose up -d postgres redis kafka zookeeper
   ```

2. **Set up environment files:**
   ```bash
   cp user-service/.env.example user-service/.env
   cp notification-service/.env.example notification-service/.env
   ```

### Running User Service

1. **Navigate to user service:**
   ```bash
   cd user-service
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the service:**
   ```bash
   go run cmd/main.go
   ```

### Running Notification Service

1. **Open new terminal and navigate:**
   ```bash
   cd notification-service
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the service:**
   ```bash
   go run cmd/main.go
   ```

## 🐛 Step 8: Debugging and Troubleshooting

### Common Issues

1. **Kafka Connection Issues:**
   ```bash
   # Check if Kafka is running
   docker-compose ps kafka
   
   # Check Kafka logs
   docker-compose logs kafka
   
   # Restart Kafka
   docker-compose restart kafka
   ```

2. **Database Connection Issues:**
   ```bash
   # Check if PostgreSQL is running
   docker-compose ps postgres
   
   # Check database logs
   docker-compose logs postgres
   
   # Connect to database
   docker-compose exec postgres psql -U postgres -d user_service
   ```

3. **Redis Connection Issues:**
   ```bash
   # Check if Redis is running
   docker-compose ps redis
   
   # Check Redis logs
   docker-compose logs redis
   
   # Connect to Redis
   docker-compose exec redis redis-cli
   ```

### Log Analysis

1. **View all logs:**
   ```bash
   docker-compose logs -f
   ```

2. **View specific service logs:**
   ```bash
   docker-compose logs -f user-service
   docker-compose logs -f notification-service
   ```

3. **Search logs:**
   ```bash
   docker-compose logs | grep "error"
   docker-compose logs | grep "user.created"
   ```

## 📈 Step 9: Performance Monitoring

### Key Metrics to Monitor

1. **Response Times:**
   - API response times
   - Database query times
   - Cache hit rates

2. **Throughput:**
   - Requests per second
   - Messages per second (Kafka)
   - Database connections

3. **Error Rates:**
   - HTTP error rates
   - Kafka consumer lag
   - Database connection errors

### Monitoring Tools

1. **Built-in Health Checks:**
   ```bash
   curl http://localhost:8081/api/v1/health
   curl http://localhost:8082/api/v1/health
   ```

2. **Kafka UI:** http://localhost:8080
3. **Redis Commander:** http://localhost:8083

## 🚀 Step 10: Production Considerations

### Security

1. **Environment Variables:**
   - Use `.env` files for development only
   - Use Kubernetes secrets or AWS Secrets Manager in production
   - Never commit secrets to version control

2. **Network Security:**
   - Use HTTPS in production
   - Implement proper authentication/authorization
   - Use VPN or private networks

3. **Database Security:**
   - Use strong passwords
   - Enable SSL connections
   - Regular security updates

### Scaling

1. **Horizontal Scaling:**
   - Run multiple instances of each service
   - Use load balancers
   - Implement auto-scaling

2. **Database Scaling:**
   - Read replicas for read-heavy workloads
   - Database sharding for large datasets
   - Connection pooling

3. **Kafka Scaling:**
   - Multiple Kafka brokers
   - Topic partitioning
   - Consumer group scaling

### Monitoring and Alerting

1. **Logging:**
   - Centralized logging (ELK stack)
   - Structured logging
   - Log retention policies

2. **Metrics:**
   - Prometheus for metrics collection
   - Grafana for visualization
   - Custom business metrics

3. **Alerting:**
   - Service health alerts
   - Performance degradation alerts
   - Error rate alerts

## 🎓 Next Steps

### Learning Path

1. **Deepen Your Knowledge:**
   - Read "Building Microservices" by Sam Newman
   - Study event-driven architecture patterns
   - Learn about distributed systems

2. **Advanced Topics:**
   - Service mesh (Istio, Linkerd)
   - API Gateway patterns
   - Circuit breaker patterns
   - Saga pattern for distributed transactions

3. **Tools to Explore:**
   - Kubernetes for orchestration
   - Prometheus for monitoring
   - Jaeger for distributed tracing
   - Elasticsearch for logging

### Practice Projects

1. **Add Authentication Service:**
   - JWT token management
   - OAuth2 integration
   - Role-based access control

2. **Add Payment Service:**
   - Payment processing
   - Transaction management
   - Integration with payment gateways

3. **Add Analytics Service:**
   - User behavior tracking
   - Business metrics
   - Data visualization

## 📚 Additional Resources

### YouTube Tutorials
- "Microservices with Go" by Tech With Tim
- "Building Microservices with Go" by Golang Dojo
- "Kafka Tutorial for Beginners" by Confluent
- "Redis Tutorial for Beginners" by Redis University

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [Docker Documentation](https://docs.docker.com/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)
- [Redis Documentation](https://redis.io/documentation)

### Books
- "Building Microservices" by Sam Newman
- "Designing Data-Intensive Applications" by Martin Kleppmann
- "Go Programming Language" by Alan Donovan and Brian Kernighan

## 🎉 Congratulations!

You've successfully:
- ✅ Set up a complete microservices architecture
- ✅ Implemented Redis caching
- ✅ Configured Kafka event-driven communication
- ✅ Created a working system with two services
- ✅ Tested the functionality
- ✅ Understood the key concepts

You're now ready to build your own microservices applications!