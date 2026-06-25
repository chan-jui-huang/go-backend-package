package scheduler_test

import (
	"sync/atomic"
	"testing"
	"time"

	scheduler "github.com/chan-jui-huang/go-backend-package/v2/pkg/scheduler"
)

type dummyJob struct {
	name           string
	cronExpression string
	executed       int32
}

func (j *dummyJob) Name() string           { return j.name }
func (j *dummyJob) CronExpression() string { return j.cronExpression }
func (j *dummyJob) Execute()               { atomic.AddInt32(&j.executed, 1) }

func TestQueueJobsAndStart(t *testing.T) {
	s := scheduler.NewScheduler(nil)
	dj := &dummyJob{name: "job1", cronExpression: "*/1 * * * * *"}
	s.QueueJobs([]scheduler.Job{dj})
	s.Start()
	defer s.Stop()

	deadline := time.Now().Add(1500 * time.Millisecond)
	for atomic.LoadInt32(&dj.executed) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("expected backlog job to run after Start")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestScheduleAndUnscheduleJob(t *testing.T) {
	s := scheduler.NewScheduler(nil)
	dj := &dummyJob{name: "j1", cronExpression: "*/1 * * * * *"}
	if err := s.ScheduleJob(dj); err != nil {
		t.Fatalf("ScheduleJob error: %v", err)
	}

	s.Start()
	deadline := time.Now().Add(1500 * time.Millisecond)
	for atomic.LoadInt32(&dj.executed) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("expected added job to run")
		}
		time.Sleep(10 * time.Millisecond)
	}

	executedBeforeRemove := atomic.LoadInt32(&dj.executed)
	s.UnscheduleJob("j1")
	time.Sleep(1100 * time.Millisecond)
	executedAfterRemove := atomic.LoadInt32(&dj.executed)
	s.Stop()

	if executedAfterRemove != executedBeforeRemove {
		t.Fatalf("expected removed job not to run again, got before=%d after=%d", executedBeforeRemove, executedAfterRemove)
	}
}

func TestStopReturnsContext(t *testing.T) {
	s := scheduler.NewScheduler(nil)
	dj := &dummyJob{name: "j2", cronExpression: "*/1 * * * * *"}
	s.QueueJobs([]scheduler.Job{dj})
	s.Start()
	ctx := s.Stop()
	if ctx == nil {
		t.Fatalf("Stop returned nil context")
	}
	// ensure Stop returned a context we can select on without blocking indefinitely
	select {
	case <-ctx.Done():
		// ok
	case <-time.After(200 * time.Millisecond):
		// not done yet, but Stop returned successfully
	}
}
