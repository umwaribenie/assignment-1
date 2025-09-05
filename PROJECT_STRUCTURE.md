# User Management System - Complete Folder Structure

## 📁 Project Root Structure

```
user-management-system/
│
├── 📄 main.go                    # Application entry point
├── 📄 go.mod                     # Go module definition
├── 📄 go.sum                     # Go dependencies lock file
├── 📄 .env                       # Environment variables (create this)
├── 📄 .env.example               # Example environment variables
├── 📄 .gitignore                 # Git ignore file
├── 📄 README.md                  # Project documentation
├── 📄 Makefile                   # Build and development commands
├── 📄 docker-compose.yml         # Docker compose configuration
├── 📄 Dockerfile                 # Docker container definition
│
├── 📂 config/                    # Configuration management
│   ├── 📄 config.go             # Main configuration loader
│   ├── 📄 database.go           # Database configuration
│   ├── 📄 server.go             # Server configuration
│   └── 📄 jwt.go                # JWT configuration
│
├── 📂 database/                  # Database layer
│   ├── 📄 database.go           # Database connection and setup
│   ├── 📂 migrations/           # Database migrations
│   │   ├── 📄 001_create_users_table.up.sql
│   │   ├── 📄 001_create_users_table.down.sql
│   │   ├── 📄 002_create_sessions_table.up.sql
│   │   └── 📄 002_create_sessions_table.down.sql
│   └── 📂 seeds/                # Database seed data
│       └── 📄 users.sql
│
├── 📂 models/                    # Data models
│   ├── 📄 models.go             # All models in one file (current)
│   ├── 📄 user.go               # User model (if splitting)
│   ├── 📄 session.go            # Session model
│   ├── 📄 audit.go              # Audit log model
│   └── 📄 role.go               # Role and permission models
│
├── 📂 handlers/                  # HTTP handlers
│   ├── 📄 auth_handlers.go      # Authentication handlers
│   ├── 📄 user_handlers.go      # User CRUD handlers
│   ├── 📄 upload_handlers.go    # File upload handlers
│   ├── 📄 admin_handlers.go     # Admin-specific handlers
│   ├── 📄 profile_handlers.go   # User profile handlers
│   └── 📄 health_handlers.go    # Health check handlers
│
├── 📂 middleware/                # HTTP middleware
│   ├── 📄 middleware.go         # All middleware (current)
│   ├── 📄 auth.go               # Authentication middleware
│   ├── 📄 cors.go               # CORS middleware
│   ├── 📄 ratelimit.go          # Rate limiting middleware
│   ├── 📄 logging.go            # Request logging middleware
│   └── 📄 security.go           # Security headers middleware
│
├── 📂 routes/                    # Route definitions
│   ├── 📄 routes.go             # Main route setup
│   ├── 📄 auth_routes.go        # Authentication routes
│   ├── 📄 user_routes.go        # User management routes
│   ├── 📄 admin_routes.go       # Admin routes
│   └── 📄 api_routes.go         # API version grouping
│
├── 📂 services/                  # Business logic layer
│   ├── 📄 auth_service.go       # Authentication service
│   ├── 📄 user_service.go       # User management service
│   ├── 📄 email_service.go      # Email service
│   ├── 📄 upload_service.go     # File upload service
│   ├── 📄 audit_service.go      # Audit logging service
│   └── 📄 cache_service.go      # Caching service
│
├── 📂 repositories/              # Data access layer
│   ├── 📄 user_repository.go    # User data access
│   ├── 📄 session_repository.go # Session data access
│   └── 📄 audit_repository.go   # Audit log data access
│
├── 📂 utils/                     # Utility functions
│   ├── 📄 utils.go              # General utilities
│   ├── 📄 validators.go         # Input validation
│   ├── 📄 jwt.go                # JWT utilities
│   ├── 📄 password.go           # Password hashing
│   ├── 📄 response.go           # HTTP response helpers
│   └── 📄 errors.go             # Error handling utilities
│
├── 📂 dto/                       # Data Transfer Objects
│   ├── 📄 auth_dto.go           # Auth request/response DTOs
│   ├── 📄 user_dto.go           # User request/response DTOs
│   └── 📄 common_dto.go         # Common DTOs
│
├── 📂 validators/                # Request validation
│   ├── 📄 auth_validator.go     # Auth request validation
│   └── 📄 user_validator.go     # User request validation
│
├── 📂 uploads/                   # File uploads directory
│   ├── 📂 avatars/              # User avatar images
│   ├── 📂 documents/            # User documents
│   └── 📂 temp/                 # Temporary uploads
│
├── 📂 logs/                      # Application logs
│   ├── 📄 app.log               # Application log
│   ├── 📄 error.log             # Error log
│   └── 📄 access.log            # Access log
│
├── 📂 tests/                     # Test files
│   ├── 📂 unit/                 # Unit tests
│   │   ├── 📄 handlers_test.go
│   │   ├── 📄 services_test.go
│   │   └── 📄 utils_test.go
│   ├── 📂 integration/          # Integration tests
│   │   ├── 📄 auth_test.go
│   │   └── 📄 user_test.go
│   └── 📂 e2e/                  # End-to-end tests
│       └── 📄 api_test.go
│
├── 📂 scripts/                   # Utility scripts
│   ├── 📄 migrate.sh            # Database migration script
│   ├── 📄 seed.sh               # Database seeding script
│   └── 📄 deploy.sh             # Deployment script
│
├── 📂 docs/                      # Documentation
│   ├── 📄 docs.go               # Swagger documentation
│   ├── 📄 swagger.json          # Swagger JSON
│   ├── 📄 swagger.yaml          # Swagger YAML
│   ├── 📂 api/                  # API documentation
│   │   ├── 📄 authentication.md
│   │   └── 📄 users.md
│   └── 📂 guides/               # User guides
│       ├── 📄 installation.md
│       └── 📄 deployment.md
│
├── 📂 deployments/              # Deployment configurations
│   ├── 📂 kubernetes/           # Kubernetes manifests
│   │   ├── 📄 deployment.yaml
│   │   ├── 📄 service.yaml
│   │   └── 📄 configmap.yaml
│   └── 📂 docker/               # Docker configurations
│       └── 📄 docker-compose.prod.yml
│
└── 📂 .github/                  # GitHub specific files
    └── 📂 workflows/            # GitHub Actions
        ├── 📄 ci.yml            # Continuous Integration
        └── 📄 cd.yml            # Continuous Deployment
```

