#!/bin/bash

# API Testing Script for Microservices
set -e

echo "🧪 Testing Microservices APIs..."

BASE_URL_USER="http://localhost:8081"
BASE_URL_NOTIFICATION="http://localhost:8082"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to test endpoint
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4
    local expected_status=${5:-200}
    
    echo -e "${BLUE}Testing: $description${NC}"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url")
    fi
    
    # Split response and status code
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC} - Status: $status_code"
        if [ -n "$body" ]; then
            echo "Response: $body" | jq . 2>/dev/null || echo "Response: $body"
        fi
    else
        echo -e "${RED}❌ FAIL${NC} - Expected: $expected_status, Got: $status_code"
        echo "Response: $body"
    fi
    echo ""
}

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 5

# Test health endpoints
echo "🏥 Testing Health Endpoints"
test_endpoint "GET" "$BASE_URL_USER/health" "" "User Service Health Check"
test_endpoint "GET" "$BASE_URL_NOTIFICATION/health" "" "Notification Service Health Check"

# Test user registration
echo "👤 Testing User Management"
USER_DATA='{
  "client_id": "test123",
  "email": "testuser@example.com",
  "first_name": "Test",
  "last_name": "User",
  "password": "password123",
  "phone": "+1234567890"
}'

test_endpoint "POST" "$BASE_URL_USER/api/v1/users" "$USER_DATA" "User Registration" 201

# Test login
echo "🔐 Testing Authentication"
LOGIN_DATA='{
  "username": "testuser@example.com",
  "password": "password123"
}'

login_response=$(curl -s -X POST "$BASE_URL_USER/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_DATA")

echo -e "${BLUE}Testing: User Login${NC}"
if echo "$login_response" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ PASS${NC} - Login successful"
    JWT_TOKEN=$(echo "$login_response" | jq -r '.data.access_token')
    echo "JWT Token obtained: ${JWT_TOKEN:0:50}..."
else
    echo -e "${RED}❌ FAIL${NC} - Login failed"
    echo "Response: $login_response"
    JWT_TOKEN=""
fi
echo ""

# Test protected endpoint (get users)
if [ -n "$JWT_TOKEN" ]; then
    echo "🔒 Testing Protected Endpoints"
    echo -e "${BLUE}Testing: Get All Users (Protected)${NC}"
    
    users_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL_USER/api/v1/users" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    status_code=$(echo "$users_response" | tail -n1)
    body=$(echo "$users_response" | sed '$d')
    
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}✅ PASS${NC} - Status: $status_code"
        echo "Response: $body" | jq . 2>/dev/null || echo "Response: $body"
    else
        echo -e "${RED}❌ FAIL${NC} - Expected: 200, Got: $status_code"
    fi
    echo ""
fi

# Test notification service
echo "📧 Testing Notification Service"

# Test email sending
EMAIL_DATA='{
  "to": "test@example.com",
  "subject": "Test Email",
  "content": "This is a test email from the microservices system.",
  "is_html": false
}'

test_endpoint "POST" "$BASE_URL_NOTIFICATION/api/v1/notifications/email" "$EMAIL_DATA" "Send Email"

# Test template email
TEMPLATE_EMAIL_DATA='{
  "to": "test@example.com",
  "template_id": "welcome",
  "variables": {
    "FirstName": "Test",
    "AppName": "Microservices Demo"
  }
}'

test_endpoint "POST" "$BASE_URL_NOTIFICATION/api/v1/notifications/email/template" "$TEMPLATE_EMAIL_DATA" "Send Template Email"

# Test SMS (will use mock implementation)
SMS_DATA='{
  "to": "+1234567890",
  "content": "Test SMS from microservices system"
}'

test_endpoint "POST" "$BASE_URL_NOTIFICATION/api/v1/notifications/sms" "$SMS_DATA" "Send SMS"

# Test getting templates
test_endpoint "GET" "$BASE_URL_NOTIFICATION/api/v1/notifications/templates/email" "" "Get Email Templates"
test_endpoint "GET" "$BASE_URL_NOTIFICATION/api/v1/notifications/templates/sms" "" "Get SMS Templates"

echo "🎉 API Testing Complete!"
echo ""
echo "📊 Summary:"
echo "   • User Service tests completed"
echo "   • Notification Service tests completed"
echo "   • Check above for any failed tests"
echo ""
echo "💡 Tips:"
echo "   • Check logs with: docker-compose logs -f"
echo "   • Monitor Kafka with: http://localhost:8080"
echo "   • View Redis data with: docker exec -it redis_cache redis-cli"