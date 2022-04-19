package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
	"fmt"
	"common/utility/method"
)

var localCache *cache.Cache

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

func Load[T any](f func() T, expire time.Duration, suffix interface{}) T {
	methodName := method.AppendName(suffix, 3)
	value, found := Get[T](methodName); if found { return value }
	value = f()
	Set(methodName, value, expire)
	return value
}

func DisplayAllCaches() {
	list := localCache.Items()
	for key, value := range list {
		fmt.Printf("%v -> %T\n", key, value.Object)
	}
}
