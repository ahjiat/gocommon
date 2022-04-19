package parallel

import (
	"fmt"
	"sync"
)

func Create(name string, maxThreadNum int) {
	if maxThreadNum == 0 { maxThreadNum = 1 }
	container[name] = &Parallel{
		maxThreadNum: maxThreadNum,
		channel: make(chan bool, maxThreadNum),
	}
}

func Get(name string) *Parallel {
	v, found := container[name]
	if ! found { panic(fmt.Sprintf("parallel name:[%s] not initialize", name)) }
	return v
}

type Parallel struct {
	maxThreadNum int
	channel chan bool
	wg sync.WaitGroup
}

func (self *Parallel) ParallelRunOrWait(f func(...interface{}), args ...interface{}) {
	self.wg.Add(1)
	self.channel <- true
	go func() {
		defer func() {
			<-self.channel
			self.wg.Done()
		}()
		f(args...)
	}()
}

func (self *Parallel) Wait() {
	self.wg.Wait()
}

var container = map[string]*Parallel{}
