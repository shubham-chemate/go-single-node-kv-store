package main

import "fmt"

func ProcessCommand(clientAddress string, cmd []string) error {
	fmt.Printf("[%s] PROCESS COMMAND, received: %s\n", clientAddress, cmd)

	if len(cmd) < 2 {
		return fmt.Errorf("INVALID COMMAND")
	}

	return nil
}
