package main

import (
	"log/slog"
	"sync"
	"time"
)

type Entry struct {
	val       string
	expiresAt int64 // unix milli, -1 for never expiring
}

type KVStore struct {
	mp map[string]Entry
	mu sync.RWMutex
}

// if value already exist, override it
func (kv *KVStore) SetValue(key, val string, ttlInSecond int64) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	// ttl seconds from now
	expirationTime := time.Now().UnixMilli() + ttlInSecond*1_000
	if ttlInSecond == -1 {
		expirationTime = -1
	}
	// slog.Info("adding key to map", "key", key, "val", val, "expiration time", expirationTime)
	kv.mp[key] = Entry{val, expirationTime}
}

func (kv *KVStore) GetValue(key string) string {
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

func (kv *KVStore) DeleteKey(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.mp, key)
}

// in one iteration we only check 20% random keys from map
func (kv *KVStore) StartStoreCleaner() {
	ticker := time.NewTicker(CLEANER_FREQUENCY * time.Second)
	datasetSize := len(kv.mp) / 5
	for range ticker.C {
		slog.Info("starting store cleaner", "unix milli now", time.Now().UnixMilli())

		kv.mu.Lock()

		checked := 0
		deleted := 0
		for k, entry := range kv.mp {
			if entry.expiresAt != -1 && entry.expiresAt <= time.Now().UnixMilli() {
				delete(kv.mp, k)
				deleted++
			}
			checked++
			if checked > datasetSize {
				break
			}
		}

		kv.mu.Unlock()

		slog.Info("store cleaning complete", "checked", checked, "cleaned", deleted, "unix millis now", time.Now().UnixMilli())
	}

}
