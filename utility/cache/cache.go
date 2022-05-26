package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
	"fmt"
	"common/utility/method"
	"sync"
)

var mutexCtrls = map[string]*sync.Mutex{}

var localCache *cache.Cache

const (
	NoExpiration time.Duration = cache.NoExpiration
)

func init() {
	localCache = cache.New(cache.NoExpiration, 5 * time.Minute)
}

func Set(key string, value interface{}, t time.Duration) {
	localCache.Set(key, value, t)
}

func Get[T any](key string) (T, bool) {
	if value, found := localCache.Get(key); found {
		return value.(T), true
	}
	return *new(T), false
}

func Delete(key string) {
	localCache.Delete(key)
}

func Load[T any](f func() (T,bool), expire time.Duration, suffix interface{}) T {
	methodName := method.AppendName(suffix, 3)
	value, found := Get[T](methodName); if found { return value }
	value, isSave := f()
	if isSave { Set(methodName, value, expire) }
	return value
}

func LoadSync[T any](f func() (T,bool), expire time.Duration, suffix interface{}) T {
	methodName := method.AppendName(suffix, 3)
	value, found := Get[T](methodName); if found { return value }

	/*
		max 1 goroutine bypass
	*/
	var mux *sync.Mutex
	mux, found = mutexCtrls[methodName]; if ! found {
		mux = &sync.Mutex{}
		mutexCtrls[methodName] = mux
	}
	mux.Lock()
	defer func() {
		mux.Unlock()
		delete(mutexCtrls, methodName)
	}()

	value, found = Get[T](methodName); if found { return value }
	value, isSave := f()
	if isSave { Set(methodName, value, expire) }

	return value
}

func DisplayAllCaches() {
	list := localCache.Items()
	for key, value := range list {
		fmt.Printf("%v -> %T\n", key, value.Object)
	}
}


