package database_test

import (
	"testing"
	"time"

	database "github.com/chan-jui-huang/go-backend-package/v2/pkg/database"
	"gorm.io/gorm/logger"
)

func TestGetDriverValidAndInvalid(t *testing.T) {
	// valid
	if got := database.GetDriver(database.MySql); got != database.MySql {
		t.Fatalf("expected MySql, got %v", got)
	}
	if got := database.GetDriver(database.PgSql); got != database.PgSql {
		t.Fatalf("expected PgSql, got %v", got)
	}
	if got := database.GetDriver(database.Sqlite); got != database.Sqlite {
		t.Fatalf("expected Sqlite, got %v", got)
	}

	// invalid should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid driver")
		}
	}()
	// This should panic
	database.GetDriver(database.Driver("invalid"))
}

func TestGetGormLogLevelMapping(t *testing.T) {
	if lvl := database.GetGormLogLevel(database.Info); lvl != logger.Info {
		t.Fatalf("expected logger.Info, got %v", lvl)
	}
	if lvl := database.GetGormLogLevel(database.Warn); lvl != logger.Warn {
		t.Fatalf("expected logger.Warn, got %v", lvl)
	}
	if lvl := database.GetGormLogLevel(database.Error); lvl != logger.Error {
		t.Fatalf("expected logger.Error, got %v", lvl)
	}
	if lvl := database.GetGormLogLevel(database.Silent); lvl != logger.Silent {
		t.Fatalf("expected logger.Silent, got %v", lvl)
	}
	// default
	if lvl := database.GetGormLogLevel(database.LogLevel("unknown")); lvl != logger.Info {
		t.Fatalf("expected default logger.Info, got %v", lvl)
	}
}

func TestNewSqliteConfigAppliedAndPool(t *testing.T) {
	cfg := database.Config{
		Driver:          database.Sqlite,
		Database:        "file::memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		LogLevel:        database.Warn,
	}

	db := database.New(cfg)
	if db == nil {
		t.Fatalf("expected non-nil db")
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// validate gorm config flags propagated
	if db.Config.SkipDefaultTransaction != true {
		t.Fatalf("expected SkipDefaultTransaction=true")
	}
	if db.Config.DisableNestedTransaction != true {
		t.Fatalf("expected DisableNestedTransaction=true")
	}
	if db.Config.PrepareStmt != true {
		t.Fatalf("expected PrepareStmt=true")
	}

	// validate connection pool MaxOpenConns applied via sql.DB.Stats
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error: %v", err)
	}
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != cfg.MaxOpenConns {
		t.Fatalf("expected MaxOpenConnections=%d, got %d", cfg.MaxOpenConns, stats.MaxOpenConnections)
	}

	// simple query to ensure DB works
	execErr := db.Exec("CREATE TABLE IF NOT EXISTS test_tbl (id INTEGER PRIMARY KEY, name TEXT);").Error
	if execErr != nil {
		t.Fatalf("failed to execute create table: %v", execErr)
	}

	res := db.Exec("INSERT INTO test_tbl(name) VALUES(?)", "alice")
	if res.Error != nil {
		t.Fatalf("insert failed: %v", res.Error)
	}

	// query back
	var name string
	row := db.Raw("SELECT name FROM test_tbl WHERE id = ?", 1).Row()
	err = row.Scan(&name)
	if err != nil {
		t.Fatalf("query scan failed: %v", err)
	}
	if name != "alice" {
		t.Fatalf("expected name alice, got %s", name)
	}
}

func TestNewInvalidDriverPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid driver in New")
		}
	}()
	database.New(database.Config{Driver: database.Driver("invalid")})
}
