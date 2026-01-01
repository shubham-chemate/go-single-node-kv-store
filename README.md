# go-single-node-kv-store

Benchmarking
- High Reads vs High Writes
- Multiple Clients (100K), each focused on huge reads/writes, 20% writes, 80% reads

Next:
- graceful shutdown
- io in go
- limit go routines / TCP connection
    - create pool of collection handlers / workers
    - assigne worker from pool to new connection
- slowries attack
    - client send 1 byte every 99 second / threshhold of timeout
- if one go routine panics, don't let entire program to crash
    - need panic recovery mechanism