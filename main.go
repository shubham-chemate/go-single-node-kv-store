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
)

// concurrent tcp-server
// graceful shutdown
// read deadline
// panic recovery
// reading data
func main() {
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
			fmt.Printf("[%s] PANIC while dealing with connection: %v", conn.RemoteAddr(), r)
		}
	}()

	conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

	clientAddress := conn.RemoteAddr().String()
	fmt.Printf("client [%s] connected\n", clientAddress)

	reader := bufio.NewReader(conn)
	// scanner := bufio.NewScanner(conn)
	// buffer := make([]byte, 1024)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("received: %s, client [%s] disconnected.\n", err, clientAddress)
			return
		}

		resp := "ACK"
		if strings.HasPrefix(message, "*") || !strings.HasSuffix(message, "\r\n") {
			message = message[1 : len(message)-2]

			tokenCnt, err := strconv.Atoi(message)
			if err != nil {
				fmt.Println("error parsing the token count:", err)
				return
			}

			fmt.Println("token size:", tokenCnt)

			// switch message {
			// case "SET":

			// case "GET":

			// case "DEL":

			// default:
			// 	resp = "UNKNOWN COMMAND, received: " + message
			// }
		} else {
			resp = "UNKNOWN FORMAT, received: " + message
		}

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

		conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

		// time.Sleep(3 * time.Second)

		// resp := fmt.Sprintf("ACK: %s\n", message)
		conn.Write([]byte(resp))
	}

	// if err := scanner.Err(); err != nil {
	// 	fmt.Println("error in scanner:", err)
	// }
}
