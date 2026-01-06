package main

import (
	"fmt"
	"net"
	"syscall"
)

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
