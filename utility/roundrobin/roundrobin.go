package roundrobin

import (
	"fmt"
	"sync"
)

type RoundRobin struct {
	idx int
	values []interface{}
	mux sync.Mutex
}

var container = map[string]*RoundRobin{}

func Create(name string, values []interface{} ) {
	container[name] = &RoundRobin{idx:0, values:values}
}

func Get(name string) interface{} {
	v, found := container[name]
	if ! found { panic(fmt.Sprintf("roundrobin name:[%s] not initialize", name)) }
	v.mux.Lock()
	i := v.values[v.idx]
	v.idx++
	if v.idx == len(v.values) { v.idx = 0 }
	v.mux.Unlock()
	return i
}
