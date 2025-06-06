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
make test         # Run all tests

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
├── migrations/                 # Database schema migrations
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