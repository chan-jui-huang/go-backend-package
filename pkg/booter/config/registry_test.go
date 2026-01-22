package config

import (
	"strconv"
	"sync"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Name string `mapstructure:"name"`
}

func TestRegistrySetGetRegisterUnset(t *testing.T) {
	v := viper.New()
	r := NewRegistry(v)

	cfg := &testConfig{Name: "alpha"}
	r.Set("cfg", cfg)

	got := r.Get("cfg").(testConfig)
	assert.Equal(t, *cfg, got)

	v.Set("reg.name", "beta")
	regCfg := &testConfig{}
	r.Register("reg", regCfg)
	registered := r.Get("reg").(testConfig)
	assert.Equal(t, "beta", registered.Name)

	r.Unset("cfg")
	r.Set("a", &testConfig{Name: "one"})
	r.Set("b", &testConfig{Name: "two"})
	r.UnsetMany("a", "b")

	r.Set("a", &testConfig{Name: "one"})
	assert.NotEmpty(t, r.Get("a"))
}

func TestRegistryConcurrentSetGet(t *testing.T) {
	r := NewRegistry(viper.New())

	for i := 0; i < 100; i++ {
		r.Set(strconv.Itoa(i), &testConfig{Name: strconv.Itoa(i)})
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Get(strconv.Itoa(i)).(testConfig)
		}()
	}

	for i := 100; i < 200; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Set(strconv.Itoa(i), &testConfig{Name: strconv.Itoa(i)})
		}()
	}

	wg.Wait()
}
