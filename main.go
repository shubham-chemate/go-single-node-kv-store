package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
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
		panic("Error connecting to OS for tcp listening")
	}

	fmt.Println("successfully connected to OS for tcp port 8080!")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	clientsLimiter := make(chan struct{}, 2)

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
		clientsLimiter <- struct{}{}
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error while accepting client connection:", err)
			return
		}

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

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	clientAddress := conn.RemoteAddr().String()
	fmt.Printf("client [%s] connected\n", clientAddress)

	// reader := bufio.NewReader(conn)
	// scanner := bufio.NewScanner(conn)
	buffer := make([]byte, 1024)

	for {
		// message, err := reader.ReadString('\n')
		// if err != nil {
		// 	fmt.Printf("received: %s, client [%s] disconnected.\n", err, clientAddress)
		// 	return
		// }

		// message := scanner.Text()

		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[%s] got EOF (nothing to read): %s\n", clientAddress, err)
			} else {
				fmt.Printf("[%s] error while reading from connection: %s\n", clientAddress, err)
			}
			return
		}

		message := string(buffer[:n])

		message = strings.TrimSpace(message)
		fmt.Printf("[%s] received message: %s\n", clientAddress, message)

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		time.Sleep(3 * time.Second)

		resp := fmt.Sprintf("ACK: %s\n", message)
		conn.Write([]byte(resp))
	}

	// if err := scanner.Err(); err != nil {
	// 	fmt.Println("Error in scanner:", err)
	// }
}
