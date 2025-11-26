## Project Overview

`go-backend-package` is a reusable Go library (1,060 LOC) providing core infrastructure for building production-ready backend services. Extracted from the [Go Backend Framework](https://github.com/chan-jui-huang/go-backend-framework), it's designed as a foundational dependency with modular, composable components. Apache 2.0 licensed. Go 1.25.4 required.

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

### Core Pattern: Dual-Registry System

The application uses two global singleton registries for dependency injection:

1. **Config Registry** (`pkg/booter/config/registry.go`)
   - Stores configuration objects unmarshaled from YAML via Viper
   - Enforces pointer types via reflection
   - Supports environment variable expansion (`${VAR_NAME}`) in YAML
   - Accessed: `config.Registry.Get(key)`, `config.Registry.Register(key, configStruct)`

2. **Service Registry** (`pkg/booter/service/registry.go`)
   - Stores initialized service instances (pointers or functions)
   - Dependency injection container for the application
   - `Clone()` returns dereferenced copy; `Get()` returns raw value
   - Accessed: `service.Registry.Set(key, instance)`, `service.Registry.Get(key)`

### Bootstrap Flow

The `booter` package orchestrates startup via `Registrar` interface pattern:

```
Boot() {
    loadEnvFunc()           // Load environment (user-provided function)
    bootConfigRegistry()    // Parse YAML config (with env expansion)
    registrarCenter.BeforeExecute()  // Pre-init hook
    registrarCenter.Execute() {      // For each Registrar:
        registrar.Boot()             //   Initialize component
        registrar.Register()         //   Register with registries
    }
    registrarCenter.AfterExecute()  // Post-init hook
}
```

**Key interfaces**:
- `Registrar`: Boot() + Register() - implemented by components for modular initialization
- `RegisterExecutor`: BeforeExecute() + Execute() + AfterExecute() - lifecycle hooks around registrar execution

**Configuration Loading**:
- YAML file: `{rootDir}/{configFileName}` (default: `./config.yml`)
- CLI flags: `--rootDir`, `--configFileName`, `--debug`, `--testing`
- Config registry stores booter config: `config.Registry.Get("booter")` → `*booter.Config`

---

## Package Details

### 1. App Lifecycle (`pkg/app/`)
Orchestrates application startup, execution, and graceful shutdown.

**Components**:
- `App` struct: Manages five callback phases via `sync.WaitGroup`
- `New()`: Accepts five callback slice types for each lifecycle phase
- `Run(executorFunc)`: Orchestrates phase execution

**Phases** (in order):
1. Starting callbacks: Run before main execution
2. Main executor: Runs in goroutine
3. Started callbacks: Run after main executor starts
4. Signal callbacks: Listen for OS signals, execute cleanup handlers
5. Async callbacks: Fire-and-forget goroutines (non-blocking)
6. Terminated callbacks: Final cleanup after `wg.Wait()` completes

**Signal handling**: Each `SignalCallback` registers signal listeners that trigger context-aware handlers for graceful shutdown.

**Lifecycle Diagram**:

```
┌────────────────────────────────────────┐
│        app.Run(executor)               │
└────────────────────────────────────────┘
              │
              ▼
     ┌────────────────────┐
     │ PHASE 1: STARTING  │
     │ Sequential Blocking│
     └────────────────────┘
              │
              ▼
     ┌────────────────────┐
     │ PHASE 2: EXECUTION │
     │ go executor()      │
     │ + wg.Add(1)        │
     └────────────────────┘
              │
              ▼
     ┌────────────────────┐
     │ PHASE 3: STARTED   │
     │ Sequential Blocking│
     └────────────────────┘
              │
      ┌───────┴───────┐
      ▼               ▼
  ┌────────────┐  ┌────────────┐
  │PHASE 4A:   │  │PHASE 4B:   │
  │SIGNALS     │  │ASYNC       │
  │go routine  │  │go routine  │
  │wg.Add(N)   │  │no wg       │
  └────────────┘  └────────────┘
      │
      └───────┬───────┘
              ▼
   ══ MAIN BLOCKING ══
      wg.Wait()
              │
              ▼
    ┌────────────────────┐
    │PHASE 5: TERMINATED │
    │Sequential Blocking │
    └────────────────────┘
              │
              ▼
        Application Exit
```

**Phase Characteristics**:

| Phase      | Execution   | Blocking | WaitGroup | Purpose          |
|------------|-------------|----------|-----------|------------------|
| STARTING   | Sequential  | Yes      | No        | Pre-startup      |
| EXECUTION  | Goroutine   | No       | Yes       | Main app         |
| STARTED    | Sequential  | Yes      | No        | Post-startup     |
| SIGNALS    | Goroutines  | Yes      | Yes       | Signal handling  |
| ASYNC      | Goroutines  | No       | No        | Background tasks |
| TERMINATED | Sequential  | Yes      | No        | Cleanup & exit   |

---

### 2. Database (`pkg/database/`)
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

### 3. Logger (`pkg/logger/`)
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

### 4. Authentication (`pkg/authentication/`)
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

### 5. Scheduler (`pkg/scheduler/`)
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

### 6. Pagination (`pkg/pagination/`)
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

### 7. Authentication Utilities

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

### 8. Redis (`pkg/redis/`)
Go-redis/v9 client wrapper.

**Factory**: `New(config Config) *redis.Client`

**Config fields**: Address, Password, DB, MinIdleConns, ConnMaxLifetime

---

### 9. ClickHouse (`pkg/clickhouse/`)
Official ClickHouse Go driver v2 client initialization.

**Features**:
- Compression: LZ4 enabled
- Dial timeout: 30 seconds
- Connection pooling: MaxOpenConns, MaxIdleConns (default 5)
- Debug mode support with custom debug function

**Config fields**: Addr ([]string), Database, Username, Password, pool settings

---

### 10. Stacktrace (`pkg/stacktrace/`)
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
├── pkg/                     # 12 packages (app, argon2, auth, booter, clickhouse, db, logger, pagination, random, redis, scheduler, stacktrace)
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

---

## Usage Example: Application Bootstrap

```go
package main

import (
    "os"
    "github.com/chan-jui-huang/go-backend-package/pkg/app"
    "github.com/chan-jui-huang/go-backend-package/pkg/booter"
    "github.com/chan-jui-huang/go-backend-package/pkg/booter/config"
    "github.com/chan-jui-huang/go-backend-package/pkg/booter/service"
)

// Define registrars
type MyRegistrar struct{}

func (r *MyRegistrar) Boot() {
    // Initialize components (database, logger, etc.)
}

func (r *MyRegistrar) Register() {
    // Register services with service.Registry
}

func main() {
    // Bootstrap: load config, initialize services
    booter.Boot(
        loadEnv,                            // User function to load .env
        booter.NewConfigWithCommand,        // Parse CLI flags
        booter.NewRegistrarCenter([]booter.Registrar{&MyRegistrar{}}),
    )

    // Define lifecycle callbacks
    myApp := app.New(
        []func(){},                   // Starting
        []func(){},                   // Started
        []app.SignalCallback{},       // Signal handlers
        []func(){},                   // Async
        []func(){},                   // Terminated
    )

    // Run application
    myApp.Run(func() {
        // Main application logic
        // Access services: service.Registry.Get("myService")
    })
}
```

---

## Quick Reference: Common Patterns

### Access Config/Services
```go
// Get config
dbConfig := config.Registry.Get("database").(*database.Config)

// Get service
db := service.Registry.Get("db").(*gorm.DB)

// Clone service (dereferenced copy)
clonedService := service.Registry.Clone("service").(*MyService)
```

### Initialize & Register a Service Group
```go
// DatabaseRegistrar - registers database service
type DatabaseRegistrar struct{}

func (r *DatabaseRegistrar) Boot() {
    // Initialize database connections, migrations, etc.
}

func (r *DatabaseRegistrar) Register() {
    dbConfig := config.Registry.Get("database").(*database.Config)
    db := database.New(dbConfig)
    service.Registry.Set("db", db)
}
```

```go
// AuthenticationRegistrar - registers auth service
type AuthenticationRegistrar struct{}

func (r *AuthenticationRegistrar) Boot() {
    // Pre-auth setup if needed
}

func (r *AuthenticationRegistrar) Register() {
    authConfig := config.Registry.Get("authentication").(*authentication.Config)
    auth, _ := authentication.NewAuthenticator(authConfig)
    service.Registry.Set("auth", auth)
}
```

```go
// SchedulerRegistrar - registers scheduler with jobs
type SchedulerRegistrar struct{}

func (r *SchedulerRegistrar) Boot() {
    // Initialize job definitions
}

func (r *SchedulerRegistrar) Register() {
    scheduler.Scheduler.BacklogJobs(map[string]scheduler.Job{
        "cleanup": &CleanupJob{},
        "sync":    &SyncJob{},
    })
}
```

Then in main():
```go
booter.Boot(
    loadEnv,
    booter.NewConfigWithCommand,
    booter.NewRegistrarCenter([]booter.Registrar{
        &DatabaseRegistrar{},
        &AuthenticationRegistrar{},
        &SchedulerRegistrar{},
    }),
)
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
- ✅ Modular, composable 12-package architecture
- ✅ Consistent APIs (Config structs, New factories)
- ✅ Security-first (Ed25519, Argon2id, CSPRNG)
- ✅ Flexible lifecycle management (callback-driven, signal-aware)
- ✅ Modern tooling (golangci-lint, GitHub automation)
- ✅ Production-ready (connection pooling, log rotation, graceful shutdown)

**Core Abstractions**:
- Dual-registry DI system (config + services)
- Registrar plugin pattern for modular startup
- Lifecycle callbacks for orchestration

**Best Used For**: Building RESTful APIs, microservices, CLIs requiring structured logging, databases, auth, scheduling, and graceful lifecycle management.
