package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

// DefaultShutdownTimeout is the maximum duration to wait for goroutines to finish.
const DefaultShutdownTimeout = 30 * time.Second

type SignalCallback struct {
	Signals      []os.Signal
	CallbackFunc func(ctx context.Context)
}

type App struct {
	wg                  sync.WaitGroup
	ctx                 context.Context
	cancel              context.CancelFunc
	startingCallbacks   []func()
	startedCallbacks    []func()
	signalCallbacks     []SignalCallback
	asyncCallbacks      []func()
	terminatedCallbacks []func()
}

func New() *App {
	return &App{
		wg: sync.WaitGroup{},
	}
}

// OnStarting adds callbacks to be executed before the app starts.
func (app *App) OnStarting(callbacks ...func()) *App {
	app.startingCallbacks = append(app.startingCallbacks, callbacks...)
	return app
}

// OnStarted adds callbacks to be executed after the app starts.
func (app *App) OnStarted(callbacks ...func()) *App {
	app.startedCallbacks = append(app.startedCallbacks, callbacks...)
	return app
}

// OnSignal adds callbacks to be executed when the specified signals are received.
func (app *App) OnSignal(signals []os.Signal, callbacks ...func(context.Context)) *App {
	for _, callback := range callbacks {
		app.signalCallbacks = append(app.signalCallbacks, SignalCallback{
			Signals:      signals,
			CallbackFunc: callback,
		})
	}

	return app
}

// OnAsync adds callbacks to be executed asynchronously.
func (app *App) OnAsync(callbacks ...func()) *App {
	app.asyncCallbacks = append(app.asyncCallbacks, callbacks...)
	return app
}

// OnTerminated adds callbacks to be executed after the app terminates.
func (app *App) OnTerminated(callbacks ...func()) *App {
	app.terminatedCallbacks = append(app.terminatedCallbacks, callbacks...)
	return app
}

func (app *App) runStartingCallbacks() {
	for _, startingCallback := range app.startingCallbacks {
		startingCallback()
	}
}

func (app *App) runStartedCallbacks() {
	for _, startedCallback := range app.startedCallbacks {
		startedCallback()
	}
}

func (app *App) runSignalCallBacks() {
	app.wg.Add(len(app.signalCallbacks))

	for _, signalCallback := range app.signalCallbacks {
		ctx, stop := signal.NotifyContext(app.ctx, signalCallback.Signals...)
		callback := signalCallback.CallbackFunc
		go func() {
			defer app.wg.Done()
			defer stop()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("recovered panic: %v\n", r)
				}
			}()
			<-ctx.Done()
			app.cancel()
			callback(ctx)
		}()
	}
}

func (app *App) runAsyncCallbacks() {
	for _, asyncCallback := range app.asyncCallbacks {
		callback := asyncCallback
		app.wg.Add(1)
		go func() {
			defer app.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("recovered panic: %v\n", r)
				}
			}()
			callback()
		}()
	}
}

func (app *App) runTerminatedCallbacks() {
	for _, terminatedCallback := range app.terminatedCallbacks {
		terminatedCallback()
	}
}

func (app *App) Run(executerFunc func()) {
	app.ctx, app.cancel = context.WithCancel(context.Background())
	defer app.cancel()

	app.runStartingCallbacks()
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer app.cancel()
		executerFunc()
	}()
	app.runStartedCallbacks()
	app.runSignalCallBacks()
	app.runAsyncCallbacks()

	done := make(chan struct{})
	go func() {
		app.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-app.ctx.Done():
	}

	select {
	case <-done:
	case <-time.After(DefaultShutdownTimeout):
	}

	app.runTerminatedCallbacks()
}
