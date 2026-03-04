package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewReturnsClient(t *testing.T) {
	cfg := Config{Address: "127.0.0.1:6379"}
	c := New(cfg)
	require.NotNil(t, c)
}

func TestNewClientOptions(t *testing.T) {
	cfg := Config{
		Address:         "localhost:6379",
		Password:        "secret",
		DB:              2,
		MinIdleConns:    5,
		ConnMaxLifetime: 30 * time.Second,
	}
	c := New(cfg)
	opts := c.Options()

	require.Equal(t, cfg.Address, opts.Addr)
	require.Equal(t, cfg.Password, opts.Password)
	require.Equal(t, cfg.DB, opts.DB)
	require.Equal(t, cfg.MinIdleConns, opts.MinIdleConns)
	require.Equal(t, cfg.ConnMaxLifetime, opts.ConnMaxLifetime)
}

func TestNewClientZeroValues(t *testing.T) {
	cfg := Config{Address: "redis.example:1234"}
	c := New(cfg)
	opts := c.Options()

	require.Equal(t, cfg.Address, opts.Addr)
	// zero values should be preserved
	require.Equal(t, "", opts.Password)
	require.Equal(t, 0, opts.DB)
}
