# go-single-node-kv-store

Benchmarking
- High Reads vs High Writes
- Multiple Clients (100K), each focused on huge reads/writes, 20% writes, 80% reads

upcoming
- [x] write server doc
- [x] limit reader to read only defined number of characters
- [x] validate command
- [x] process command using concurrent map
- [ ] write parser doc
    - should include parsing protocol
    - should include reader
- [ ] add log / goroutine for current active clients (printing after every xy seconds)
- [ ] remove fmt.Printf, adder logger
- [ ] TTL for set command
    - fix on TTL strategies
        - automatic eviction
        - eviction on query
    - add TTL in RESP
- [ ] benchmarking way
- [ ] multiple maps for better concurrency
- [ ] append only file / persistence
- [ ] to RESP POC, it will help in benchmarking and other things
- [ ] utilize size of string to have bulk reading in client reading