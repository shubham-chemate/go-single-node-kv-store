package main

import (
	"bufio"
	"fmt"
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

	go handleClients(listener, &wg)

	<-quit
	fmt.Println("Shutdown signal received, shutting down code!")
	listener.Close()
	close(quit)

	fmt.Println("waiting for active connections to close!")
	wg.Wait()

	fmt.Println("all connections closed.")
}

func handleClients(listener net.Listener, wg *sync.WaitGroup) {
	fmt.Println("Press CTRL+C to stop gracefully")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error while accepting client connection:", err)
			return
		}

		wg.Add(1)
		go handleClientConnection(conn, wg)
	}
}

func handleClientConnection(conn net.Conn, wg *sync.WaitGroup) {
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

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("received: %s, client [%s] disconnected.\n", err, clientAddress)
			return
		}

		message = strings.TrimSpace(message)
		fmt.Printf("[%s] received message: %s\n", clientAddress, message)

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		time.Sleep(5 * time.Second)

		resp := fmt.Sprintf("ACK: %s\n", message)
		conn.Write([]byte(resp))

	}
}
