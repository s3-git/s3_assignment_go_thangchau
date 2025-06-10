# Friends Management API

A REST API for managing user relationships including friendships, subscriptions, and blocking functionality. Built with Go, Gin, PostgreSQL, and SQLBoiler following Clean Architecture principles.

## Features

- Create and manage friend connections between users
- Subscribe to updates from other users
- Block users to prevent friend connections and updates
- Retrieve friends lists and common friends
- Get eligible recipients for user updates (mentions, friends, subscribers)
- Type-safe database operations with SQLBoiler
- Hot reload development environment with Air
- Comprehensive test coverage

## Prerequisites

- Docker and Docker Compose
- Make (optional, for convenience commands)

## Quick Start

### Development (with hot reload)

```bash
# Start development environment with hot reload
make dev

# Or manually
docker-compose up --build
```

The API will be available at `http://localhost:8080` with automatic hot reload when you change Go files.

### Production

```bash
# Start in background
make up

# Or manually
docker-compose up --build -d
```

## Available Commands

```bash
# Development
make dev          # Start with hot reload (foreground)
make restart      # Clean restart development environment

# Production
make up           # Build and run in background
make down         # Stop containers
make down-v       # Stop containers and remove volumes

# Testing
make test                # Run all tests
make test-coverage       # Run tests with coverage report
make test-unit          # Run unit tests only
make test-integration   # Run integration tests only

# Mock generation
make install-mockgen    # Install mockgen tool
make generate-mocks     # Generate mocks from interfaces
make clean-mocks        # Remove generated mocks
make regen-mocks        # Regenerate all mocks

# Database
make gensql       # Generate database models with SQLBoiler

# Local development (requires local PostgreSQL)
make run          # Run API locally
make install-air  # Install Air for hot reload
make air          # Run with Air locally
```

## Environment Variables

Configuration is managed through a `.env` file. Copy and modify as needed:

```bash
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=assignment-db
DB_USER=postgres
DB_PASSWORD=password
DB_SSLMODE=disable

# Server Configuration
PORT=8080
```

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `postgres` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `assignment-db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `PORT` | `8080` | Server port |

## Project Structure

```
├── cmd/api/                    # Application entry point
├── internal/                   # Private application code
│   ├── config/                 # Configuration management
│   ├── controller/             # Business logic layer (use cases)
│   ├── domain/                 # Core business entities and interfaces
│   │   ├── entities/           # Domain entities (User, Friend, etc.)
│   │   └── interfaces/         # Repository and controller interfaces
│   ├── handler/                # HTTP presentation layer (Gin routes)
│   ├── infrastructure/         # External dependencies
│   │   └── database/models/    # SQLBoiler generated models
│   └── repository/             # Data access implementations
├── mocks/                      # Generated test mocks (GoMock)
├── db/migrations/              # Database schema migrations
├── pkg/                        # Shared utilities and packages
│   ├── errors/                 # Error handling utilities
│   ├── response/               # Response formatting
│   ├── utils/                  # General utilities
│   └── validator/              # Input validation
├── sqlboiler_config/           # SQLBoiler configuration
├── tmp/                        # Temporary build files
├── CLAUDE.md                   # Project guidance for AI assistants
├── Dockerfile                  # Container definition
├── Makefile                    # Build and development commands
├── docker-compose.yaml         # Docker services definition
├── go.mod                      # Go module definition
└── go.sum                      # Go module checksums
```

## Development Workflow

1. **Start development environment:**
   ```bash
   make dev
   ```

2. **Make changes to Go files** - they will automatically rebuild and restart

3. **Run tests:**
   ```bash
   make test
   ```

4. **Stop development:**
   ```bash
   Ctrl+C
   ```

## Hot Reload

The development setup uses [Air](https://github.com/air-verse/air) for hot reload:

- Watches `.go` files for changes
- Automatically rebuilds and restarts the application
- Build logs are saved to `tmp/build-errors.log`
- Excludes test files from watching

## Database

PostgreSQL 15 with automatic migrations on startup. The database is accessible at `localhost:5432` with credentials:
- Database: `assignment-db`
- User: `postgres` 
- Password: `password`

## API Endpoints

The API will be available at `http://localhost:8080` once running.

### User Management Endpoints

All endpoints are under `/api/v1/user`

#### Create Friendship
- **POST** `/api/v1/user/friends`
- Creates a friendship connection between two users
- **Request:**
  ```json
  {
    "friends": ["user1@example.com", "user2@example.com"]
  }
  ```
- **Response:**
  ```json
  {
    "success": true
  }
  ```

