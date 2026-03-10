## Project Overview

`go-backend-package` is a reusable Go library (1,888 LOC) providing core infrastructure for building production-ready backend services. Extracted from the [Go Backend Framework](https://github.com/chan-jui-huang/go-backend-framework), it's designed as a foundational dependency with modular, composable components. Apache 2.0 licensed. Go 1.25.4 required.

## Common Commands

### Testing
```bash
make test                     # Run all tests
make test args=./pkg/argon2  # Run tests for specific package
make benchmark                # Run benchmarks with memory stats
make benchmark args=./pkg/random  # Benchmark specific package
```

### Code Quality
```bash
make linter    # Run golangci-lint (5-min timeout, checks: errcheck, gosec)
```

## Architecture Overview

### Core Pattern: Config Loader

The `booter` package now focuses only on configuration loading:

1. **Config Loader** (`pkg/booter/config/loader.go`)
   - Wraps `viper.Viper` for configuration unmarshaling
   - Supports environment variable expansion (`${VAR_NAME}`) in YAML
   - Exposes `Unmarshal(key, config)` and `UnmarshalMany(configs)`
   - Created via `config.New(v)` and usually obtained from `booter.BootConfigLoader(...)`

### Bootstrap Flow

The `booter` package builds a config loader from YAML:

```
BootConfigLoader() {
    read config file
    expand environment variables
    parse yaml with viper
    return config loader
}
```

**Configuration Loading**:
- YAML file: `{rootDir}/{configFileName}` (default: `./config.yml`)
- CLI flags: `--rootDir`, `--configFileName`, `--debug`
- Typical usage:
  `loader := booter.BootConfigLoader(booter.NewConfigWithCommand())`
  `loader.Unmarshal("database", &database.Config{})`

---

## Package Details

### 1. Database (`pkg/database/`)
GORM-based factory for MySQL, PostgreSQL, SQLite with connection pooling.

**Features**:
- Factory function: `New(config Config) *gorm.DB` returns configured instance
- Driver-specific constructors: `NewMySqlDatabase()`, `NewPgSqlDatabase()`, `NewSqliteDatabase()`
- GORM optimization: `SkipDefaultTransaction: true`, `DisableNestedTransaction: true`, `PrepareStmt: true`
- Connection pooling: `SetMaxOpenConns()`, `SetMaxIdleConns()`, `SetConnMaxLifetime()`
- Log level mapping: Config.LogLevel → gorm logger.LogMode

**Config fields**:
- Driver, Host, Port, Username, Password, Database
- MaxOpenConns, MaxIdleConns, ConnMaxLifetime
- LogLevel (Info/Warn/Error/Silent)

---

### 2. Logger (`pkg/logger/`)
Zap-based structured logging with lumberjack rotation.

**Features**:
- Two types: Console (stdout) or File (with rotation)
- Two encoders: JSON (structured) or Console (human-readable)
- Log rotation: MaxSize, MaxBackups, MaxAge, Compress flags
- Sampling enabled by default: First 100 logs per second, then 5 per second
- Caller tracking enabled
- Log levels: Debug, Info, Warn, Error, Dpanic, Panic, Fatal

**Config fields**:
- Type (console|file), LogPath, MaxSize, MaxBackups, MaxAge, Compress, Level
- Factory: `NewLogger(config, encoder, opts...)` → `(*zap.Logger, error)`

---

### 3. Authentication (`pkg/authentication/`)
JWT authentication using Ed25519 (EdDSA) digital signatures.

**Security approach**:
- Ed25519: More secure than RSA/HMAC, resistant to quantum attacks
- Base64-encoded keys (RawURLEncoding)
- Token expiration validation
- Constant-time signature verification

**Methods**:
- `IssueAccessToken(subject string)` → token with "access" audience + configured lifetime
- `VerifyJwt(tokenString string)` → validates signature, expiration, returns claims
- `IssueJwt(claims jwt.MapClaims)` → generic token with custom claims

**Config fields**:
- PrivateKey, PublicKey (base64 Ed25519 keys)
- AccessTokenLifeTime, RefreshTokenLifeTime (time.Duration)

---

### 4. Scheduler (`pkg/scheduler/`)
Cron job scheduling with second-level precision via robfig/cron/v3.

**Architecture**:
- Global singleton: `Scheduler` instance initialized in `init()`
- `Job` interface: `GetFrequency() string` (cron expression), `Execute()` (logic)
- Backlog pattern: Queue jobs before scheduler starts
- Lifecycle: `BacklogJobs()` → `Start()` → schedules all → `Stop()` returns context for shutdown

**Methods**:
- `BacklogJobs(map[string]Job)`: Queue jobs before start
- `Start()`: Schedules backlog, begins cron execution
- `AddJob(key, job)`: Add job after start
- `RemoveJob(key)`: Remove scheduled job
- `Stop()`: Return shutdown context

---

### 5. Pagination (`pkg/pagination/`)
Database query pagination with GORM, supporting dynamic filters and ordering.

**Components**:
- `Paginator`: Encapsulates pagination logic and GORM DB instance
- `PaginationRequest`: API input (Page, PerPage, OrderBy + validation tags)
- `PaginationResponse`: API output (LastPage, Total)

**Methods**:
- `AddWhereConditions(whereMap map[string]func(*gorm.DB) *gorm.DB)`: Apply filters
- `OrderBy(orderMap map[string]string)`: Apply ordering
- `GetTotalAndLastPage()`: Calculate metadata
- `Execute(model interface{}) *gorm.DB`: Run paginated query with offset/limit

**Pattern**: Map column names to filter functions or order directions for type-safe dynamic queries.

---

### 6. Authentication Utilities

#### Argon2 (`pkg/argon2/`)
Password hashing using Argon2id (GPU-resistant, side-channel resistant).

**Features**:
- Default config: 64MB memory, 1 iteration, 2 threads, 32-byte key, 16-byte salt
- PHC string format: `$argon2id$v=19$m=65536,t=1,p=2$salt$hash`
- Constant-time verification via hmac.Equal()

**Functions**:
- `MakeArgon2IdHash(password string)` → hash
- `MakeArgon2IdHashWithConfig(password, config)` → hash with custom params
- `VerifyArgon2IdHash(password, hash string)` → bool

#### Random (`pkg/random/`)
Cryptographically secure random string generation (crypto/rand CSPRNG).

**Features**:
- Base64 encoding with '/' and '+' stripped (URL-safe subset)
- Handles edge cases where encoding yields fewer chars than requested
- `RandomString(n int)` → exactly n-character string

---

### 7. Redis (`pkg/redis/`)
Go-redis/v9 client wrapper.

**Factory**: `New(config Config) *redis.Client`

**Config fields**: Address, Password, DB, MinIdleConns, ConnMaxLifetime

---

### 8. ClickHouse (`pkg/clickhouse/`)
Official ClickHouse Go driver v2 client initialization.

**Features**:
- Compression: LZ4 enabled
- Dial timeout: 30 seconds
- Connection pooling: MaxOpenConns, MaxIdleConns (default 5)
- Debug mode support with custom debug function

**Config fields**: Addr ([]string), Database, Username, Password, pool settings

---

### 9. Stacktrace (`pkg/stacktrace/`)
Extract stack traces from errors wrapped with pkg/errors.

**Function**: `GetStackStrace(err error) []string` → slice of frame strings (empty for nil errors)

---

## Testing

### Test Coverage
- **argon2/argon2_test.go**: Hash → verify assertion test using testify
- **random/random_test.go**: Parallel table tests with subtests (1-100 chars)
- **authentication/authenticator_test.go**: Suite-based tests with keypair setup

### Philosophy
Focus on cryptographic and core utilities; infrastructure packages tested via integration testing in consuming applications.

### Running Tests
```bash
make test                      # All tests
make test args=./pkg/argon2   # Specific package
go test -run TestNamePrefix    # Run specific test
go test -v                     # Verbose output
```

---

## Development Workflow

### Project Structure
```
go-backend-package/
├── .github/workflows/
│   ├── pr-labeler.yml       # Auto-label PRs by branch pattern
│   └── release-drafter.yml  # Auto-generate release notes
├── pkg/                     # 11 packages (argon2, authentication, booter, clickhouse, database, logger, pagination, random, redis, scheduler, stacktrace)
├── .golangci.yml            # Linter config (5-min timeout, errcheck+gosec)
├── go.mod, go.sum
├── Makefile
├── LICENSE (Apache 2.0)
└── README.md
```

### Linting Configuration
- **Timeout**: 5 minutes
- **Enabled linters**: errcheck, gosec
- **Gosec rules**: G101-G602 (hardcoded credentials, weak crypto, SQL injection, etc.)
- **Exclusions**: Test files exempt from gosec/errcheck

### GitHub Automation

**PR Labeler** (pr-labeler.yml):
- Auto-labels on PR open based on branch name:
  - `feat/*` → feat
  - `fix/*`, `hotfix/*` → fix/hotfix
  - `chore/*`, `test/*`, `refactor/*`, `docs/*` → corresponding labels

**Release Drafter** (release-drafter.yaml):
- Triggers on merge to main
- Auto-generates release notes with semantic versioning
- Groups: Features, Bug Fixes, Maintenance
- Labels determine version bump: major, minor, patch

### Typical Branch Flow
```
1. Create feature branch (feat/my-feature)
2. Develop locally
3. make linter && make test
4. Push & create PR
   → Auto-labeled by branch name
5. Merge to main
   → Release notes auto-drafted
6. Tag version
   → Release published with auto-generated notes
```

### Commit Convention: Conventional Commits 1.0.0

This project follows [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/#specification) for standardized commit messages.

**Format**:
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Type** (required):
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `style` - Code style changes (formatting, missing semicolons, etc.)
- `refactor` - Code refactoring without feature/bug changes
- `perf` - Performance improvements
- `test` - Test additions or modifications
- `chore` - Build, CI/CD, dependency updates
- `ci` - CI/CD configuration changes

**Scope** (optional):
- Package or component name: `feat(logger)`, `fix(database)`, `docs(pagination)`

**Description** (required):
- Imperative mood: use "add" not "added", "fix" not "fixed"
- No capitalization, no period at end
- Short summary: ~50 characters

**Body** (optional):
- Motivates the change and contrasts with previous behavior
- Separated by blank line from description
- Wrapped at 72 characters

**Footer** (optional):
- Reference issues: `Closes #123`, `Fixes #456`
- Breaking changes: `BREAKING CHANGE: description`

**Examples**:
```
feat(authentication): add JWT token refresh endpoint

Implement automatic token refresh mechanism for better user experience.
Tokens now refresh automatically when 80% expired.

Closes #89
```

```
fix(database): resolve connection pool leak on error

Previously, failed connections weren't properly returned to the pool,
causing gradual pool exhaustion. Now all paths return connections.

Fixes #124
```

```
docs: update quickstart guide with scheduler examples
```

```
chore(deps): upgrade Go to 1.25.4
```

**Branch Naming Convention** (aligns with commit type):
- Feature: `feat/my-feature`
- Fix: `fix/issue-description` or `hotfix/critical-bug`
- Documentation: `docs/component-guide`
- Chore: `chore/dependency-update`
- Testing: `test/auth-coverage`
- Refactoring: `refactor/registry-pattern`

---

## Usage Example: Configuration Bootstrap

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

---

## Quick Reference: Common Patterns

### Access Config
```go
loader := booter.BootConfigLoader(booter.NewConfigWithCommand())
dbConfig := &database.Config{}
loader.Unmarshal("database", dbConfig)
```

### Initialize Components From Config
```go
loader := booter.BootConfigLoader(booter.NewConfigWithCommand())

dbConfig := &database.Config{}
loader.Unmarshal("database", dbConfig)
db := database.New(dbConfig)

authConfig := &authentication.Config{}
loader.Unmarshal("authentication", authConfig)
auth, _ := authentication.NewAuthenticator(authConfig)

_ = db
_ = auth
```

### Schedule Jobs
```go
type MyJob struct{}

func (j *MyJob) GetFrequency() string { return "*/5 * * * *" } // Every 5 min
func (j *MyJob) Execute() { /* job logic */ }

scheduler.Scheduler.BacklogJobs(map[string]scheduler.Job{"myjob": &MyJob{}})
scheduler.Scheduler.Start()
```

### Pagination
```go
pagination := &pagination.Paginator{
    Db: db,
    WhereConditionMap: map[string]func(*gorm.DB) *gorm.DB{
        "status": func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", "active") },
    },
}
pagination.AddWhereConditions()
total, lastPage := pagination.GetTotalAndLastPage()
pagination.Execute(&users)
```

---

## Summary

**go-backend-package** is a production-focused foundational library with:

**Strengths**:
- ✅ Modular, composable 11-package architecture
- ✅ Consistent APIs (Config structs, New factories)
- ✅ Security-first (Ed25519, Argon2id, CSPRNG)
- ✅ Focused bootstrap and infrastructure primitives
- ✅ Modern tooling (golangci-lint, GitHub automation)
- ✅ Production-ready (connection pooling, log rotation, scheduling)

**Core Abstractions**:
- Config loader built on top of Viper

**Best Used For**: Building RESTful APIs, microservices, and CLIs requiring structured logging, databases, auth, scheduling, and configuration loading.

## Code Modification Rules
- **Strict Scope:** Do NOT modify code that you were not explicitly asked to change. You may suggest improvements for other parts of the code, but do not apply them without clear confirmation.
- **Focus:** maintain strict focus on the specific code segments relevant to the user's request.
