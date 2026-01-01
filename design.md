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

## How To Create Multi-Threaded TCP Server

- server is a software that "serves" something as per request from client
- how something is served on the network level?
- we will understand this by considering a example where we are are downloading huge amount of data from some server, eg. movie from netflix
- there are network layers

### Network Layer 1 & 2
- we receive data in form of 0 & 1 bits in layer 1
- that bits are grouped to form bytes
- network layer 2 & above consumes those bytes by following certain protocols to fetch the required information
- network layer 2 always gets 1518 (or less) bytes, called frame, this is the capacity of transfer for standard ethernet
- in those 1518 byes, first 6 bytes are src MAC address, next 6 bytes are dest MAC address, next 2 bytes are signal for Layer 3 and remaining bytes are data
- so each frame is collection of unique src MAC, unique dest MAC, Layer 3 helper & data

following go code asks os for frames and fetches the frame details from the received frames
```go
func NetworkLayer2Main() {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, 768)
	if err != nil {
		fmt.Println("Error: Do you have sudo?", err)
		return
	}
	defer syscall.Close(fd)

	fmt.Println("Listening for layer 2 frames..")
	incoming := make([]byte, 1518)

	for {
		n, _, err := syscall.Recvfrom(fd, incoming, 0)
		if err != nil {
			continue
		}
		if n <= 14 {
			continue
		}

		dstMac := net.HardwareAddr(incoming[0:6])
		srcMac := net.HardwareAddr(incoming[6:12])

		// layer - 3 protocol, IP or ARP
		etherType := fmt.Sprintf("%x%x", incoming[12], incoming[13])

		fmt.Printf("Frame: %s -> %s [Type: 0x%s] Size: %d bytes\n", srcMac, dstMac, etherType, n)
	}
}
```