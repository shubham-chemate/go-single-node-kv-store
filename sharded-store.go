package main

import "hash/fnv"

type ShardedStore struct {
	shards []*kvstore
	count  int
}

func GetNewShardedStore(shardsCount int) *ShardedStore {
	store := &ShardedStore{
		shards: make([]*kvstore, shardsCount),
		count:  shardsCount,
	}
	for i := range shardsCount {
		store.shards[i] = &kvstore{mp: make(map[string]Entry)}
		go store.shards[i].StartStoreCleaner()
	}
	return store
}

func (store *ShardedStore) SetValue(key, val string, ttlInSecond int64) {
	h := getKeyHash(key)
	store.shards[h].SetValue(key, val, ttlInSecond)
}

func (kv *ShardedStore) GetValue(key string) string {
	h := getKeyHash(key)
	return store.shards[h].GetValue(key)
}

func (kv *ShardedStore) DeleteKey(key string) {
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
