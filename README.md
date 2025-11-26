# go-backend-package

A production-ready Go library providing modular, composable infrastructure components for building backend services. Extracted from the [Go Backend Framework](https://github.com/chan-jui-huang/go-backend-framework).

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-blue.svg)](go.mod)

## Features

- **12 Modular Packages**: Each independent, can be used standalone
- **Zero Cross-Package Dependencies**: Composable architecture with no tight coupling
- **Production-Ready**: Connection pooling, log rotation, graceful shutdown, error handling
- **Security-First**: Ed25519 JWT auth, Argon2id password hashing, CSPRN random generation
- **Flexible Lifecycle Management**: Event-driven callbacks for startup/shutdown orchestration
- **Database Agnostic**: GORM support for MySQL, PostgreSQL, SQLite
- **Structured Logging**: Zap-based with log rotation via Lumberjack

## Packages

| Package | Purpose | Key Features |
|---------|---------|--------------|
| **app** | Lifecycle management | Signal handling, graceful shutdown, callback phases |
| **booter** | Bootstrap & DI | Config/service registries, YAML parsing, plugin architecture |
| **database** | GORM wrapper | MySQL, PostgreSQL, SQLite drivers, connection pooling |
| **logger** | Structured logging | Zap, log rotation, console/file output, sampling |
| **authentication** | JWT auth | Ed25519 signatures, token lifecycle, claims validation |
| **scheduler** | Cron jobs | Second-precision scheduling, dynamic job management |
| **pagination** | Query pagination | GORM integration, dynamic filters, ordering |
| **redis** | Redis client | Wrapper around go-redis/v9 |
| **clickhouse** | ClickHouse client | Official driver, LZ4 compression, connection pooling |
| **argon2** | Password hashing | Argon2id, GPU-resistant, constant-time verification |
| **random** | Random strings | Cryptographically secure (CSPRNG), URL-safe |
| **stacktrace** | Error handling | Stack trace extraction from wrapped errors |

## Quick Start

### Installation

```bash
go get github.com/chan-jui-huang/go-backend-package
```

### Basic Usage

```go
package main

import (
    "github.com/chan-jui-huang/go-backend-package/pkg/app"
    "github.com/chan-jui-huang/go-backend-package/pkg/booter"
    "github.com/chan-jui-huang/go-backend-package/pkg/booter/service"
)

// Create a registrar for your application components
type AppRegistrar struct{}

func (r *AppRegistrar) Boot() {
    // Initialize your dependencies
}

func (r *AppRegistrar) Register() {
    // Register services with service.Registry
}

func main() {
    // Bootstrap: load config and initialize services
    booter.Boot(
        loadEnv,
        booter.NewConfigWithCommand,
        booter.NewRegistrarCenter([]booter.Registrar{
            &AppRegistrar{},
        }),
    )

    // Define application lifecycle
    myApp := app.New(
        []func(){},           // Starting callbacks
        []func(){},           // Started callbacks
        []app.SignalCallback{}, // Signal handlers
        []func(){},           // Async callbacks
        []func(){},           // Terminated callbacks
    )

    // Run application
    myApp.Run(func() {
        // Main application logic
        // Access services: service.Registry.Get("myService")
    })
}

func loadEnv() {
    // Load .env file or environment variables
}
```

## Core Architecture

### Dual-Registry System

**Config Registry**: Stores application configuration unmarshaled from YAML via Viper
- Environment variable expansion: `${VAR_NAME}`
- Type-safe configuration objects
- Accessed: `config.Registry.Get(key)`

**Service Registry**: Stores initialized service instances (dependency injection container)
- Stores pointers or functions
- Accessed: `service.Registry.Set(key, instance)`, `service.Registry.Get(key)`

### Bootstrap Flow

```
1. Load environment variables
2. Parse YAML config (with env expansion)
3. Execute registrars:
   - Boot() → initialize components
   - Register() → wire services into registry
4. Application ready
```

### Application Lifecycle Phases

1. **STARTING**: Sequential, blocking setup before main execution
2. **EXECUTION**: Main application logic in goroutine
3. **STARTED**: Sequential, blocking setup after main execution starts
4. **SIGNALS**: Listen for OS signals (SIGINT/SIGTERM), execute gracefully
5. **ASYNC**: Background tasks (independent goroutines)
6. **TERMINATED**: Sequential, blocking cleanup on shutdown

See [AGENTS.md](AGENTS.md) for detailed lifecycle diagram.

## Common Commands

### Testing

