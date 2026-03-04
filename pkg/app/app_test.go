package app

import (
	"context"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestAppBasicLifecycle(t *testing.T) {
	app := New()

	var started int32
	var starting int32
	var terminated int32

	app.OnStarting(func() { atomic.StoreInt32(&starting, 1) })
	app.OnStarted(func() { atomic.StoreInt32(&started, 1) })
	app.OnTerminated(func() { atomic.StoreInt32(&terminated, 1) })

	// executor just returns immediately
	app.Run(func() {})

	if atomic.LoadInt32(&starting) != 1 {
		t.Fatalf("starting callbacks not executed")
	}
	if atomic.LoadInt32(&started) != 1 {
		t.Fatalf("started callbacks not executed")
	}
	if atomic.LoadInt32(&terminated) != 1 {
		t.Fatalf("terminated callbacks not executed")
	}
}

func TestAppAsyncCallbacksAndRecovery(t *testing.T) {
	app := New()

	var asyncRan int32
	app.OnAsync(func() { atomic.StoreInt32(&asyncRan, 1) })
	// also add a panicking async callback to ensure recovery doesn't crash
	app.OnAsync(func() { panic("boom") })

	app.Run(func() {})

	if atomic.LoadInt32(&asyncRan) != 1 {
		t.Fatalf("async callback did not run")
	}
}

func TestAppSignalCallbacks(t *testing.T) {
	app := New()

	var sigRan int32
	// use SIGUSR1 for test; callback will set flag
	app.OnSignal([]os.Signal{syscall.SIGUSR1}, func(ctx context.Context) { atomic.StoreInt32(&sigRan, 1) })

	// executor will send the signal shortly after start
	app.Run(func() {
		// give Run some time to set up signal handlers
		time.Sleep(10 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
		// wait a moment for handler to run and cause cancellation
		time.Sleep(50 * time.Millisecond)
	})

	if atomic.LoadInt32(&sigRan) != 1 {
		t.Fatalf("signal callback did not run")
	}
}

func TestAppSignalCallbackWithEmptySignalsCancelsOnParent(t *testing.T) {
	app := New()
	var ran int32
	// empty signals slice: NotifyContext will cancel on parent cancel as well
	app.OnSignal([]os.Signal{}, func(ctx context.Context) { atomic.StoreInt32(&ran, 1) })

	// We need access to app.cancel to trigger parent cancel; tests are in same package so we can access unexported fields.
	app.Run(func() {
		// give Run setup time
		time.Sleep(10 * time.Millisecond)
		// cancel parent context to simulate signal
		if app.cancel != nil {
			app.cancel()
		}
		// allow callbacks to run
		time.Sleep(20 * time.Millisecond)
	})

	if atomic.LoadInt32(&ran) != 1 {
		t.Fatalf("signal callback did not run on parent cancel")
	}
}
