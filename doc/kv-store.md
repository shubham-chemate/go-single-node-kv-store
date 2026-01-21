## Key Value Store

- our KV store must support concurrent operations
- we are using go's internal map
- when we are adding key to the map, we will lock it completely, so that no-one can read / write to the map
- when we are reading from a map, we will use shared read lock
- our RWMutex lock allows single write at a time but multiple reads at a time, when we are writing no-one is allowed to read/write, when we are reading others are allowed to read
- when deleting, we are using exclusive lock again
- we have added store cleaner, which runs in configured periodic time, the goal is to clear the long lived keys from the KV store, we are having TTL for each key, if key is expired we are removing it from the map
- some keys are stored with infinite time

### TTL Methodology
- in our KV store, along with value of key, we are storing the expiration time in unix milliseconds
- we are making sure that we are not returning expired keys in two ways
    - when user asks for key, we explicitly checks weather it is expired or not
    - we are also running background cleaner so that after configured time we are cleaning the expired keys


### Shareded KV Store
- to reduce the lock contension on single store, we have introduced shards
- we created shared store which is basically a collection of regular KV stores
- the methods are simple, for each method we will going to first create hash of the input key to determin the shard and then we will make that shard to handle that particular input key
- we have used fnv hash function provided by golang standard library, as per my research it is fast and distribute keys evenly, we really don't need cryptographic hash functions as they are computationaly heavy

![Alt text for the image](./sharded-kv-store.png)