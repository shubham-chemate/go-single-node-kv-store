package main

import (
	"fmt"
	"strings"
)

func ProcessCommand(clientAddress string, cmd []string) (string, error) {
	fmt.Printf("[%s] PROCESS COMMAND, %s\n", clientAddress, cmd)

	if len(cmd) < 1 {
		return "", fmt.Errorf("INVALID COMMAND LENGTH")
	}

	cmd[0] = strings.ToUpper(cmd[0])

	resp := ""
	switch cmd[0] {
	case "PING":
		resp = "+OK\r\n"
	case "SET":
		if len(cmd) != 3 {
			return "", fmt.Errorf("INVALID COMMAND")
		}
		store.SetValue(cmd[1], cmd[2])
		resp = "+OK\r\n"
	case "GET":
		if len(cmd) != 2 {
			return "", fmt.Errorf("INVALID COMMAND")
		}
		val := store.GetValue(cmd[1])
		if val != "" {
			resp = fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
		} else {
			resp = "$-1\r\n"
		}
	case "DEL":
		if len(cmd) != 2 {
			return "", fmt.Errorf("INVALID COMMAND")
		}
		store.DeleteKey(cmd[1])
		resp = "+OK\r\n"
	default:
		return "", fmt.Errorf("unknown command type")
	}

	return resp, nil
}
