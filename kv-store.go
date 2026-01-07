package main

import "sync"

type kvstore struct {
	mp map[string]string
	mu sync.RWMutex
}

// if value already exist, override it
func (kv *kvstore) SetValue(key, val string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.mp[key] = val
}

func (kv *kvstore) GetValue(key string) string {
	kv.mu.RLock()
	defer kv.mu.Unlock()
	val, ok := kv.mp[key]
	if !ok {
		val = "NULL"
	}
	return val
}

func (kv *kvstore) DeleteKey(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.mp, key)
}
