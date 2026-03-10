# go-backend-package

A production-ready Go library providing modular, composable infrastructure components for building backend services. Extracted from the [Go Backend Framework](https://github.com/chan-jui-huang/go-backend-framework).

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25.4-blue.svg)](go.mod)

## Features

- **11 Modular Packages**: Each independent, can be used standalone
- **Zero Cross-Package Dependencies**: Composable architecture with no tight coupling
- **Production-Ready**: Connection pooling, log rotation, scheduling, and error handling
- **Security-First**: Ed25519 JWT auth, Argon2id password hashing, CSPRN random generation
- **Config-Driven Bootstrap**: YAML loading with environment variable expansion
- **Database Agnostic**: GORM support for MySQL, PostgreSQL, SQLite
- **Structured Logging**: Zap-based with log rotation via Lumberjack

## Packages

| Package | Purpose | Key Features |
|---------|---------|--------------|
| **booter** | Config bootstrap | YAML parsing, env expansion, config registry |
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
    "github.com/chan-jui-huang/go-backend-package/pkg/booter"
    "github.com/chan-jui-huang/go-backend-package/pkg/database"
)

func main() {
    loader := booter.BootConfigLoader(booter.NewConfigWithCommand())

    dbConfig := &database.Config{}
    loader.Unmarshal("database", dbConfig)

    db := database.New(dbConfig)
    _ = db
}
```

## Core Architecture

### Config Loader

**Config Loader**: Wraps Viper for configuration unmarshaling
- Environment variable expansion: `${VAR_NAME}`
- Batch or per-key unmarshaling
- Accessed through `booter.BootConfigLoader(...)`

### Bootstrap Flow

```
1. Build booter config
2. Read YAML config file
3. Expand environment variables in YAML
4. Parse into Viper
5. Return config loader
```

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
- `booter` only handles config loading

### Factory Pattern

All infrastructure packages follow consistent API:
```go
New(config Config) *Client
```

### Config Bootstrap

Load config and unmarshal only the sections you need:
```go
loader := booter.BootConfigLoader(booter.NewConfigWithCommand())

dbConfig := &database.Config{}
loader.Unmarshal("database", dbConfig)
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
├── argon2/           # Password hashing
├── authentication/   # JWT authentication
├── booter/           # Config bootstrap
│   └── config/       # Config loader
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
loader := booter.BootConfigLoader(booter.NewConfigWithCommand())

dbConfig := &database.Config{}
loader.Unmarshal("database", dbConfig)

db := database.New(dbConfig)
```

### Setup Logger

```go
loader := booter.BootConfigLoader(booter.NewConfigWithCommand())

logConfig := &logger.Config{}
loader.Unmarshal("logger", logConfig)

log, _ := logger.NewLogger(logConfig, logger.JsonEncoder)
_ = log
```

### Configure Authentication

```go
loader := booter.BootConfigLoader(booter.NewConfigWithCommand())

authConfig := &authentication.Config{}
loader.Unmarshal("authentication", authConfig)

auth, _ := authentication.NewAuthenticator(authConfig)
_ = auth
```

### Schedule Jobs

```go
scheduler.Scheduler.BacklogJobs(map[string]scheduler.Job{
    "cleanup": &CleanupJob{},
    "sync":    &SyncJob{},
})

scheduler.Scheduler.Start()
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## References

- [Go Backend Framework](https://github.com/chan-jui-huang/go-backend-framework)
- [AGENTS.md](AGENTS.md) - Detailed developer guide
