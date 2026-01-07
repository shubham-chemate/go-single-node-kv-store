package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	READ_DEADLINE_TIME = 100
	MAX_CLIENT_CONN    = 2
	MAX_KEY_VAL_SIZE   = 5
)

var store *kvstore

// concurrent tcp-server
// graceful shutdown
// read deadline
// panic recovery
// reading data
func main() {
	store = &kvstore{mp: make(map[string]string)}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		panic("error connecting to OS for tcp listening")
	}

	fmt.Println("successfully connected to OS for tcp port 8080!")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	clientsLimiter := make(chan struct{}, MAX_CLIENT_CONN)

	go handleClients(listener, &wg, clientsLimiter)

	<-quit
	fmt.Println("Shutdown signal received, shutting down code!")
	listener.Close()
	close(quit)

	fmt.Println("waiting for active connections to close!")
	wg.Wait()

	fmt.Println("all connections closed.")
}

func handleClients(listener net.Listener, wg *sync.WaitGroup, clientsLimiter chan struct{}) {
	fmt.Println("Press CTRL+C to stop gracefully")

	for {
		conn, err := listener.Accept()
		if err != nil {
			// our goroutine was blocked on listener.Accept()
			// this specific error mostly occurs due to listener.Close()
			// we want to ignore that error and exit the goroutine
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			} else {
				fmt.Println("error while accepting client connection:", err)
				continue
			}
		}

		clientsLimiter <- struct{}{}

		wg.Add(1)
		go handleClientConnection(conn, wg, clientsLimiter)
	}
}

func handleClientConnection(conn net.Conn, wg *sync.WaitGroup, clientsLimiter chan struct{}) {
	defer func() { <-clientsLimiter }()
	defer wg.Done()
	defer conn.Close()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s] panic when processing client, message: %v", conn.RemoteAddr(), r)
		}
	}()

	conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

	clientAddress := conn.RemoteAddr().String()
	fmt.Printf("[%s] client connected\n", clientAddress)

	reader := bufio.NewReader(conn)
	// scanner := bufio.NewScanner(conn)
	// buffer := make([]byte, 1024)

	for {
		// *3\r\n$3\r\nSET\r\n$3\r\npin\r\n$6\r\n414103\r\n

		message, err := readFromClient(clientAddress, reader)
		if err != nil {
			fmt.Printf("[%s] client disconnected, received: %s\n", clientAddress, err)
			resp := "CONNECTION ERROR\n"
			conn.Write([]byte(resp))
			return
		}
		conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

		if !strings.HasPrefix(message, "*") || !strings.HasSuffix(message, "\r\n") {
			fmt.Printf("[%s] unknown format received", clientAddress)
			resp := "ERROR UNKNOWN FORMAT, received: " + message + "\n"
			conn.Write([]byte(resp))
			return
		}

		message = strings.TrimSpace(message[1:])

		arraySize, err := strconv.Atoi(message)
		if err != nil {
			fmt.Printf("[%s] error parsing the array size: %s", clientAddress, err.Error())
			resp := "ERROR PARSING THE ARRAY SIZE"
			conn.Write([]byte(resp))
			return
		}

		fmt.Printf("[%s] reading array of size: %d\n", clientAddress, arraySize)

		inputCommand := []string{}
		for range arraySize {
			message, err := readFromClient(clientAddress, reader)
			if err != nil {
				fmt.Printf("[%s] client disconnected, received: %s\n", clientAddress, err)
				resp := "CONNECTION CLOSED\n"
				conn.Write([]byte(resp))
				return
			}
			conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

			if !strings.HasPrefix(message, "$") || !strings.HasSuffix(message, "\r\n") {
				fmt.Printf("[%s] unknown format received", clientAddress)
				resp := "ERROR UNKNOWN FORMAT, received: " + message + "\n"
				conn.Write([]byte(resp))
				return
			}

			message = strings.TrimSpace(message[1:])
			cmdSize, err := strconv.Atoi(message)
			if err != nil {
				fmt.Printf("[%s] error parsing the command size: %s", clientAddress, err.Error())
				resp := "ERROR PARSING THE COMMAND STRING SIZE"
				conn.Write([]byte(resp))
				return
			}
			// put limit on cmdSize, thinking about 64K Bytes
			if cmdSize > MAX_KEY_VAL_SIZE {
				fmt.Printf("[%d] command input exceeded max size allowed\n", cmdSize)
				resp := "ERROR command input exceeded max size allowed"
				conn.Write([]byte(resp))
				return
			}

			message, err = readFromClient(clientAddress, reader)
			if err != nil {
				fmt.Printf("[%s] client disconnected, received: %s\n", clientAddress, err)
				resp := "CONNECTION CLOSED\n"
				conn.Write([]byte(resp))
				return
			}
			conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

			message = strings.TrimSpace(message)
			inputCommand = append(inputCommand, message)
		}

		val, err := ProcessCommand(clientAddress, inputCommand)
		if err != nil {
			fmt.Printf("[%s] error occurred while processing the command\n", clientAddress)
			resp := "ERROR PROCESSING THE COMMAND, message: " + err.Error() + "\n"
			conn.Write([]byte(resp))
			return
		}
		resp := "ACK\n"
		resp += val + "\n"

		// message := scanner.Text()

		// n, err := conn.Read(buffer)
		// if err != nil {
		// 	if err == io.EOF {
		// 		fmt.Printf("[%s] got EOF (nothing to read): %s\n", clientAddress, err)
		// 	} else {
		// 		fmt.Printf("[%s] error while reading from connection: %s\n", clientAddress, err)
		// 	}
		// 	return
		// }

		// message := string(buffer[:n])

		// message = strings.TrimSpace(message)
		// fmt.Printf("[%s] received message: %s\n", clientAddress, message)

		// time.Sleep(3 * time.Second)

		// resp := fmt.Sprintf("ACK: %s\n", message)
		conn.Write([]byte(resp))
	}

	// if err := scanner.Err(); err != nil {
	// 	fmt.Println("error in scanner:", err)
	// }
}
