package clickhouse_test

import (
	"reflect"
	"testing"

	clickhouse "github.com/chan-jui-huang/go-backend-package/v2/pkg/clickhouse"
)

func TestConfigFields(t *testing.T) {
	var c clickhouse.Config
	typ := reflect.TypeOf(c)
	expected := []string{"Addr", "Database", "Username", "Password", "MaxOpenConns", "MaxIdleConns"}
	for _, name := range expected {
		if _, ok := typ.FieldByName(name); !ok {
			t.Fatalf("expected Config to have field %s", name)
		}
	}
}

func TestNewWithEmptyAddrReturnsError(t *testing.T) {
	t.Skip("requires ClickHouse server or mock; blocked in CI")
}

func TestNewOptionsBlocked(t *testing.T) {
	t.Skip("requires ClickHouse server or mock; blocked in CI")
}
