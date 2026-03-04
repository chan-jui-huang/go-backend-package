package service

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testService struct {
	Name string
}

func TestRegistry_SetGetCloneUnset(t *testing.T) {
	r := NewRegistry()

	svc := &testService{Name: "alpha"}
	r.Set("svc", svc)

	got := r.Get("svc").(*testService)
	assert.Same(t, svc, got)

	clone := r.Clone("svc").(testService)
	assert.Equal(t, *svc, clone)

	r.Unset("svc")
	r.Set("a", svc)
	r.Set("b", &testService{Name: "beta"})
	r.UnsetMany("a", "b")

	r.Set("a", svc)
	assert.NotNil(t, r.Get("a"))
}

func TestRegistry_ConcurrentSetGet(t *testing.T) {
	r := NewRegistry()

	for i := 0; i < 100; i++ {
		r.Set(strconv.Itoa(i), &testService{Name: strconv.Itoa(i)})
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Get(strconv.Itoa(i)).(*testService)
		}()
	}

	for i := 100; i < 200; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Set(strconv.Itoa(i), &testService{Name: strconv.Itoa(i)})
		}()
	}

	wg.Wait()
}
