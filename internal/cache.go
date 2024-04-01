package cacheserver

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
)

type Cache struct {
	mp          sync.Map
	commitIndex int
	LogFile     string
	LogEntry    []string //"SET KEY VAL" or "DEL KEY"
}

func (cache *Cache) getData(key any) string {
	value, ok := cache.mp.Load(key)
	if !ok {
		return ""
	}
	return value.(string) + "\r\n"
}

func (cache *Cache) storeData(key, value any) string {
	cache.LogEntry = append(cache.LogEntry, fmt.Sprintf("SET %v %v\r\n", key, value))
	// wal.AppendLog(cache.LogFile, fmt.Sprintf("SET %v %v\r\n", key, value))
	cache.commitIndex += 1
	cache.mp.Store(key, value)
	fmt.Println(cache.commitIndex)
	return "OK\r\n"
}

func (cache *Cache) deleteData(key any) string {
	value, loaded := cache.mp.LoadAndDelete(key)
	if !loaded {
		return ""
	}

	// wal.AppendLog(cache.LogFile, fmt.Sprintf("DEL %v\r\n", key))

	cache.LogEntry = append(cache.LogEntry, fmt.Sprintf("DEL %v\r\n", key))
	cache.commitIndex += 1

	fmt.Println(cache.commitIndex)
	return value.(string) + "\r\n"
}

func (cache *Cache) appendEntries(datab64 string) string {
	data, _ := base64.StdEncoding.DecodeString(datab64)

	splits := strings.Split(string(data), ",")

	for _, split := range splits {
		cmd := strings.Fields(split)
		if len(cmd) > 0 {
			if cmd[0] == "SET" {
				cache.storeData(cmd[1], cmd[2])
			} else if cmd[0] == "DEL" {
				cache.deleteData(cmd[1])
			}
		}
	}

	return "ok"
}
