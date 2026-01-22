package service

import (
	"reflect"
	"sync"
)

type registry struct {
	services map[string]any
	mu       sync.RWMutex
}

var Registry *registry

func init() {
	Registry = &registry{
		services: map[string]any{},
	}
}

func NewRegistry() *registry {
	return &registry{
		services: map[string]any{},
	}
}

func (r *registry) Set(key string, service any) {
	if !(reflect.ValueOf(service).Kind() == reflect.Pointer || reflect.ValueOf(service).Kind() == reflect.Func) {
		panic("service is not the pointer or function")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[key] = service
}

func (r *registry) SetMany(services map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, service := range services {
		if !(reflect.ValueOf(service).Kind() == reflect.Pointer || reflect.ValueOf(service).Kind() == reflect.Func) {
			panic("service is not the pointer or function")
		}
		r.services[key] = service
	}
}

func (r *registry) Get(key string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v := reflect.ValueOf(r.services[key])

	return v.Interface()
}

func (r *registry) Unset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, key)
}

func (r *registry) UnsetMany(keys ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, key := range keys {
		delete(r.services, key)
	}
}

func (r *registry) Clone(key string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v := reflect.ValueOf(r.services[key])

	return v.Elem().Interface()
}
