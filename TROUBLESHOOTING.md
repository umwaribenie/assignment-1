# General User Management - Troubleshooting Guide

## Common Issues and Solutions

### 1. Database Connection Issues (Most Common)

**Symptoms:** 
- Red lines in IDE
- Error: `dial tcp [::1]:5432: connect: connection refused`
- Application fails to start

**Solutions:**

#### Option A: Install and Setup PostgreSQL Locally

1. **Install PostgreSQL:**
   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   
   # macOS
   brew install postgresql
   brew services start postgresql
   
   # Windows
   # Download from https://www.postgresql.org/download/windows/
   ```

2. **Create Database:**
   ```bash
   sudo -u postgres psql
   ```
   
   Then run:
   ```sql
   CREATE DATABASE generalusermanagement;
   CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
   \q
   ```

3. **Set Password (if needed):**
   ```bash
   sudo -u postgres psql
   ALTER USER postgres PASSWORD 'password';
   \q
   ```

#### Option B: Use Docker (Recommended)

1. **Create docker-compose.yml:**
   ```yaml
   version: '3.8'
   services:
     postgres:
       image: postgres:15
       container_name: general_user_management_db
       environment:
         POSTGRES_DB: generalusermanagement
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: password
       ports:
         - "5432:5432"
       volumes:
         - postgres_data:/var/lib/postgresql/data
   
   volumes:
     postgres_data:
   ```

2. **Start Database:**
   ```bash
   docker-compose up -d
   ```

#### Option C: Use SQLite (Quick Solution)

1. **Install SQLite driver:**
   ```bash
   go get github.com/mattn/go-sqlite3
   ```

2. **Modify database/database.go** to use SQLite instead of PostgreSQL.

### 2. IDE Configuration Issues

**Symptoms:**
- Red underlines in editor
- Import errors
- Function not found errors

**Solutions:**

1. **Ensure Go is properly installed:**
   ```bash
   go version
   ```

2. **Verify GOPATH and GOROOT:**
   ```bash
   go env GOPATH
   go env GOROOT
   ```

3. **Clean module cache:**
   ```bash
   go clean -modcache
   go mod download
   ```

4. **Restart Language Server:**
   - In VS Code: `Ctrl+Shift+P` → "Go: Restart Language Server"
   - In GoLand: File → Invalidate Caches and Restart

### 3. Missing Dependencies

**Solution:**
```bash
go mod tidy
go mod download
```

### 4. Import Path Issues

**Symptoms:**
- Package not found errors
- Module resolution issues

**Solution:**
Ensure your `go.mod` file has the correct module name:
```go
module generalusermanagement

go 1.24.2
```

### 5. Environment Variables

**Create or Update .env file:**
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=generalusermanagement
DB_SSL_MODE=disable

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key_change_in_production
JWT_EXPIRY_HOURS=24

# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8082

# File Upload Configuration
UPLOAD_PATH=./uploads
MAX_FILE_SIZE=5242880

# Environment
ENV=development
```

## Quick Start Commands

1. **Setup Database (Docker):**
   ```bash
   docker run --name postgres-db -e POSTGRES_DB=generalusermanagement -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:15
   ```

2. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

3. **Format Code:**
   ```bash
   gofmt -w .
   ```

4. **Run Application:**
   ```bash
   go run .
   ```

5. **Test Build:**
   ```bash
   go build -v
   ```

## Common Error Messages and Solutions

| Error | Solution |
|-------|----------|
| `dial tcp [::1]:5432: connect: connection refused` | Start PostgreSQL service or use Docker |
| `package generalusermanagement/xxx is not in GOROOT` | Run `go mod tidy` |
| `undefined: function_name` | Check import statements and package structure |
| `cannot find module providing package` | Add missing dependency with `go get` |

## IDE-Specific Solutions

### VS Code
1. Install Go extension
2. Set `"go.useLanguageServer": true` in settings
3. Run `Go: Install/Update Tools` command

### GoLand/IntelliJ
1. Ensure Go SDK is configured
2. Check Project SDK settings
3. Rebuild project indexes

### Vim/Neovim
1. Install vim-go plugin
2. Run `:GoInstallBinaries`
3. Set up LSP with gopls

## Testing Database Connection

Create a simple test file `test_db.go`:
```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=password dbname=generalusermanagement sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    err = db.Ping()
    if err != nil {
        log.Fatal("Cannot connect to database:", err)
    }
    
    fmt.Println("✅ Database connection successful!")
}
```

Run with: `go run test_db.go`