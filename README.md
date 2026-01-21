# go-single-node-kv-store

- In the ./doc folder, I've written every small decisions and learnings. I've typed all by myself and not used LLM for even a single word
- Logs are much better for concurrent applications since they are thread safe, just mentioning.
- Initially the performance with log of logs was too bad, we were hitting on an average 13K RPS for SET and 10K RPS for GET, the bottleneck was I was printing every single read command to the log (10K concurrent clients, each sending 100K requests)
- After removing the unecessary logs and optimizing logging for readability, was able to improve by huge margin, got SET & GET at 180K RPS (10K concurrent clients, each sending 100K requests)

### Future Enhancements
- AOF, append only file and persistence to disk
- make it distributed, reliable, available, fault-tolerant

upcoming
- [x] write server doc
- [x] limit reader to read only defined number of characters
- [x] validate command
- [x] process command using concurrent map
- [x] benchmarking way
- [x] to RESP POC, it will help in benchmarking and other things
- [x] TTL for set command
    - fix on TTL strategies
        - automatic eviction
        - eviction on query
    - add TTL in RESP
- [x] remove fmt.Printf, adder logger
- [ ] write doc
    - should include parsing protocol
    - should include reader strategies, types of reader, byte reader, bulk reading
    - hashing that is used to select from multiple maps
- [x] multiple maps for better concurrency
- [x] utilize size of string to have bulk reading in client reading
- [ ] add log / goroutine for current active clients, current cache size (printing after every xy seconds)
- [ ] append only file / persistence


## Benchmarks

#### v1 : baseline
command: redis-benchmark -p 6379 -t set,get -c 10000 -n 100000 -q
```go
	READ_DEADLINE_TIME = 60
	MAX_CLIENT_CONN    = 12000
	MAX_KEY_VAL_SIZE   = 1000
	CLEANER_FREQUENCY  = 40
```
iteration1:
SET: 10278.55 requests per second, p50=776.191 msec
GET: 15121.73 requests per second, p50=410.879 msec
iteration2:
SET: 13007.28 requests per second, p50=759.807 msec
GET: 18409.43 requests per second, p50=389.119 msec
iteration3:
SET: 10722.71 requests per second, p50=797.183 msec
GET: 18315.02 requests per second, p50=392.959 msec

#### v2 : added logger, slog
command: redis-benchmark -p 6379 -t set,get -c 10000 -n 100000 -q
```go
	READ_DEADLINE_TIME = 60
	MAX_CLIENT_CONN    = 12000
	MAX_KEY_VAL_SIZE   = 1000
	CLEANER_FREQUENCY  = 40
```
iteration1:
SET: 5088.02 requests per second, p50=1958.911 msec                     
GET: 7352.94 requests per second, p50=1078.271 msec
iteration2:
SET: 4073.15 requests per second, p50=2138.111 msec                      
GET: 5868.89 requests per second, p50=1265.663 msec 
iteration3:
SET: 4019.94 requests per second, p50=2125.823 msec                      
GET: 7231.70 requests per second, p50=1096.703 msec

improved a bit on logging (removed HOT PATH logs)
iteration1:
SET: 183150.19 requests per second, p50=25.951 msec                     
GET: 187265.92 requests per second, p50=25.615 msec
iteration2:
SET: 178890.88 requests per second, p50=26.191 msec                     
GET: 190114.06 requests per second, p50=25.231 msec                     
iteration3:
SET: 176366.86 requests per second, p50=26.463 msec                     
GET: 175746.92 requests per second, p50=26.815 msec

#### v3 : read optimizations
command: redis-benchmark -p 6379 -t set,get -c 10000 -n 100000 -q
iteration1:
SET: 179211.45 requests per second, p50=26.063 msec                     
GET: 187265.92 requests per second, p50=25.103 msec
iteration2:
SET: 180831.83 requests per second, p50=25.807 msec                     
GET: 187969.92 requests per second, p50=25.727 msec
iteration3:
SET: 179211.45 requests per second, p50=26.255 msec
GET: 180831.83 requests per second, p50=26.207 msec

#### v4 : multiple maps, hashing
command: redis-benchmark -p 6379 -t set,get -c 10000 -n 100000 -q
iteration1:
SET: 178890.88 requests per second, p50=26.079 msec                     
GET: 178253.12 requests per second, p50=26.719 msec  
iteration2:
SET: 184501.84 requests per second, p50=25.695 msec
GET: 186219.73 requests per second, p50=25.647 msec
iteration3:
SET: 180505.41 requests per second, p50=25.967 msec
GET: 181488.20 requests per second, p50=26.111 msec

command: redis-benchmark -p 6379 -t set,get -c 50 -n 100000 -q
SET: 238095.25 requests per second, p50=0.127 msec
GET: 272479.56 requests per second, p50=0.111 msec
here we can see that there is immense improvement in the p50 latency