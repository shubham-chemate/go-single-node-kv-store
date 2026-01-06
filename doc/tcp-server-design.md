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
- network layer 2 always gets data in groups of 1518 (or less) bytes, called frame, this is the capacity of transfer for standard ethernet
- in those 1518 byes, first 6 bytes are src MAC address, next 6 bytes are dest MAC address, next 2 bytes are signal for Layer 3 and remaining bytes are data
- so each frame is collection of unique src MAC, unique dest MAC, Layer 3 helper & data

following go code asks OS for frames and fetches the frame details from the received frames
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

### Network Layer 3

- Data is passed from layer 2 to layer 3, 18 bits (MAC address + info for layer 3) are stripped from frame and frames are grouped as per their ip addresses by OS to form packets
- so each packet is a group of related frames that OS collects over a network, large movie is broken into frames, transfered over network, then that frames are again collected to form layer 3 packet

this code asks OS for ip4 packets and print some details from the packet, it is also possible to print the data from the packet (in for loop)
```go
func NetworkLayer3Main() {
	conn, err := net.ListenIP("ip4:icmp", &net.IPAddr{IP: net.ParseIP("0.0.0.0")})
	if err != nil {
		fmt.Println("Error (Did you run with sudo?):", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Listening for Raw ICMP Packets...")

	incoming := make([]byte, 1024)
	for {

		length, srcIp, err := conn.ReadFrom(incoming)
		if err != nil {
			fmt.Println("got error: ", err)
			continue
		}

		fmt.Printf("Received %d bytes from %s\n", length, srcIp)

		data := incoming[:length]

		// fmt.Printf("Data: %v\n", data)
		fmt.Printf("Text: %s\n", string(data))
	}
}
```

### Network Layer 4 (TCP)

- this is real layer where work is kind of handed from OS to software program
- packets are out of order when received in layer 3, OS collects all packets and order them, OS provides ordered packets to our program as a stream of bytes
- packets in this layer are called as segments
- we request OS to give certain kind of segments to us by mentioning PORT, since there are many other application simultaneously looking for TCP segments, we want segments who are sent for us (mentioned by PORT)

We asks OS to give the segments from PORT 8080 to us
simultaneously EMAIL will be asking for TCP segments from PORT 1010, browser might be asking for TCP segments from PORT 4010 etc  

Now OS will redirect stream of packets from PORT 8080 to our program
```go
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		panic("Error connecting to os for tcp listening")
	}
	defer listener.Close()
```

There are multiple clients that are sending segments on PORT 8080, we create connection for each of the client separately and that connection will be open until we close it or client close it, OS don't close the connection

```go
func NetworkLayer4Main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		panic("Error connecting to os for tcp listening")
	}
	defer listener.Close()

	fmt.Println("successfully connected to os for tcp port 8080!")

	for {
		conn, _ := listener.Accept()

		fmt.Println("accepted client conection..")

		// reading data in chunks of 1024 bytes
		buffer := make([]byte, 1024)
		length, _ := conn.Read(buffer)

		message := string(buffer[:length])

		fmt.Println("received message:", message)
		fmt.Println("processing incoming message...")
		time.Sleep(10 * time.Second)
		fmt.Println("incoming message processing completed!")

		if strings.TrimSpace(message) == "hi" {
			conn.Write([]byte("hey there!\n"))
		}

		fmt.Println("closing the connection!")
		conn.Close()
	}

}
```

## Concurrent TCP Server

- The TCP server that is mentioned above is very simple
- It asks OS to redirect segments for PORT 8080
- But it handles the single client at a time, this means second client have to wait until connection of first client is closed
- We can work through that by acceptiong connections from clients as they come and handling each connection in a separate goroutine
- This again create different set of challenges

```go
func ConcurrentTCPServerMain() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		panic("Error connecting to os for tcp listening")
	}
	defer listener.Close()

	fmt.Println("successfully connected to os for tcp port 8080!")

	for {
		conn, _ := listener.Accept()
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddress := conn.RemoteAddr().String()
	fmt.Printf("client [%s] connected\n", clientAddress)

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[%s] disconnected\n", clientAddress)
			return
		}

		message = strings.TrimSpace(message)
		fmt.Printf("[%s] received message: %s\n", clientAddress, message)

		resp := fmt.Sprintf("server received: %s\n", message)
		conn.Write([]byte(resp))

		time.Sleep(5 * time.Second)
	}
}
```

### Challenge: Client Establishes Connection but doesn't send data - SetReadDeadline
- That's bad, OS don't have TCP timeout, it can keep the connection open as long as you want
- Read operation is blocking
- The goroutine will be blocked until clients sends the data, so you have to wait indefinetly if client never send the data
- If client connects to your server but never sends the data, it will keep the connection open but you won’t be reading any data from it, this is bad since you have active blocked goroutine waiting for data
- To avoid this we use SetReadDeadline, where we put the deadline on client to send the data, if he don’t in that timeframe, we simply close the connection
- Why client is not sending data even after establishing connection - buggy, malicious, attack, slow etc reasons
- A common and effective pattern that is used is, resetting deadline after each successful read