#### Get Friend List
- **POST** `/api/v1/user/friends/list`
- Retrieves all friends for a specific user
- **Request:**
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response:**
  ```json
  {
    "success": true,
    "friends": ["friend1@example.com", "friend2@example.com"],
    "count": 2
  }
  ```

#### Get Common Friends
- **POST** `/api/v1/user/friends/common`
- Retrieves mutual friends between two users
- **Request:**
  ```json
  {
    "friends": ["user1@example.com", "user2@example.com"]
  }
  ```
- **Response:**
  ```json
  {
    "success": true,
    "friends": ["mutual1@example.com", "mutual2@example.com"],
    "count": 2
  }
  ```

#### Create Subscription
- **POST** `/api/v1/user/subscriptions`
- Creates a subscription where requestor subscribes to target's updates
- **Request:**
  ```json
  {
    "requestor": "subscriber@example.com",
    "target": "publisher@example.com"
  }
  ```
- **Response:**
  ```json
  {
    "success": true
  }
  ```

#### Create Block
- **POST** `/api/v1/user/blocks`
- Blocks a user and removes any existing friendship/subscription
- **Request:**
  ```json
  {
    "requestor": "blocker@example.com",
    "target": "blocked@example.com"
  }
  ```
- **Response:**
  ```json
  {
    "success": true
  }
  ```

#### Get Update Recipients
- **POST** `/api/v1/user/recipients`
- Gets all users who should receive updates from a sender
- **Request:**
  ```json
  {
    "sender": "sender@example.com",
    "text": "Hello @mention@example.com, this is an update!"
  }
  ```
- **Response:**
  ```json
  {
    "success": true,
    "recipients": ["friend1@example.com", "subscriber@example.com", "mention@example.com"]
  }
  ```

## Testing

This project uses a comprehensive testing strategy with unit tests, integration tests, and mocking.

### Test Structure

```
├── internal/
│   ├── controller/
│   │   ├── user_controller_test.go      # Unit tests for business logic
│   │   └── user_controller_test_example.go  # GoMock usage examples
│   ├── handler/
│   │   └── user_handler_test.go         # Integration tests for HTTP layer
│   └── repository/
│       └── user_repository_test.go      # Integration tests for data layer
└── mocks/
    └── mock_repository.go               # Generated mocks
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run only unit tests (internal package)
make test-unit

# Run integration tests (requires database)
make test-integration

# View coverage report (after running test-coverage)
open coverage.html
```

### Mock Generation

This project uses [GoMock](https://github.com/uber-go/mock) for generating test mocks:

```bash
# Install mockgen tool
make install-mockgen

# Generate mocks from interfaces
make generate-mocks

# Clean and regenerate all mocks
make regen-mocks
```

#### Mock Usage Example

```go
func TestCreateFriendship(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
    
    // Setup expectations
    user1 := &entities.User{ID: 1, Email: "user1@example.com"}
    user2 := &entities.User{ID: 2, Email: "user2@example.com"}
    
    mockRepo.EXPECT().GetUserByEmail("user1@example.com").Return(user1, nil)
    mockRepo.EXPECT().GetUserByEmail("user2@example.com").Return(user2, nil)
    mockRepo.EXPECT().CheckBidirectionalBlock(1, 2).Return(false, nil)
    mockRepo.EXPECT().CreateFriendship(user1, user2).Return(nil)
    
    controller := NewUserController(mockRepo)
    err := controller.CreateFriendship("user1@example.com", "user2@example.com")
    
    assert.NoError(t, err)
}
```

### Test Categories

1. **Unit Tests**: Test business logic in isolation using mocks
   - Controller layer tests (`*_controller_test.go`)
   - Pure function tests
   - Mock all external dependencies

2. **Integration Tests**: Test components working together
   - Handler tests with real database via testcontainers
   - Repository tests with real database
   - End-to-end API tests

3. **Test Data**: Uses testcontainers for isolated database testing
   - Automatic PostgreSQL container setup
   - Database migrations run automatically
   - Clean state for each test

### Dependencies

- **GoMock**: Mock generation from interfaces
- **Testify**: Assertions and test utilities  
- **Testcontainers**: Database integration testing
- **SQLBoiler**: Type-safe database models and queries

## Troubleshooting

**Port already in use:**
```bash
make down
# Or kill specific processes using the port
```

**Database connection issues:**
```bash
# Check if PostgreSQL container is healthy
docker-compose ps
```

**Hot reload not working:**
```bash
# Restart development environment
make restart
```