# go-single-node-kv-store

Benchmarking
- High Reads vs High Writes
- Multiple Clients (100K), each focused on huge reads/writes, 20% writes, 80% reads

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
- [ ] write doc
    - should include parsing protocol
    - should include reader strategies, types of reader, byte reader, bulk reading
    - should include concurrent map
    - should include TTL
    - should include logging vs printing : may drop this doc, since it's pretty obvious to have logs instead of printf
- [ ] remove fmt.Printf, adder logger
- [ ] multiple maps for better concurrency
- [ ] utilize size of string to have bulk reading in client reading
- [ ] add log / goroutine for current active clients, current cache size (printing after every xy seconds)
- [ ] append only file / persistence


benchmarks

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

### v2 : added logger, slog
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