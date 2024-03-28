package cacheserver

import (
	"fmt"
	"sync"

	wal "cached.0jk6.github.io/pkg"
)

type Cache struct {
	mp      sync.Map
	LogFile string
}

func (cache *Cache) getData(key any) string {
	value, ok := cache.mp.Load(key)

	if !ok {
		return ""
	}

	return value.(string) + "\r\n"
}

func (cache *Cache) storeData(key, value any) string {
	wal.AppendLog(cache.LogFile, fmt.Sprintf("SET %v %v\r\n", key, value))
	cache.mp.Store(key, value)
	return "OK\r\n"
}

func (cache *Cache) deleteData(key any) string {
	value, loaded := cache.mp.LoadAndDelete(key)
	if !loaded {
		return ""
	}

	wal.AppendLog(cache.LogFile, fmt.Sprintf("DEL %v\r\n", key))
	return value.(string) + "\r\n"
}