```bash
# Run all tests
make test

# Run tests for specific package
make test args=./pkg/argon2

# Run benchmarks with memory stats
make benchmark

# Run benchmarks for specific package
make benchmark args=./pkg/random
```

### Code Quality

```bash
# Run linter (golangci-lint, 5-min timeout)
# Checks: errcheck, gosec
make linter
```

## Configuration

Configuration is loaded from a YAML file (default: `./config.yml`).

### CLI Flags

```bash
--rootDir string       # Root directory (default: current working directory)
--configFileName string # Config file name (default: "config.yml")
--debug bool          # Enable debug mode
--testing bool        # Enable testing mode
```

### Environment Variables

YAML supports environment variable expansion:

```yaml
database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  username: ${DB_USER}
  password: ${DB_PASSWORD}
```

## Architecture Highlights

### Package Independence

- No internal cross-package dependencies
- Each package can be used standalone
- Only orchestration point: booter's registries

### Factory Pattern

All infrastructure packages follow consistent API:
```go
New(config Config) *Client
```

### Plugin Architecture

Implement `Registrar` interface for modular startup:
```go
type Registrar interface {
    Boot()      // Initialize component
    Register()  // Register with registries
}
```

## Security

- **Authentication**: Ed25519 digital signatures (quantum-resistant)
- **Password Hashing**: Argon2id (GPU-resistant, side-channel resistant)
- **Random Generation**: Cryptographically secure via crypto/rand
- **Linter**: golangci-lint with gosec rules enabled

## Development

### Project Structure

```
pkg/
├── app/              # Application lifecycle
├── argon2/           # Password hashing
├── authentication/   # JWT authentication
├── booter/           # Bootstrap & DI
│   ├── config/       # Config registry
│   └── service/      # Service registry
├── clickhouse/       # ClickHouse client
├── database/         # GORM database wrapper
├── logger/           # Structured logging
├── pagination/       # Query pagination
├── random/           # Random string generation
├── redis/            # Redis client
├── scheduler/        # Cron job scheduling
└── stacktrace/       # Stack trace extraction
```

### Contributing

1. Create feature branch: `git checkout -b feat/my-feature`
2. Run linter & tests: `make linter && make test`
3. Create pull request
4. PR auto-labeled by branch name (feat/*, fix/*, etc.)
5. On merge to main, release notes auto-drafted

## Dependencies

### Key External Libraries

- **GORM**: Database ORM (gorm.io/gorm)
- **Zap**: Structured logging (go.uber.org/zap)
- **Viper**: Configuration management (spf13/viper)
- **JWT**: Token generation (golang-jwt/jwt)
- **Cron**: Job scheduling (robfig/cron)
- **Go-Redis**: Redis client (redis/go-redis)
- **ClickHouse**: ClickHouse driver (ClickHouse/clickhouse-go)

See [go.mod](go.mod) for complete dependency list.

## Documentation

- **[AGENTS.md](AGENTS.md)**: Detailed architecture, design patterns, and advanced usage
- **[LICENSE](LICENSE)**: Apache 2.0

## Examples

### Initialize Database

```go
type DatabaseRegistrar struct{}

func (r *DatabaseRegistrar) Boot() {}

func (r *DatabaseRegistrar) Register() {
    dbConfig := config.Registry.Get("database").(*database.Config)
    db := database.New(dbConfig)
    service.Registry.Set("db", db)
}
```

### Setup Logger

```go
type LoggerRegistrar struct{}

func (r *LoggerRegistrar) Boot() {}

func (r *LoggerRegistrar) Register() {
    logConfig := config.Registry.Get("logger").(*logger.Config)
    logger, _ := logger.NewLogger(logConfig, logger.JsonEncoder)
    service.Registry.Set("logger", logger)
}
```

### Configure Authentication

```go
type AuthRegistrar struct{}

func (r *AuthRegistrar) Boot() {}

func (r *AuthRegistrar) Register() {
    authConfig := config.Registry.Get("authentication").(*authentication.Config)
    auth, _ := authentication.NewAuthenticator(authConfig)
    service.Registry.Set("auth", auth)
}
```

### Schedule Jobs

```go
type SchedulerRegistrar struct{}

func (r *SchedulerRegistrar) Boot() {}

func (r *SchedulerRegistrar) Register() {
    scheduler.Scheduler.BacklogJobs(map[string]scheduler.Job{
        "cleanup": &CleanupJob{},
        "sync":    &SyncJob{},
    })
}
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## References

- [Go Backend Framework](https://github.com/chan-jui-huang/go-backend-framework)
- [AGENTS.md](AGENTS.md) - Detailed developer guide
