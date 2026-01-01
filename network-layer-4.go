package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

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
