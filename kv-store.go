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
	expirationTime := time.Now().UnixMilli() + int64(40*1_000)
	fmt.Println("set expiration time to:", expirationTime)
	kv.mp[key] = Entry{val, expirationTime}
}

func (kv *kvstore) GetValue(key string) string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	entry, ok := kv.mp[key]
	if !ok {
		return ""
	}
	if entry.expiresAt != -1 && entry.expiresAt <= time.Now().UnixMilli() {
		return ""
	}
	return entry.val
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
		fmt.Printf("starting store cleaner, time: %d\n", time.Now().UnixMilli())

		kv.mu.Lock()

		checked := 0
		deleted := 0
		for k, entry := range kv.mp {
			if entry.expiresAt <= time.Now().UnixMilli() {
				delete(kv.mp, k)
				deleted++
			}
			checked++
			if checked > datasetSize {
				break
			}
		}

		kv.mu.Unlock()

		fmt.Printf("store cleaning complete, checked %d & cleaned %d keys, time: %d\n", checked, deleted, time.Now().UnixMilli())
	}

}
