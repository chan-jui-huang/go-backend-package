# go-backend-package

A useful Go package that's a component of the [Go Backend Framework](https://github.com/chan-jui-huang/go-backend-framework).

## Run test case
make test

## Folder structure
```
go-backend-package
├── go.mod
├── go.sum
├── Makefile
├── pkg
│   ├── app
│   │   └── app.go
│   ├── argon2
│   │   ├── argon2_test.go
│   │   └── argon2.go
│   ├── authentication
│   │   ├── authenticator_test.go
│   │   ├── authenticator.go
│   │   └── config.go
│   ├── booter
│   │   ├── booter.go
│   │   ├── config
│   │   │   └── registry.go
│   │   └── service
│   │       └── registry.go
│   ├── clickhouse
│   │   ├── clickhouse.go
│   │   └── config.go
│   ├── database
│   │   ├── config.go
│   │   └── database.go
│   ├── logger
│   │   ├── config.go
│   │   └── logger.go
│   ├── pagination
│   │   └── pagination.go
│   ├── random
│   │   ├── random_test.go
│   │   └── random.go
│   ├── redis
│   │   ├── config.go
│   │   └── redis.go
│   ├── scheduler
│   │   └── scheduler.go
│   └── stacktrace
│       └── stacktrace.go
└── README.md
```
