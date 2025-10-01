package scheduler

import (
	"sync"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"time"
	"fmt"
	"reflect"
	"log"
	"runtime/debug"
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
		//gocron.NewTask(f, args...),
		gocron.NewTask(makeSafeTask(f, args...)), // <-- zero-arg safe wrapper
		//safeTask(f, args...),
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

func (self *Scheduler) AddWithDuration(name string, duration time.Duration, f any, args ...any) {
	// create a scheduler
	var err error
	if self.schedule == nil {
		self.schedule, err = gocron.NewScheduler(); if err != nil {
			panic(err)
		}
	}

	// add a job to the scheduler
	_, err = self.schedule.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(duration))),
		//gocron.NewTask(f, args...),
		gocron.NewTask(makeSafeTask(f, args...)), // <-- zero-arg safe wrapper
		//safeTask(f, args...),
		gocron.WithName(name),
		gocron.WithTags(name),
		gocron.WithEventListeners(
			gocron.AfterJobRuns(
				func(jobID uuid.UUID, jobName string) {
					self.schedule.RemoveJob(jobID)
				},
			),
			gocron.AfterJobRunsWithError(
				func(jobID uuid.UUID, jobName string, err error) {
					self.schedule.RemoveJob(jobID)
				},
			),
			gocron.BeforeJobRuns(
				func(jobID uuid.UUID, jobName string) {
				},
			),
		),
	); if err != nil { panic(err) }
}

func (self *Scheduler) RemoveByTags(tags ...string) {
	self.schedule.RemoveByTags(tags...)
}

func (self *Scheduler) Start() {
	if self.schedule == nil {
		panic("No Scheduler Jobs")
	}
	self.schedule.Start()
}

func (self *Scheduler) Monitor() {
	for _, j := range self.schedule.Jobs() {
		fmt.Printf("ID:%v, Name:%v, Tags:%v\n", j.ID(), j.Name(), j.Tags())
	}
}

func safeTask(f any, args ...any) gocron.Task {
	return gocron.NewTask(func(innerArgs ...any) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic in job: %v\n", r)
			}
		}()
		switch fn := f.(type) {
		case func():
			fn()
		case func(...any):
			fn(args...)
		default:
			panic("unsupported function type for task")
		}
	}, args...)
}

// makeSafeTask returns a zero-arg task (either func() or func() error) so
// NewTask never complains about mismatched parameters.
// It calls f with args via reflection, recovers panics, and (if applicable)
// returns the underlying error so gocron can trigger AfterJobRunsWithError.
func makeSafeTask(f any, args ...any) any {
	fnVal := reflect.ValueOf(f)
	if fnVal.Kind() != reflect.Func {
		panic("scheduler: makeSafeTask: f must be a function")
	}
	fnType := fnVal.Type()

	// Build reflect.Value slice for arguments
	in := make([]reflect.Value, len(args))
	for i, a := range args {
		in[i] = reflect.ValueOf(a)
	}

	// Helper to call and collect a trailing error (if any)
	callAndMaybeError := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("recovered panic: %v", r)
				log.Println("-> Error", err)
				log.Println("-> Stack: ", string(debug.Stack()))
			}
		}()
		outs := fnVal.Call(in)
		if fnType.NumOut() == 1 && fnType.Out(0) == reflect.TypeOf((*error)(nil)).Elem() {
			if !outs[0].IsNil() {
				err = outs[0].Interface().(error)
			}
		}
		return
	}

	// If the original function returns error, give gocron a func() error
	if fnType.NumOut() == 1 && fnType.Out(0) == reflect.TypeOf((*error)(nil)).Elem() {
		return func() (err error) { return callAndMaybeError() }
	}

	// Otherwise give gocron a plain func()
	return func() { _ = callAndMaybeError() }
}
