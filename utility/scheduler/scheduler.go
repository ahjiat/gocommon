package scheduler

import (
	"sync"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

func New() Scheduler {
	return Scheduler{Jobs:map[string]*Job{}}
}

type Job struct {
	Job gocron.Job
	Name string
	IsRunning bool
	CompletedCount int
	CompletedCountError int
}

type Scheduler struct {
	Jobs map[string]*Job
	InRunningCount int
	mu sync.Mutex
	schedule gocron.Scheduler
}

func (self *Scheduler) Add(name string, cron string, f any, args ...any) {
	// create a scheduler
	var err error
	if self.schedule == nil {
		self.schedule, err = gocron.NewScheduler(); if err != nil {
			panic(err)
		}
	}

	var myjob *Job

	// add a job to the scheduler
	job, err := self.schedule.NewJob(
		gocron.CronJob(cron, false),
		gocron.NewTask(f, args...),
		gocron.WithName(name),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(
				func(jobID uuid.UUID, jobName string) {
					myjob.IsRunning = false
					myjob.CompletedCount++
					myjob.CompletedCountError++
				},
			),
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					myjob.IsRunning = false
					myjob.CompletedCountError++
				},
			),
			gocron.BeforeJobRuns(
				func(jobID uuid.UUID, jobName string) {
					myjob.IsRunning = true
				},
			),
		),
	); if err != nil { panic(err) }

	myjob = &Job{
		Job: job,
		Name: name,
	}
	self.Jobs[job.ID().String()] = myjob
}

func (self *Scheduler) Start() {
	if self.schedule == nil {
		panic("No Scheduler Jobs")
	}
	self.schedule.Start()
}


