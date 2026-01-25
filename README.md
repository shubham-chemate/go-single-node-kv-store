# go-single-node-kv-store

- In the ./doc folder, I've written every small decisions and learnings. I've typed all by myself and not used LLM for even a single word
- Logs are much better (over printf) for concurrent applications since they are thread safe, and useful in analyze on server.
- We have graceful shutdown method, so until and unless every client connection is closed program won't terminate after receiving the SYSCALL from OS
- We have configurable number of concurrent connections, so how much goroutines will going to spawn has ceiling
- Redis RESP protocol is used to read the data from client, more details will be in ./doc folder but the key is we have our own validations and optimizations for reading method
- There is read deadline on each client, after read deadline timeout connection will be closed from server
- We have active and passive TTL support
- To avoid fast concurrent access to our map storage, we have 32 shards where shard for key is chosen on the basis of FNV hash value. FNV is fast and gives uniform distribution
- Initially the performance with lots of logs was too bad, we were hitting on an average 13K RPS for SET and 10K RPS for GET, the bottleneck was I was printing every single read command to the log (10K concurrent clients, each sending 100K requests)
- After removing the unecessary logs and optimizing logging for readability, was able to improve by huge margin, got SET & GET at 180K RPS (10K concurrent clients, sending 100K requests)

## Performance Testing

testing command:
```shell 
redis-benchmark -p 6379 -t set,get -c 10000 -n 1000000 -q
```
configuration:
```go
	READ_DEADLINE_TIME = 60
	MAX_CLIENT_CONN    = 12000
	MAX_KEY_VAL_SIZE   = 1000
	CLEANER_FREQUENCY  = 40

	Macbook M4 Pro, 24 GB Ram, VS Code devcontainer, Go Alpine Image
```

Our key value store
- SET: 139K RPS, p50=36.8 ms
- GET: 141K RPS, p50=36.3 ms

Redis
- SET: 181K RPS, p50=27.8 ms
- GET: 177K RPS, p50=28.1 ms

## Future Enhancements
- [ ] AOF, append only file and persistence to disk
- [ ] Additional logging to print current number of active clients
- [ ] make it distributed, reliable, available, fault-tolerant