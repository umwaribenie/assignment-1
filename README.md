# General User Management System

A comprehensive user management system built with Go (Golang) using the Gin web framework. This system provides robust authentication, authorization, user management, and file upload capabilities.

## Features

### Authentication & Authorization
- JWT-based authentication
- User registration and login
- Password reset with OTP
- Token blacklisting for secure logout
- Role-based access control (User, Admin)
- Secure password hashing with bcrypt

### User Management
- Create, read, update, delete (CRUD) operations
- Soft delete functionality
- User roles and status management
- Pagination and filtering for user lists
- Search functionality across user fields

### File Upload
- Profile picture upload with validation
- File type and size restrictions
- Secure file storage and serving
- Profile picture deletion

### API Features
- RESTful API design
- Comprehensive Swagger documentation
- CORS support
- Request validation
- Structured error handling
- Standardized API responses

## Project Structure

```
generalusermanagement/
├── config/
│   └── config.go           # Configuration management
├── database/
│   └── database.go         # Database connection and setup
├── docs/
│   └── docs.go            # Swagger documentation
├── handlers/
│   ├── auth_handlers.go    # Authentication endpoints
│   ├── user_handlers.go    # User management endpoints
│   └── upload_handlers.go  # File upload endpoints
├── middleware/
│   └── middleware.go       # Authentication and CORS middleware
├── models/
│   └── models.go          # Data structures and types
├── routes/
│   └── routes.go          # Route definitions
├── utils/
│   └── utils.go           # Utility functions
├── uploads/               # File upload directory (auto-created)
├── .env                   # Environment variables
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── main.go               # Application entry point
└── README.md             # Project documentation
```

## Prerequisites

- Go 1.23.0 or later
- PostgreSQL database
- Git

## Installation & Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd generalusermanagement
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Database Setup
Create a PostgreSQL database and update the `.env` file with your database credentials.

### 4. Environment Configuration
Update the `.env` file with your specific configuration:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=generalusermanagement
DB_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key_here_make_it_long_and_random
JWT_EXPIRY_HOURS=24

# Server Configuration
SERVER_PORT=8082
SERVER_HOST=localhost

# File Upload Configuration
UPLOAD_PATH=./uploads
MAX_FILE_SIZE=5242880

# Environment
ENV=development
```

### 5. Generate Swagger Documentation (Optional)
```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swag init
```

### 6. Run the Application
```bash
go run main.go
```

The server will start on `http://localhost:8082`

## API Documentation

Once the server is running, you can access the Swagger documentation at:
- http://localhost:8082/swagger/index.html

## API Endpoints

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout (requires authentication)
- `GET /auth/check` - Check authentication status (requires authentication)
- `POST /auth/password-reset/request` - Request password reset OTP
- `POST /auth/password-reset/confirm` - Confirm password reset with OTP

### User Management
- `GET /users` - Get all users with pagination/filtering (admin only)
- `GET /users/{id}` - Get user by ID (user can access own data, admin can access all)
- `PUT /users/{id}` - Update user (user can update own data, admin can update all)
- `DELETE /users/{id}` - Delete user (admin only)
- `PUT /users/password` - Update password (requires current password)

### File Upload
- `POST /upload/profile-picture` - Upload profile picture (requires authentication)
- `DELETE /upload/profile-picture` - Delete profile picture (requires authentication)
- `GET /files/{filename}` - Serve uploaded files (public access)

### Utility
- `GET /health` - Health check endpoint

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    national_id VARCHAR(255),
    passport_number VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    profile_picture VARCHAR(500),
    username VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### Token Blacklist Table
```sql
CREATE TABLE token_blacklist (
    id SERIAL PRIMARY KEY,
    token VARCHAR(1000) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Password Resets Table
```sql
CREATE TABLE password_resets (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    otp VARCHAR(10) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## User Roles

- **user**: Standard user with limited access (can only access/modify own data)
- **admin**: Administrator with full access to all user management functions

## User Status

- **active**: User can log in and access the system
- **inactive**: User account exists but cannot log in
- **suspended**: User account is temporarily suspended

## Security Features

1. **Password Security**: Passwords are hashed using bcrypt with a cost of 14
2. **JWT Security**: Tokens include expiration time and can be blacklisted
3. **Input Validation**: All inputs are validated using Gin's binding features
4. **File Upload Security**: File type and size validation, secure file naming
5. **CORS Protection**: Configurable CORS middleware
6. **SQL Injection Protection**: All database queries use parameterized statements

## Configuration

The application uses environment variables for configuration. All settings can be found in the `.env` file:

- **Database Settings**: Connection parameters for PostgreSQL
- **JWT Settings**: Secret key and token expiration time
- **Server Settings**: Host and port configuration
- **File Upload Settings**: Upload directory and file size limits

## Error Handling

The API returns standardized error responses:

```json
{
    "success": false,
    "message": "Error description",
    "error": "Detailed error information"
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the Apache 2.0 License.

## Support

For support, please contact the API support team or create an issue in the repository.