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

#### v1
command: redis-benchmark -p 6379 -t set,get -c 100000 -n 100000 -q
output1:
output2:
output3:
