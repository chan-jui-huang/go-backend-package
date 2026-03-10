package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Name string `mapstructure:"name"`
}

func TestNew(t *testing.T) {
	v := viper.New()
	v.Set("service.name", "alpha")

	r := New(v)
	cfg := &testConfig{}
	r.Unmarshal("service", cfg)

	assert.Equal(t, "alpha", cfg.Name)
}

func TestLoaderUnmarshal(t *testing.T) {
	v := viper.New()
	r := New(v)

	v.Set("reg.name", "beta")
	cfg := &testConfig{}
	r.Unmarshal("reg", cfg)

	assert.Equal(t, "beta", cfg.Name)
}

func TestLoaderUnmarshalMany(t *testing.T) {
	v := viper.New()
	v.Set("a.name", "one")
	v.Set("b.name", "two")

	r := New(v)

	cfgA := &testConfig{}
	cfgB := &testConfig{}
	r.UnmarshalMany(map[string]any{
		"a": cfgA,
		"b": cfgB,
	})

	assert.Equal(t, "one", cfgA.Name)
	assert.Equal(t, "two", cfgB.Name)
}
