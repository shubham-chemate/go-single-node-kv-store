package main

import "hash/fnv"

type ShardedKVStore struct {
	shards []*KVStore
	count  int
}

func GetNewShardedKVStore(shardsCount int) *ShardedKVStore {
	store := &ShardedKVStore{
		shards: make([]*KVStore, shardsCount),
		count:  shardsCount,
	}
	for i := range shardsCount {
		store.shards[i] = &KVStore{mp: make(map[string]Entry)}
		go store.shards[i].StartStoreCleaner()
	}
	return store
}

func (store *ShardedKVStore) SetValue(key, val string, ttlInSecond int64) {
	h := getKeyHash(key)
	store.shards[h].SetValue(key, val, ttlInSecond)
}

func (kv *ShardedKVStore) GetValue(key string) string {
	h := getKeyHash(key)
	return store.shards[h].GetValue(key)
}

func (kv *ShardedKVStore) DeleteKey(key string) {
	h := getKeyHash(key)
	store.shards[h].DeleteKey(key)
}

/*
FNV is
- non cryptographic
- fast
- uniform distribution
*/
func getKeyHash(key string) int64 {
	hs := fnv.New32a()
	hs.Write([]byte(key))
	return int64(hs.Sum32()) % NUMBER_OF_SHARDS
}
