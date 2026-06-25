package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
)

type Job interface {
	Name() string
	CronExpression() string
	Execute()
}

type scheduler struct {
	crontab    *cron.Cron
	queuedJobs map[string]Job
	jobs       map[string]cron.EntryID
}

var Scheduler *scheduler

func init() {
	Scheduler = NewScheduler(nil)
}

func NewScheduler(jobs []Job) *scheduler {
	return &scheduler{
		crontab:    cron.New(cron.WithSeconds()),
		queuedJobs: newQueuedJobs(jobs),
		jobs:       map[string]cron.EntryID{},
	}
}

func newQueuedJobs(jobs []Job) map[string]Job {
	queuedJobs := map[string]Job{}
	for _, job := range jobs {
		queuedJobs[job.Name()] = job
	}

	return queuedJobs
}

func (s *scheduler) QueueJobs(jobs []Job) {
	for _, job := range jobs {
		s.queuedJobs[job.Name()] = job
	}
}

func (s *scheduler) RemoveQueuedJobs(names []string) {
	for _, name := range names {
		delete(s.queuedJobs, name)
	}
}

func (s *scheduler) ClearQueuedJobs() {
	s.queuedJobs = map[string]Job{}
}

func (s *scheduler) ScheduleJob(job Job) error {
	id, err := s.crontab.AddFunc(job.CronExpression(), job.Execute)
	if err != nil {
		return err
	}
	s.jobs[job.Name()] = id

	return nil
}

func (s *scheduler) UnscheduleJob(name string) {
	s.crontab.Remove(s.jobs[name])
	delete(s.jobs, name)
}

func (s *scheduler) Start() {
	for _, job := range s.queuedJobs {
		if err := s.ScheduleJob(job); err != nil {
			panic(err)
		}
	}
	s.queuedJobs = map[string]Job{}
	s.crontab.Start()
}

func (s *scheduler) Stop() context.Context {
	return s.crontab.Stop()
}
