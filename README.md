# 🌐 Post Ur World - High-Performance Social Platform Backend

[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16.3-4169E1?style=for-the-badge&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-6.2-DC382D?style=for-the-badge&logo=redis)](https://redis.io/)
[![Swagger](https://img.shields.io/badge/Swagger-OpenAPI--2.0-85EA2D?style=for-the-badge&logo=swagger)](http://localhost:8080/v1/swagger/index.html)
[![React Frontend](https://img.shields.io/badge/Frontend-React%20%2B%20Vite-61DAFB?style=for-the-badge&logo=react)](./web)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue?style=for-the-badge)](./LICENSE)

---

## 📌 Table of Contents

- [📖 Overview](#-overview)
- [🏗️ Architectural Philosophy](#️-architectural-philosophy)
- [✨ Key Features](#-key-features)
- [📁 Directory & Folder Structure](#-directory--folder-structure)
  - [📂 Deep Dive into Folder Responsibilities](#-deep-dive-into-folder-responsibilities)
- [🛠️ Tech Stack & Requirements](#️-tech-stack--requirements)
- [🚀 Guided Step-by-Step Setup](#-guided-step-by-step-setup)
  - [Step 1: Repository Setup](#step-1-repository-setup)
  - [Step 2: Environment File Setup](#step-2-environment-file-setup)
  - [Step 3: Database & Caching Services (Docker Compose)](#step-3-database--caching-services-docker-compose)
  - [Step 4: Executing Database Migrations](#step-4-executing-database-migrations)
  - [Step 5: Database Seeding & Resetting](#step-5-database-seeding--resetting)
  - [Step 6: Generating Swagger Documentation](#step-6-generating-swagger-documentation)
  - [Step 7: Launching the Backend Server](#step-7-launching-the-backend-server)
  - [Step 8: Launching the React Web Frontend](#step-8-launching-the-react-web-frontend)
- [📜 Makefile Command Reference](#-makefile-command-reference)
- [🔌 API Endpoints Summary](#-api-endpoints-summary)
- [🧪 Testing & Benchmarking](#-testing--benchmarking)
  - [Unit & Integration Tests](#unit--integration-tests)
  - [Concurrency & Optimistic Locking Testing](#concurrency--optimistic-locking-testing)
  - [Rate Limiter & Load Testing](#rate-limiter--load-testing)
- [🚢 CI/CD & Automated Release Pipeline](#-cicd--automated-release-pipeline)
- [🌐 Deployment Guide](#-deployment-guide)

---

## 📖 Overview

**Post Ur World** is a production-grade, highly scalable backend API built in **Go (Golang)** designed for modern social networking platforms. It provides a complete ecosystem for software engineers and users to register accounts, activate profiles via email verification, create posts with tags, comment on posts, build social connections (following/unfollowing), and consume customized activity feeds.

Engineered with enterprise-grade standards, the platform features **Clean Layered Architecture**, **JWT Authentication**, **Role-Based Access Control (RBAC)**, **Optimistic Concurrency Control**, **Redis Profile Caching**, **IP-based Rate Limiting**, **Uber Zap Logging**, **Standard Metrics Monitoring**, and **Automated CI/CD Workflows**.

---

## 🏗️ Architectural Philosophy

The backend is built strictly adhering to clean software engineering principles:

```
                  +-------------------------------------------------------+
                  |                 HTTP Transport Layer                  |
                  |     (Chi Router, Middlewares, Handlers, DTOs)        |
                  +-------------------------------------------------------+
                                              |
                                              v
                  +-------------------------------------------------------+
                  |               Core Application Layer                  |
                  |     (JWT Auth, Rate Limiter, SendGrid Mailer)         |
                  +-------------------------------------------------------+
                        |                                   |
                        v                                   v
  +-----------------------------------+               +-----------------------------------+
  |       Storage Layer (Postgres)    |               |      Cache Storage (Redis)        |
  |  (Users, Posts, Comments, Roles)  |               |  (Cached User Profiles & TTL)     |
  +-----------------------------------+               +-----------------------------------+
```

1. **Separation of Concerns (SoC):** Distinct boundaries separate transport handling (`cmd/api`), domain abstractions/services (`internal/store`, `internal/auth`, `internal/mailer`), and infrastructure logic (`internal/db`, `internal/store/cache`).
2. **Dependency Inversion Principle (DIP):** Top-level domain handlers depend on abstract interfaces (`store.Store`, `cache.Storage`, `auth.Authenticator`, `mailer.Client`) rather than concrete implementations. This enables seamless mock testing and component substitution.
3. **Optimistic Locking:** Post updates utilize a strict `version` column mechanism to guarantee safe concurrent modifications without row-locking performance penalties.
4. **Resilient Infrastructure:** Implements connection pooling for PostgreSQL, fallback mechanisms for Redis caching, and context cancellation with graceful server shutdown.

---

## ✨ Key Features

- 🔐 **Authentication & Authorization:**
  - JWT Bearer Authentication with 3-day validity, configurable issuer, and target audience validation.
  - User registration flow integrated with SendGrid email activation links using SHA-256 token hashing.
  - HTTP Basic Auth securing infrastructure operations (`/v1/health`, `/v1/metrics`).
  - Role-Based Access Control (RBAC) with granular permissions (`user`, `moderator`, `admin`).
- 📝 **Post & Content Management:**
  - Create, read, update, delete (CRUD) posts with multi-tag support.
  - Optimistic concurrency control via incremental `version` checking.
  - Commenting system linked to posts with relational integrity constraints.
- 🤝 **Social Network Graph & Activity Feed:**
  - User follow/unfollow capabilities with user connection mapping.
  - Fan-out activity feed aggregate endpoint supporting pagination (limit/offset), tag filters, search queries, and sorting.
- ⚡ **Performance & Caching:**
  - High-performance Redis caching layer for fast user profile retrieval.
  - Feature toggle flag (`REDIS_ENABLED`) to gracefully bypass cache when Redis is disabled.
- 🛡️ **Rate Limiting & Safety:**
  - Custom, thread-safe Fixed-Window Rate Limiting middleware tracking IP traffic.
- 📊 **Observability & Diagnostics:**
  - High-speed structured logging with Uber `zap`.
  - Exposes runtime indicators (Goroutines count, Database pool stats, Version) via standard Go `expvar`.
- 📚 **Interactive Swagger API Documentation:**
  - Fully annotated OpenAPI 2.0 specifications accessible via integrated Swagger UI (`/v1/swagger/index.html`).

---

## 📁 Directory & Folder Structure

```
backend_go/
├── .github/
│   └── workflows/              # GitHub Actions CI/CD workflows
│       ├── audit.yaml          # Static analysis, linting, and go test execution
│       ├── release-please.yaml # Automated semantic versioning and release creation
│       └── update-api-version.yaml # Automatic version sync script workflow
├── cmd/                        # Application entry points & command-line utilities
│   ├── api/                    # Primary HTTP API server application
│   │   ├── api.go              # Server initialization, Chi router setup & routes definition
│   │   ├── auth.go             # Authentication handlers (registration, token issuance)
│   │   ├── errors.go           # Standardized HTTP error responses and helper functions
│   │   ├── feeds.go            # User activity feed handlers
│   │   ├── health.go           # Health check and metrics endpoint handlers
│   │   ├── json.go             # JSON payload encoding/decoding helper utilities
│   │   ├── main.go             # Main application entry point & config parsing
│   │   ├── middlewares.go      # Custom HTTP middlewares (Auth, RBAC, Rate Limiter, Context)
│   │   ├── posts.go            # Post CRUD and ownership management handlers
│   │   ├── test_utils.go       # Helper utilities for API integration testing
│   │   ├── user_test.go        # Unit and integration tests for user endpoints
│   │   └── users.go            # User management handlers (activation, follow, unfollow)
│   └── migrate/                # Database management commands
│       ├── migrations/         # Up & Down SQL migration scripts (000001 - 000013)
│       └── seed/               # Database seeding binary entry point
│           └── main.go         # Executable entry point to trigger database seeders
├── docs/                       # Auto-generated Swagger documentation files
│   ├── docs.go                 # Go source representation of Swagger spec
│   ├── swagger.json            # JSON OpenAPI specification
│   └── swagger.yaml            # YAML OpenAPI specification
├── internal/                   # Private core application code (business logic)
│   ├── auth/                   # Authentication logic & providers
│   │   ├── auth.go             # Authenticator interface definitions
│   │   ├── jwt.go              # JWT token generation, parsing, and claim verification
│   │   └── mocks.go            # Mock authenticator for unit tests
│   ├── db/                     # Database connection pool setup & seed generators
│   │   ├── db.go               # PostgreSQL connection pool builder
│   │   └── seed.go             # Dummy data generators (faker/seeders for users, posts, comments)
│   ├── mailer/                 # Mail engine adapters
│   │   ├── mailer.go           # Mailer interface definition
│   │   ├── sendGrid.go         # SendGrid API v3 integration client
│   │   └── templates/          # HTML & text email templates
│   │       └── user_invitation.tmpl # Invitation email HTML template
│   ├── ratelimiter/            # Traffic control rate limiting engine
│   │   ├── fixed_window.go     # Thread-safe in-memory fixed window rate limiter implementation
│   │   └── ratelimiter.go      # Rate limiter interface definitions
│   └── store/                  # Data access layer (PostgreSQL abstraction)
│       ├── cache/              # Caching abstraction layer
│       │   ├── mocks.go        # Mock cache storage for testing
│       │   ├── redis.go        # Redis client factory connection setup
│       │   ├── storage.go      # Cache storage interface definition
│       │   └── users.go        # User caching logic (GET/SET/DELETE user profiles in Redis)
│       ├── comments.go         # Comments repository methods (SQL queries)
│       ├── followers.go        # Followers graph repository methods (SQL queries)
│       ├── mocks.go            # Mock database store for unit testing
│       ├── pagination.go       # Pagination helper structs & SQL query parameters
│       ├── posts.go            # Posts repository methods & feed generation SQL logic
│       ├── roles.go            # Roles repository methods & RBAC SQL queries
│       ├── storage.go          # Central Store struct holding repository interfaces
│       └── users.go            # Users repository methods (Create, Activate, GetByEmail)
├── scripts/                    # Helper utility scripts
│   ├── db_init.sql             # SQL script for database initialization
│   └── test_concurrency.go     # Concurrent HTTP requester to test optimistic locking
├── web/                        # React + Vite + TypeScript web frontend dashboard
│   ├── src/                    # Frontend React source code
│   ├── package.json            # Node.js project dependencies & scripts
│   └── vite.config.ts          # Vite build tool configuration
├── .air.toml                   # Live reloading config for Air dev server
├── .env                        # Local environment variables configuration file
├── .gitignore                  # Git untracked file exclusions
├── CHANGELOG.md                # Conventional changelog history file
├── Dockerfile                  # Production multi-stage Docker build file for backend
├── Makefile                    # Task runner automation rules
├── docker-compose.yml          # Docker Compose orchestration file (PostgreSQL, Redis, Redis Commander)
├── details.txt                 # Project architecture, commands, and setup reference notes
├── go.mod                      # Go module definitions and dependency lockfile
├── go.sum                      # Go module checksum verifications
└── README.md                   # Project documentation (this document)
```

---

### 📂 Deep Dive into Folder Responsibilities

| Directory / File | Core Responsibility |
| :--- | :--- |
| **`cmd/api`** | Transport Layer containing Chi HTTP routing, controller logic, middleware binding, JSON serialization, and error handling. |
| **`cmd/migrate`** | SQL migration scripts management. Contains versioned Up/Down migration pairs and database seed binary. |
| **`internal/store`** | Database abstraction layer executing raw SQL queries against PostgreSQL with context timeouts and model mapping. |
| **`internal/store/cache`**| Cache storage layer managing Redis key-value serialization, TTL expirations, and cache invalidation. |
| **`internal/auth`** | Cryptographic security layer handling JWT token signing, verification, claims parsing, and password hashing. |
| **`internal/mailer`** | External email communication module using SendGrid API to deliver responsive HTML registration templates. |
| **`internal/ratelimiter`**| In-memory concurrency-safe rate limiter tracking traffic thresholds per remote IP address. |
| **`docs`** | Contains Swagger 2.0 OpenAPI documentation generated automatically via `swag`. |
| **`scripts`** | Verification scripts such as `test_concurrency.go` which simulates parallel write requests to test optimistic locking. |
| **`web`** | Modern single-page web application frontend built using React 19, TypeScript, and Vite. |

---

## 🛠️ Tech Stack & Requirements

### Tech Stack
- **Language:** Go `1.26.3`
- **HTTP Framework & Router:** `chi/v5`
- **Database:** PostgreSQL `16.3`
- **Caching Engine:** Redis `6.2-alpine`
- **ORMs / SQL Execution:** Standard Go `database/sql` + `lib/pq` PostgreSQL driver
- **Authentication:** `golang-jwt/jwt/v5` & `golang.org/x/crypto/bcrypt`
- **Documentation:** `swaggo/swag` & `swaggo/http-swagger`
- **Logging:** Uber `zap` (production structured logger)
- **Email Engine:** SendGrid API (`sendgrid-go`)
- **Testing:** `stretchr/testify`
- **Dev Hot Reload:** Air (`air-verse/air`)
- **Containerization:** Docker & Docker Compose
- **Frontend Stack:** React 19, TypeScript, Vite

### System Prerequisites
Ensure the following tools are installed on your machine before setup:
1. **Go:** [Install Go 1.22+](https://golang.org/doc/install)
2. **Docker & Docker Compose:** [Install Docker Desktop / Engine](https://docs.docker.com/get-docker/)
3. **Golang Migrate CLI:** [Install migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
4. **Swag CLI (Swagger Generator):** `go install github.com/swaggo/swag/cmd/swag@latest`
5. **Air (Hot Reload Tool - Optional):** `go install github.com/air-verse/air@latest`
6. **Node.js (for Web Frontend):** [Install Node.js 18+](https://nodejs.org/)

---

## 🚀 Guided Step-by-Step Setup

Follow these exact steps to set up and run the entire application locally from scratch:

### Step 1: Repository Setup
Clone the project repository to your workspace:
```bash
git clone https://github.com/nagakushal786/post-ur-world.git
cd backend_go
```

Download and verify all Go module dependencies:
```bash
go mod download
go mod verify
```

---

### Step 2: Environment File Setup
Verify that your `.env` file is present in the workspace root:
```bash
ls -la .env
```
*(Ensure all values, especially database credentials and secret keys, match your target local setup).*

---

### Step 3: Database & Caching Services (Docker Compose)

Start the PostgreSQL 16 database, Redis server, and Redis Commander UI in detached mode:
```bash
make docker-up
```
*Alternatively using docker compose directly:*
```bash
docker compose up --build -d
```

Verify that the docker containers are active and healthy:
```bash
docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Ports}}"
```
You should see:
- `postgres-db` listening on `0.0.0.0:5432->5432/tcp`
- `puw-redis` listening on `0.0.0.0:6379->6379/tcp`
- `puw-redis-commander` listening on `127.0.0.1:8081->8081/tcp`

> 💡 **Note for local PostgreSQL service conflicts:** If PostgreSQL is already running locally on port 5432 on your Linux host, stop it using:
> ```bash
> sudo systemctl stop postgresql
> ```

---

### Step 4: Executing Database Migrations

Apply all 13 database migrations (schema tables, foreign keys, indexes, roles, permissions):
```bash
make migrate-up
```

If you need to roll back migrations:
```bash
# Rollback 1 migration step
make migrate-down 1
```

To create a new migration pair in the future:
```bash
make migrate-create create_example_table
```

---

### Step 5: Database Seeding & Resetting

To populate your database with dummy users, posts, comments, and role allocations:
```bash
make seed
```

If you want to clear existing data before re-seeding:
```bash
make truncate
make seed
```

---

### Step 6: Generating Swagger Documentation

Regenerate the OpenAPI / Swagger documentation files:
```bash
make gen-docs
```
This updates the static documentation inside the `./docs` directory.

---

### Step 7: Launching the Backend Server

#### Option A: With Live Hot-Reloading (Recommended for Development)
```bash
make watch
```
*(Or simply execute `air` in the root directory)*

#### Option B: Direct Go Run
```bash
go run ./cmd/api
```

Once launched, your server will display:
```
{"level":"info","ts":...,"caller":"api/main.go:198","msg":"Database connection pool established"}
{"level":"info","ts":...,"caller":"api/main.go:204","msg":"Redis connection pool established"}
{"level":"info","ts":...,"caller":"api/api.go:192","msg":"Server has started","addr":":8080","env":"development"}
```

You can now open your browser and verify:
- **Health Check Endpoint:** `http://localhost:8080/v1/health` (Requires Basic Auth: `admin` / `admin`)
- **Swagger Documentation:** `http://localhost:8080/v1/swagger/index.html`
- **Redis Commander UI:** `http://localhost:8081`

---

### Step 8: Launching the React Web Frontend

Navigate to the web frontend directory, install npm packages, and run the development server:
```bash
cd web
npm install
npm run dev
```
The React frontend dashboard will start at `http://localhost:5173`.

---

## 📜 Makefile Command Reference

| Command | Action | Description |
| :--- | :--- | :--- |
| `make watch` | Hot Reload Dev | Runs the backend with Air live-reloading on code changes. |
| `make test` | Run Unit Tests | Executes all Go unit & integration tests (`go test -v ./...`). |
| `make migrate-up` | Run Migrations | Applies all pending SQL `up` migrations to PostgreSQL. |
| `make migrate-down N` | Rollback Migrations | Rolls back `N` database migration steps. |
| `make migrate-create <name>` | Create Migration | Generates a new pair of versioned SQL `.up.sql` and `.down.sql` files. |
| `make seed` | Seed Database | Runs the seeding script to populate database with mock data. |
| `make truncate` | Reset Tables | Truncates `comments`, `posts`, and `users` tables with `CASCADE`. |
| `make gen-docs` | Swagger Spec | Regenerates Swagger 2.0 API documentation using `swag`. |
| `make docker-up` | Start Containers | Builds and starts PostgreSQL and Redis via Docker Compose. |
| `make docker-down` | Stop Containers | Stops and removes Docker Compose containers and networks. |
| `make test-concurrency` | Test Concurrency | Runs concurrent write simulation script against post update endpoint. |

---

## 🔌 API Endpoints Summary

### Infrastructure & Diagnostics
- `GET /v1/health` - System health check (Protected via Basic Auth)
- `GET /v1/metrics` - Real-time Go runtime & database pool stats (Protected via Basic Auth)
- `GET /v1/swagger/*` - Interactive Swagger OpenAPI Web UI documentation

### Authentication & Onboarding
- `POST /v1/authentication/register` - Register a new user profile & dispatch activation email
- `POST /v1/authentication/token` - Authenticate user & issue JWT Bearer Token
- `PUT /v1/users/activate/{token}` - Activate user account using email invitation token

### User Management & Social Graph
- `GET /v1/users/{userID}` - Fetch user details (utilizes Redis caching when enabled)
- `PUT /v1/users/{userID}/follow` - Follow target user profile
- `PUT /v1/users/{userID}/unfollow` - Unfollow target user profile
- `GET /v1/users/feed` - Fetch personalized, paginated user activity feed

### Posts & Comments
- `POST /v1/posts` - Create a new post with tags
- `GET /v1/posts/{postID}` - Fetch single post details with associated comments
- `PATCH /v1/posts/{postID}` - Update post title/content (Protected by Moderator/Owner RBAC & Version check)
- `DELETE /v1/posts/{postID}` - Delete post (Protected by Admin/Owner RBAC)

---

## 🧪 Testing & Benchmarking

### Unit & Integration Tests
To execute all backend tests:
```bash
make test
```

To run tests in a specific package:
```bash
go test -v ./cmd/api
```

---

### Concurrency & Optimistic Locking Testing
To test the backend's resilience against race conditions during concurrent post modifications:
```bash
make test-concurrency
```
This triggers `scripts/test_concurrency.go` which fires simultaneous parallel HTTP requests against the same post ID to verify that optimistic locking (`version` column checking) prevents race conditions.

---

### Rate Limiter & Load Testing
You can benchmark API performance and rate limiter thresholds using `autocannon`:

```bash
# Test rate limiter (e.g. sending 22 requests per second)
npx autocannon -r 22 -d 1 -c 1 --renderStatusCodes http://localhost:8080/v1/health
```

```bash
# Test Redis caching performance boost on user endpoint
npx autocannon http://localhost:8080/v1/users/1 \
  --connections 10 \
  --duration 5 \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>"
```

---

## 🚢 CI/CD & Automated Release Pipeline

This project incorporates automated **GitHub Actions** workflows located in `.github/workflows/`:

1. **`audit.yaml`**: Runs automatically on every push or pull request to `main`. It sets up Go, downloads dependencies, runs `staticcheck` static analysis, and executes all unit tests.
2. **`release-please.yaml`**: Automatically generates GitHub Releases, bumps application version numbers, and maintains `CHANGELOG.md` based on [Conventional Commits](https://www.conventionalcommits.org/).
   - Prefix commit messages with `feat:` for minor version upgrades (e.g., `feat: add post search filter`).
   - Prefix commit messages with `fix:` for patch upgrades (e.g., `fix: resolve jwt token parsing bug`).
3. **`update-api-version.yaml`**: Automates API version synchronization across configuration files upon release approval.

---

## 🌐 Deployment Guide

To deploy the **Post Ur World** backend to production platforms (e.g., Supabase PostgreSQL + Docker / AWS / Render / Railway):

1. **Database Setup (e.g., Supabase):**
   - Provision a PostgreSQL database on Supabase or AWS RDS.
   - Update your production `DB_URL` environment variable with the SSL connection string:
     ```env
     DB_URL=postgresql://<user>:<password>@<host>:5432/<dbname>?sslmode=require
     ```
   - Apply production database migrations:
     ```bash
     migrate -path=./cmd/migrate/migrations -database="$DB_URL" up
     ```

2. **Containerized Server Deployment:**
   - Build the lightweight Docker image using the provided multi-stage `Dockerfile`:
     ```bash
     docker build -t post-ur-world-api:latest .
     ```
   - Run container with production environment flags:
     ```bash
     docker run -d -p 8080:8080 --env-file .env post-ur-world-api:latest
     ```

---

<p center="align">
  <b>Built with ❤️ using Go, PostgreSQL, Redis & React</b>
</p>
