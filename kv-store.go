package main

import (
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	val       string
	expiresAt int64 // unix milli, -1 for never expiring
}

type kvstore struct {
	mp map[string]Entry
	mu sync.RWMutex
}

// if value already exist, override it
func (kv *kvstore) SetValue(key, val string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	// 30 seconds from now
	expirationTime := time.Now().UnixMilli() + int64(30*1_000)
	kv.mp[key] = Entry{val, expirationTime}
}

func (kv *kvstore) GetValue(key string) string {
	kv.mu.RLock()
	defer kv.mu.Unlock()
	entry, ok := kv.mp[key]
	resp := entry.val
	if !ok {
		resp = "NULL"
	}
	return resp
}

func (kv *kvstore) DeleteKey(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.mp, key)
}

func (kv *kvstore) StartStoreCleaner() {
	ticker := time.NewTicker(CLEANER_FREQUENCY * time.Second)
	// datasetSize := len(kv.mp) / 5
	datasetSize := 1
	for range ticker.C {
		fmt.Printf("starting store cleaner\n")

		kv.mu.Lock()

		count := 0
		for k, entry := range kv.mp {
			if entry.expiresAt <= time.Now().UnixMilli() {
				delete(kv.mp, k)
			}
			count++
			if count > datasetSize {
				break
			}
		}

		kv.mu.Unlock()

		fmt.Printf("store cleaning complete, cleaned %d keys\n", count)
	}

}
