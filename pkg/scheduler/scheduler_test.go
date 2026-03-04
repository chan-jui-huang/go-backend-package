package scheduler

import (
	"sync/atomic"
	"testing"
	"time"
)

type dummyJob struct {
	freq     string
	executed int32
}

func (j *dummyJob) GetFrequency() string { return j.freq }
func (j *dummyJob) Execute()             { atomic.AddInt32(&j.executed, 1) }

func TestBacklogJobsAndStart(t *testing.T) {
	s := NewScheduler(map[string]Job{})
	dj := &dummyJob{freq: "*/1 * * * * *"}
	s.BacklogJobs(map[string]Job{"job1": dj})
	if _, ok := s.backlogJobs["job1"]; !ok {
		t.Fatalf("backlog not set")
	}

	// Start should add backlog jobs to crontab and clear backlog
	s.Start()
	defer s.Stop()
	if len(s.backlogJobs) != 0 {
		t.Fatalf("backlog not cleared after Start")
	}
	if _, ok := s.jobs["job1"]; !ok {
		t.Fatalf("job not added on Start")
	}
}

func TestAddAndRemoveJob(t *testing.T) {
	s := NewScheduler(map[string]Job{})
	dj := &dummyJob{freq: "*/1 * * * * *"}
	if err := s.AddJob("j1", dj); err != nil {
		t.Fatalf("AddJob error: %v", err)
	}
	if _, ok := s.jobs["j1"]; !ok {
		t.Fatalf("job not present after AddJob")
	}

	s.RemoveJob("j1")
	if _, ok := s.jobs["j1"]; ok {
		t.Fatalf("job still present after RemoveJob")
	}
}

func TestStopReturnsContext(t *testing.T) {
	s := NewScheduler(map[string]Job{})
	dj := &dummyJob{freq: "*/1 * * * * *"}
	if err := s.AddJob("j2", dj); err != nil {
		t.Fatalf("AddJob error: %v", err)
	}
	// start the underlying cron to ensure Stop behaves
	s.crontab.Start()
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
