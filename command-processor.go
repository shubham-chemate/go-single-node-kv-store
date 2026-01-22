package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

func ProcessCommand(clientAddress string, cmd []string) (string, error) {
	if len(cmd) < 1 {
		return "", fmt.Errorf("INVALID COMMAND LENGTH")
	}

	cmd[0] = strings.ToUpper(cmd[0])

	resp := ""
	switch cmd[0] {
	case "PING":
		resp = "+OK\r\n"
	case "SET":
		if len(cmd) == 3 {
			store.SetValue(cmd[1], cmd[2], -1)
		} else if len(cmd) == 4 {
			ttl, err := strconv.Atoi(cmd[3])
			if err != nil {
				slog.Error("invalid ttl", "received", cmd[3])
				return "", fmt.Errorf("INVALID TTL")
			}
			store.SetValue(cmd[1], cmd[2], int64(ttl))
		} else {
			return "", fmt.Errorf("INVALID COMMAND")
		}
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
