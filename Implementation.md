\# URL Shortener Implementation Guide

\*Building with Claude Code on Windows 11\*



\## Table of Contents

1\. \[Prerequisites Setup](#prerequisites-setup)

2\. \[Claude Code Installation](#claude-code-installation)

3\. \[Project Structure](#project-structure)

4\. \[Implementation Steps](#implementation-steps)

5\. \[Local Testing Setup](#local-testing-setup)

6\. \[Domain Recommendations](#domain-recommendations)

7\. \[Deployment to Oracle Cloud](#deployment-to-oracle-cloud)

8\. \[Troubleshooting](#troubleshooting)



---



\## Prerequisites Setup



\### 1. Install Required Tools on Windows 11



\*\*Install Go:\*\*

```powershell

\# Using Chocolatey (recommended)

choco install golang



\# Or download from https://golang.org/dl/

\# Choose Windows installer, add to PATH

```



\*\*Install Git:\*\*

```powershell

choco install git

```



\*\*Install Docker Desktop:\*\*

```powershell

choco install docker-desktop

\# Or download from https://docs.docker.com/desktop/windows/install/

```



\*\*Install PostgreSQL (for local testing):\*\*

```powershell

choco install postgresql

\# Or use Docker (recommended for easier cleanup)

```



\*\*Install Redis (for local testing):\*\*

```powershell

choco install redis-64

\# Or use Docker (recommended)

```



\*\*Install VS Code (optional, for editing):\*\*

```powershell

choco install vscode

```



\### 2. Verify Installations



```powershell

go version          # Should show Go 1.21+

git --version       # Should show Git version

docker --version    # Should show Docker version

psql --version      # Should show PostgreSQL version

```



---



\## Claude Code Installation



\### 1. Install Claude Code CLI



```bash

\# Install via pip (requires Python)

pip install claude-code



\# Or download from https://docs.anthropic.com/en/docs/claude-code

```



\### 2. Setup Authentication



```bash

\# Set your Anthropic API key

claude auth login



\# Or set environment variable

set ANTHROPIC\_API\_KEY=your\_api\_key\_here

```



\### 3. Verify Claude Code Installation



```bash

claude --version

claude --help

```



---



\## Project Structure



\### Create Project Directory



```powershell

mkdir url-shortener-system

cd url-shortener-system

```



\### Recommended Project Structure



```

url-shortener-system/

├── cmd/

│   └── server/

│       └── main.go

├── internal/

│   ├── handlers/

│   │   ├── shorten.go

│   │   └── redirect.go

│   ├── services/

│   │   ├── shortener.go

│   │   └── analytics.go

│   ├── storage/

│   │   ├── postgres.go

│   │   └── redis.go

│   ├── models/

│   │   └── url.go

│   └── utils/

│       ├── base62.go

│       └── snowflake.go

├── configs/

│   ├── config.go

│   └── database.sql

├── docker/

│   ├── Dockerfile

│   ├── docker-compose.yml

│   └── docker-compose.local.yml

├── scripts/

│   ├── setup-local.ps1

│   └── deploy.sh

├── tests/

│   ├── integration/

│   └── unit/

├── docs/

│   └── api.md

├── .env.example

├── .gitignore

├── go.mod

├── go.sum

└── README.md

```



---



\## Implementation Steps



\### Step 1: Initialize Project with Claude Code



```bash

\# Create new Go project

claude create go-project --name url-shortener --template web-service



\# Or start from scratch

go mod init github.com/yourusername/url-shortener

```



\### Step 2: Use Claude Code for Core Development



Create a detailed prompt file: `implementation-prompt.md`



```markdown

\# URL Shortener Implementation Request



I need to build a URL shortener service in Go with these requirements:



\## Core Features:

1\. REST API to shorten URLs

2\. HTTP redirect service for short URLs  

3\. Base62 encoding with Snowflake ID generation

4\. PostgreSQL for persistence

5\. Redis for caching

6\. Basic analytics tracking

7\. Docker containerization



\## Technical Specifications:

\- Use Gin framework for HTTP routing

\- Implement distributed ID generation (Snowflake algorithm)

\- Base62 encoding for short codes

\- Database connection pooling

\- Redis caching layer

\- Structured logging

\- Health check endpoints

\- Environment-based configuration



\## API Endpoints:

\- POST /api/v1/shorten (create short URL)

\- GET /{shortCode} (redirect to original URL)

\- GET /api/v1/analytics/{shortCode} (basic stats)

\- GET /health (health check)



Please implement this following Go best practices with proper error handling, validation, and testing setup.

```



Run Claude Code:

```bash

claude code --file implementation-prompt.md --output ./

```



\### Step 3: Implement Core Components with Claude Code



\*\*For each major component, use Claude Code interactively:\*\*



```bash

\# Generate Base62 encoding utility

claude code "Implement Base62 encoding/decoding utility in Go with Snowflake ID generation for distributed systems. Include comprehensive tests."



\# Generate database layer

claude code "Create PostgreSQL storage layer for URL shortener with connection pooling, migrations, and CRUD operations for URL mappings."



\# Generate HTTP handlers

claude code "Implement Gin HTTP handlers for URL shortening API with proper validation, error handling, and JSON responses."



\# Generate caching layer

claude code "Create Redis caching layer for URL shortener with TTL, connection pooling, and fallback to database."

```



\### Step 4: Configuration and Environment Setup



Create `.env.example`:

```env

\# Server Configuration

SERVER\_PORT=8080

SERVER\_HOST=localhost

BASE\_URL=http://localhost:8080



\# Database Configuration

DB\_HOST=localhost

DB\_PORT=5432

DB\_USER=urlshortener

DB\_PASSWORD=your\_password

DB\_NAME=urlshortener\_db

DB\_MAX\_CONNECTIONS=25



\# Redis Configuration

REDIS\_HOST=localhost

REDIS\_PORT=6379

REDIS\_PASSWORD=

REDIS\_DB=0



\# Snowflake Configuration

NODE\_ID=1



\# Logging

LOG\_LEVEL=info

LOG\_FORMAT=json

```



\### Step 5: Database Setup



Create `configs/database.sql`:

```sql

-- Create database

CREATE DATABASE urlshortener\_db;



-- Create user

CREATE USER urlshortener WITH PASSWORD 'your\_password';

GRANT ALL PRIVILEGES ON DATABASE urlshortener\_db TO urlshortener;



-- Connect to database

\\c urlshortener\_db;



-- Create tables

CREATE TABLE url\_mappings (

&nbsp;   id BIGINT PRIMARY KEY,

&nbsp;   short\_code VARCHAR(10) UNIQUE NOT NULL,

&nbsp;   original\_url TEXT NOT NULL,

&nbsp;   created\_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

&nbsp;   expires\_at TIMESTAMP WITH TIME ZONE,

&nbsp;   click\_count BIGINT DEFAULT 0,

&nbsp;   is\_active BOOLEAN DEFAULT TRUE,

&nbsp;   created\_by\_ip INET,

&nbsp;   

&nbsp;   INDEX idx\_short\_code (short\_code),

&nbsp;   INDEX idx\_created\_at (created\_at),

&nbsp;   INDEX idx\_active (is\_active)

);



CREATE TABLE click\_events (

&nbsp;   id BIGINT PRIMARY KEY,

&nbsp;   short\_code VARCHAR(10) NOT NULL,

&nbsp;   clicked\_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

&nbsp;   ip\_address INET,

&nbsp;   user\_agent TEXT,

&nbsp;   referrer TEXT,

&nbsp;   country\_code CHAR(2),

&nbsp;   

&nbsp;   INDEX idx\_short\_code\_time (short\_code, clicked\_at),

&nbsp;   FOREIGN KEY (short\_code) REFERENCES url\_mappings(short\_code)

);

```



---



\## Local Testing Setup



\### 1. Create Docker Compose for Local Development



Create `docker-compose.local.yml`:

```yaml

version: '3.8'



services:

&nbsp; postgres:

&nbsp;   image: postgres:15-alpine

&nbsp;   environment:

&nbsp;     POSTGRES\_DB: urlshortener\_db

&nbsp;     POSTGRES\_USER: urlshortener

&nbsp;     POSTGRES\_PASSWORD: your\_password

&nbsp;   ports:

&nbsp;     - "5432:5432"

&nbsp;   volumes:

&nbsp;     - postgres\_data:/var/lib/postgresql/data

&nbsp;     - ./configs/database.sql:/docker-entrypoint-initdb.d/init.sql

&nbsp;   healthcheck:

&nbsp;     test: \["CMD-SHELL", "pg\_isready -U urlshortener"]

&nbsp;     interval: 10s

&nbsp;     timeout: 5s

&nbsp;     retries: 5



&nbsp; redis:

&nbsp;   image: redis:7-alpine

&nbsp;   ports:

&nbsp;     - "6379:6379"

&nbsp;   volumes:

&nbsp;     - redis\_data:/data

&nbsp;   healthcheck:

&nbsp;     test: \["CMD", "redis-cli", "ping"]

&nbsp;     interval: 10s

&nbsp;     timeout: 3s

&nbsp;     retries: 5



&nbsp; adminer:

&nbsp;   image: adminer

&nbsp;   ports:

&nbsp;     - "8081:8080"

&nbsp;   depends\_on:

&nbsp;     - postgres



volumes:

&nbsp; postgres\_data:

&nbsp; redis\_data:

```



\### 2. Create Local Setup Script



Create `scripts/setup-local.ps1`:

```powershell

\# PowerShell script for Windows 11 setup

Write-Host "Setting up URL Shortener local development environment..."



\# Copy environment file

if (!(Test-Path ".env")) {

&nbsp;   Copy-Item ".env.example" ".env"

&nbsp;   Write-Host "Created .env file - please update with your settings"

}



\# Start Docker services

Write-Host "Starting Docker services..."

docker-compose -f docker-compose.local.yml up -d



\# Wait for services to be ready

Write-Host "Waiting for services to be ready..."

Start-Sleep -Seconds 10



\# Check service health

$postgresHealth = docker-compose -f docker-compose.local.yml ps postgres --format json | ConvertFrom-Json | Select-Object -ExpandProperty Health

$redisHealth = docker-compose -f docker-compose.local.yml ps redis --format json | ConvertFrom-Json | Select-Object -ExpandProperty Health



if ($postgresHealth -eq "healthy" -and $redisHealth -eq "healthy") {

&nbsp;   Write-Host "✅ All services are healthy!"

} else {

&nbsp;   Write-Host "⚠️ Some services may not be ready. Check with: docker-compose -f docker-compose.local.yml ps"

}



\# Install Go dependencies

Write-Host "Installing Go dependencies..."

go mod tidy



\# Run database migrations (if you have them)

\# go run cmd/migrate/main.go



Write-Host "✅ Local setup complete!"

Write-Host "🔗 Database admin: http://localhost:8081 (adminer)"

Write-Host "📊 Server will run on: http://localhost:8080"

Write-Host "🚀 To start the server: go run cmd/server/main.go"

```



\### 3. Testing Commands



```powershell

\# Start local environment

.\\scripts\\setup-local.ps1



\# Run the application

go run cmd/server/main.go



\# Test API endpoints

\# Shorten URL

curl -X POST http://localhost:8080/api/v1/shorten -H "Content-Type: application/json" -d "{\\"url\\":\\"https://www.google.com\\"}"



\# Test redirect

curl -I http://localhost:8080/ABC123



\# Check analytics

curl http://localhost:8080/api/v1/analytics/ABC123



\# Health check

curl http://localhost:8080/health

```



\### 4. Create Test Suite



Use Claude Code to generate comprehensive tests:

```bash

claude code "Generate comprehensive Go test suite for URL shortener including unit tests, integration tests, and API tests using testify framework."

```



---



\## Domain Recommendations



\### Premium Short Domain Options



\*\*Tier 1 - Premium but Available:\*\*

\- `ly.to` - Libya extension, good for "link to"

\- `is.gd` - Iceland, established pattern

\- `v.gd` - Virgin Islands, very short

\- `owl.li` - Personal branding with "links"

\- `s.id` - Indonesia, perfect for "short id"



\*\*Tier 2 - Creative Options:\*\*

\- `url8.co` - Play on "URL ate"

\- `snip.ly` - Clear purpose indication  

\- `link4.me` - Personal touch

\- `qr.lu` - Luxembourg, good for QR integration

\- `x.co` - Very minimal, modern



\*\*Tier 3 - Brandable Options:\*\*

\- `zipurl.co` - Fast/compressed feeling

\- `tinypath.com` - Descriptive and memorable

\- `quicklink.io` - Tech-focused branding

\- `snapurl.net` - Quick action implication

\- `brieflink.com` - Professional sounding



\### Domain Selection Criteria



\*\*Technical Requirements:\*\*

\- \*\*Length\*\*: 4-8 characters for optimal mobile typing

\- \*\*TLD\*\*: .ly, .co, .io, .me work well for link shorteners

\- \*\*Memorability\*\*: Easy to type and remember

\- \*\*Brandability\*\*: Can build marketing around it



\*\*Cost Considerations:\*\*

\- \*\*.ly domains\*\*: $75-100/year (premium but established)

\- \*\*.co domains\*\*: $25-35/year (good balance)

\- \*\*.com domains\*\*: $10-15/year (cheapest, but longer names)



\*\*Availability Check Commands:\*\*

```powershell

\# Check domain availability (example tools)

nslookup ly.to

nslookup s.id

nslookup zipurl.co



\# Or use online tools:

\# - Namecheap.com

\# - GoDaddy.com  

\# - Google Domains

```



\### My Top Recommendations:



1\. \*\*`s.id`\*\* - Perfect meaning, short, .id extension growing in popularity

2\. \*\*`zipurl.co`\*\* - Brandable, available, good TLD

3\. \*\*`ly.to`\*\* - Clear purpose, established pattern (if available)



---



\## Deployment to Oracle Cloud



\### 1. Prepare for Production



Create `docker/Dockerfile`:

```dockerfile

\# Multi-stage build for ARM64

FROM golang:1.21-alpine AS builder



WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download



COPY . .

RUN GOOS=linux GOARCH=arm64 CGO\_ENABLED=0 go build -o url-shortener cmd/server/main.go



FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/



COPY --from=builder /app/url-shortener .

COPY --from=builder /app/configs ./configs



EXPOSE 8080

CMD \["./url-shortener"]

```



\### 2. Production Docker Compose



Create `docker-compose.yml`:

```yaml

version: '3.8'



services:

&nbsp; app:

&nbsp;   build: .

&nbsp;   ports:

&nbsp;     - "8080:8080"

&nbsp;   environment:

&nbsp;     - SERVER\_PORT=8080

&nbsp;     - DB\_HOST=postgres

&nbsp;     - REDIS\_HOST=redis

&nbsp;   depends\_on:

&nbsp;     - postgres

&nbsp;     - redis

&nbsp;   restart: unless-stopped



&nbsp; postgres:

&nbsp;   image: postgres:15-alpine

&nbsp;   environment:

&nbsp;     POSTGRES\_DB: ${DB\_NAME}

&nbsp;     POSTGRES\_USER: ${DB\_USER}

&nbsp;     POSTGRES\_PASSWORD: ${DB\_PASSWORD}

&nbsp;   volumes:

&nbsp;     - postgres\_data:/var/lib/postgresql/data

&nbsp;   restart: unless-stopped



&nbsp; redis:

&nbsp;   image: redis:7-alpine

&nbsp;   volumes:

&nbsp;     - redis\_data:/data

&nbsp;   restart: unless-stopped



&nbsp; nginx:

&nbsp;   image: nginx:alpine

&nbsp;   ports:

&nbsp;     - "80:80"

&nbsp;     - "443:443"

&nbsp;   volumes:

&nbsp;     - ./nginx.conf:/etc/nginx/nginx.conf

&nbsp;     - ./ssl:/etc/nginx/ssl

&nbsp;   depends\_on:

&nbsp;     - app

&nbsp;   restart: unless-stopped



volumes:

&nbsp; postgres\_data:

&nbsp; redis\_data:

```



\### 3. Deploy Script



Create `scripts/deploy.sh`:

```bash

\#!/bin/bash

set -e



echo "🚀 Deploying URL Shortener to Oracle Cloud..."



\# Build for ARM64

docker buildx build --platform linux/arm64 -t url-shortener:latest .



\# Copy to server

scp -r . oracle-cloud-user@your-server-ip:~/url-shortener/



\# SSH and deploy

ssh oracle-cloud-user@your-server-ip << 'EOF'

cd ~/url-shortener

docker-compose down

docker-compose up -d --build

docker-compose logs -f

EOF



echo "✅ Deployment complete!"

```



---



\## Troubleshooting



\### Common Windows 11 Issues



\*\*Docker Desktop Issues:\*\*

```powershell

\# Enable WSL2

wsl --install

wsl --set-default-version 2



\# Restart Docker Desktop

\# Enable "Use WSL 2 based engine" in Docker settings

```



\*\*Go PATH Issues:\*\*

```powershell

\# Check GOPATH

go env GOPATH



\# Add to PATH if missing

$env:PATH += ";C:\\Users\\%USERNAME%\\go\\bin"

```



\*\*PostgreSQL Connection Issues:\*\*

```powershell

\# Check if PostgreSQL is running

Get-Service -Name postgresql\*



\# Test connection

psql -h localhost -U urlshortener -d urlshortener\_db

```



\### Application Issues



\*\*Port Already in Use:\*\*

```powershell

\# Find process using port 8080

netstat -ano | findstr :8080



\# Kill process (replace PID)

taskkill /PID <PID> /F

```



\*\*Database Migration Issues:\*\*

```bash

\# Reset database

docker-compose -f docker-compose.local.yml down -v

docker-compose -f docker-compose.local.yml up -d

```



\*\*Redis Connection Issues:\*\*

```bash

\# Test Redis connection

redis-cli ping

\# Should return PONG

```



\### Performance Testing



\*\*Load Testing with curl:\*\*

```bash

\# Simple load test

for i in {1..100}; do

&nbsp; curl -X POST http://localhost:8080/api/v1/shorten \\

&nbsp;   -H "Content-Type: application/json" \\

&nbsp;   -d "{\\"url\\":\\"https://example.com/$i\\"}" \&

done

wait

```



\*\*Using Apache Bench (if installed):\*\*

```bash

\# Install Apache Bench

choco install apache-httpd



\# Test redirect performance

ab -n 1000 -c 10 http://localhost:8080/ABC123

```



---



\## Next Steps



1\. \*\*Complete Local Development:\*\*

&nbsp;  - Run `.\\scripts\\setup-local.ps1`

&nbsp;  - Use Claude Code for detailed implementation

&nbsp;  - Test all endpoints thoroughly



2\. \*\*Domain Purchase:\*\*

&nbsp;  - Choose from recommended domains

&nbsp;  - Configure DNS settings

&nbsp;  - Setup SSL certificates



3\. \*\*Production Deployment:\*\*

&nbsp;  - Setup Oracle Cloud instance  

&nbsp;  - Configure domain and SSL

&nbsp;  - Deploy using Docker Compose

&nbsp;  - Setup monitoring and backups



4\. \*\*Feature Enhancement:\*\*

&nbsp;  - Add user authentication

&nbsp;  - Implement advanced analytics

&nbsp;  - Create web dashboard

&nbsp;  - Add API rate limiting



This implementation guide provides a complete path from local development to production deployment on Oracle Cloud Free Tier using Claude Code for development assistance.

