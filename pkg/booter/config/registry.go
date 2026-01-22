package config

import (
	"reflect"
	"sync"

	"github.com/spf13/viper"
)

type registry struct {
	viper   *viper.Viper
	configs map[string]any
	mu      sync.RWMutex
}

var Registry *registry

func init() {
	Registry = &registry{
		configs: map[string]any{},
	}
}

func NewRegistry(v *viper.Viper) *registry {
	return &registry{
		viper:   v,
		configs: map[string]any{},
	}
}

func (r *registry) Set(key string, config any) {
	if !(reflect.ValueOf(config).Kind() == reflect.Pointer) {
		panic("config is not the pointer")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[key] = config
}

func (r *registry) SetMany(configs map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, config := range configs {
		if !(reflect.ValueOf(config).Kind() == reflect.Pointer) {
			panic("config is not the pointer")
		}
		r.configs[key] = config
	}
}

func (r *registry) Register(key string, config any) {
	err := r.viper.UnmarshalKey(key, config)
	if err != nil {
		panic(err)
	}
	r.Set(key, config)
}

func (r *registry) RegisterMany(configs map[string]any) {
	for key, config := range configs {
		r.Register(key, config)
	}
}

func (r *registry) Get(key string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v := reflect.ValueOf(r.configs[key])

	return v.Elem().Interface()
}

func (r *registry) Unset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.configs, key)
}

func (r *registry) UnsetMany(keys ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, key := range keys {
		delete(r.configs, key)
	}
}

func (r *registry) SetViper(v *viper.Viper) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.viper = v
}

func (r *registry) GetViper() viper.Viper {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return *r.viper
}
