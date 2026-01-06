# go-single-node-kv-store

Benchmarking
- High Reads vs High Writes
- Multiple Clients (100K), each focused on huge reads/writes, 20% writes, 80% reads

upcoming
- write server doc
- write parser doc
- limit reader to read only defined number of characters
- validate command
- process command using concurrent map
- add log for current active clients (printing after every xy seconds)
- remove fmt.Printf, adder logger
- TTL for set command
    - fix on TTL strategies
        - automatic eviction
        - eviction on query