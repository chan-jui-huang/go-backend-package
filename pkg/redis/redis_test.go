package redis_test

import (
	"testing"
	"time"

	redis "github.com/chan-jui-huang/go-backend-package/v2/pkg/redis"
	"github.com/stretchr/testify/require"
)

func TestNewReturnsClient(t *testing.T) {
	cfg := redis.Config{Address: "127.0.0.1:6379"}
	c := redis.New(cfg)
	require.NotNil(t, c)
}

func TestNewClientOptions(t *testing.T) {
	cfg := redis.Config{
		Address:         "localhost:6379",
		Password:        "secret",
		DB:              2,
		MinIdleConns:    5,
		ConnMaxLifetime: 30 * time.Second,
	}
	c := redis.New(cfg)
	opts := c.Options()

	require.Equal(t, cfg.Address, opts.Addr)
	require.Equal(t, cfg.Password, opts.Password)
	require.Equal(t, cfg.DB, opts.DB)
	require.Equal(t, cfg.MinIdleConns, opts.MinIdleConns)
	require.Equal(t, cfg.ConnMaxLifetime, opts.ConnMaxLifetime)
}

func TestNewClientZeroValues(t *testing.T) {
	cfg := redis.Config{Address: "redis.example:1234"}
	c := redis.New(cfg)
	opts := c.Options()

	require.Equal(t, cfg.Address, opts.Addr)
	// zero values should be preserved
	require.Equal(t, "", opts.Password)
	require.Equal(t, 0, opts.DB)
}
