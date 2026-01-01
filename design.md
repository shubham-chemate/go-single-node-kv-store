## Problem Statement:  
we want to build single node kv store (redis)

key architectural decisions:
1. TCP Server
2. RESP Parsing
3. Thread Safe Storage
4. Command Support


## Why TCP and not HTTP server?
- efficiency and performance
- HTTP header contains lot of data (heavy corporate envelop)
- TCP is small (plain letter)
- TCP allows to have fast parsing
- TCP allows connection flexibility, we can keep connection open for as much as we like (hours or days)

HTTP Post (Size: 150-200 Bytes)
```http
POST /set HTTP/1.1
Host: localhost:6379
User-Agent: Go-Client/1.0
Content-Type: application/json
Content-Length: 13
Accept: */*

{"key":"a", "val":1}
```

Raw TCP (Redis Protocol, 20 Bytes)
```
*3\r\n$3\r\nSET\r\n$1\r\na\r\n$1\r\n1\r\n
```