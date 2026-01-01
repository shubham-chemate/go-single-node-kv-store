package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// concurrent tcp-server
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