## 📋 Directory Descriptions

### Core Directories

#### `/config`
- **Purpose**: Centralized configuration management
- **Contents**: Environment variables, database config, server settings, JWT configuration

#### `/database`
- **Purpose**: Database connection, migrations, and seeds
- **Contents**: Database initialization, migration files, seed data

#### `/models`
- **Purpose**: Data structures and database models
- **Contents**: User, Session, Role, Audit models with GORM tags

#### `/handlers`
- **Purpose**: HTTP request handlers (Controllers)
- **Contents**: Functions that handle HTTP requests and responses

#### `/middleware`
- **Purpose**: HTTP middleware components
- **Contents**: Auth, CORS, rate limiting, logging, security middleware

#### `/routes`
- **Purpose**: API route definitions
- **Contents**: Route registration and grouping

### Business Logic Layer

#### `/services`
- **Purpose**: Business logic and orchestration
- **Contents**: Complex operations, external API calls, business rules

#### `/repositories`
- **Purpose**: Data access layer (DAL)
- **Contents**: Database queries and data persistence logic

### Supporting Directories

#### `/utils`
- **Purpose**: Shared utility functions
- **Contents**: Helpers, validators, JWT utilities, password hashing

#### `/dto`
- **Purpose**: Data Transfer Objects
- **Contents**: Request/response structures for API endpoints

#### `/validators`
- **Purpose**: Input validation logic
- **Contents**: Request validation rules and schemas

### Storage & Logs

#### `/uploads`
- **Purpose**: File storage
- **Contents**: User uploads, avatars, documents

#### `/logs`
- **Purpose**: Application logs
- **Contents**: Application, error, and access logs

### Development & Testing

#### `/tests`
- **Purpose**: Test suites
- **Contents**: Unit, integration, and e2e tests

#### `/scripts`
- **Purpose**: Development and deployment scripts
- **Contents**: Migration, seeding, deployment scripts

### Documentation

#### `/docs`
- **Purpose**: API and project documentation
- **Contents**: Swagger files, API guides, deployment guides

## 🚀 Implementation Steps

1. **Create Base Structure**
   ```bash
   mkdir -p config database/migrations database/seeds models handlers middleware routes
   mkdir -p services repositories utils dto validators
   mkdir -p uploads/avatars uploads/documents uploads/temp
   mkdir -p logs tests/unit tests/integration tests/e2e
   mkdir -p scripts docs/api docs/guides
   mkdir -p deployments/kubernetes deployments/docker
   mkdir -p .github/workflows
   ```

2. **Move Existing Files**
   - Keep current files in their locations
   - Gradually refactor and split large files

3. **Create New Files**
   - Split `models.go` into separate model files
   - Create service layer files
   - Add repository layer for data access
   - Create DTOs for API contracts

4. **Environment Setup**
   - Create `.env.example` with all required variables
   - Add `.gitignore` to exclude sensitive files
   - Create `Makefile` for common commands

5. **Docker Setup**
   - Create `Dockerfile` for containerization
   - Add `docker-compose.yml` for local development

## 🔧 Best Practices

1. **Separation of Concerns**
   - Handlers: HTTP layer only
   - Services: Business logic
   - Repositories: Data access
   - Models: Data structures

2. **File Naming**
   - Use snake_case for file names
   - Group related functionality
   - Keep files focused and small

3. **Dependencies**
   - Handlers → Services → Repositories → Models
   - Utils can be used anywhere
   - Avoid circular dependencies

4. **Testing**
   - Unit tests for services and utils
   - Integration tests for API endpoints
   - E2E tests for complete workflows

5. **Documentation**
   - Keep Swagger docs updated
   - Document complex business logic
   - Maintain README with setup instructions