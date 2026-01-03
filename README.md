# go-single-node-kv-store

Benchmarking
- High Reads vs High Writes
- Multiple Clients (100K), each focused on huge reads/writes, 20% writes, 80% reads

Next:
- io in go
- limit go routines / TCP connection
    - create pool of collection handlers / workers
    - assigne worker from pool to new connection