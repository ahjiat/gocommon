package pool

import (
	"reflect"
	"sync"
	"log"
	"runtime/debug"
)

// Pool limits how many goroutines run at once.
type Pool struct {
	sem chan struct{}
	wg  sync.WaitGroup
}

// New creates a pool with a max concurrency of max.
// If max <= 0, it defaults to 1 (sequential).
func New(max int) *Pool {
	if max <= 0 {
		max = 1
	}
	return &Pool{
		sem: make(chan struct{}, max),
	}
}

// Go schedules fn to run subject to the pool's concurrency limit.
func (p *Pool) Go(fn any, args ...any) {
	// Type-check that fn is actually a function
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic("Go expects a function")
	}

	p.wg.Add(1)
	p.sem <- struct{}{} // acquire a slot (blocks if full)

	go func() {
		defer func() {
			if errmsg := recover(); errmsg != nil {
				log.Println("-> Error", errmsg)
				log.Println("-> Stack: ", string(debug.Stack()))
			}
			<-p.sem // release slot
			p.wg.Done()
		}()
		
		// Convert args to reflect.Value
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			in[i] = reflect.ValueOf(arg)
		}

		// Call the function dynamically
		v.Call(in)
	}()
}

// Wait blocks until all scheduled functions complete.
func (p *Pool) Wait() {
	p.wg.Wait()
}
