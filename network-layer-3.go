package main

import (
	"fmt"
	"net"
	"os"
)

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
