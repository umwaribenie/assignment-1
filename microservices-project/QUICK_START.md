# Quick Start Guide - Microservices Project

## 🚀 Get Started in 5 Minutes

### Prerequisites
- Docker and Docker Compose installed
- Port 8080, 8001, 8002 available

### Step 1: Clone and Navigate
```bash
cd microservices-project
```

### Step 2: Start All Services
```bash
docker-compose up -d
```

Wait about 30 seconds for all services to initialize.

### Step 3: Verify Services
```bash
# Check health
curl http://localhost:8080/health
```

### Step 4: Create Your First User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "username": "johndoe",
    "password": "securepass123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Step 5: Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

Save the JWT token from the response!

### Step 6: Access Your Profile
```bash
# Replace YOUR_JWT_TOKEN with the token from login response
curl http://localhost:8080/api/v1/users/YOUR_USER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 📊 Monitoring

- **Kafka UI**: http://localhost:8090
- **Redis Commander**: http://localhost:8081
- **pgAdmin**: http://localhost:5050 (admin@admin.com/admin)

## 🛠️ Development Helper

Use the development script for easy management:
```bash
./dev.sh
```

Or use direct commands:
```bash
./dev.sh start      # Start all services
./dev.sh stop       # Stop all services
./dev.sh logs       # View all logs
./dev.sh test-user  # Create a test user
```

## 🔧 Troubleshooting

If services don't start:
```bash
# Check logs
docker-compose logs

# Restart everything
docker-compose down
docker-compose up -d
```

## 📝 Next Steps

1. Check the notification service received the registration event in Kafka UI
2. Configure SMTP settings to receive actual emails
3. Explore the API endpoints in the main README
4. Try updating user profiles and changing passwords

Happy coding! 🎉