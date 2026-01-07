package main

import "fmt"

func ProcessCommand(clientAddress string, cmd []string) (string, error) {
	fmt.Printf("[%s] PROCESS COMMAND, received: %s\n", clientAddress, cmd)

	if len(cmd) < 2 {
		return "NULL", fmt.Errorf("INVALID COMMAND")
	}

	resp := ""
	switch cmd[0] {
	case "SET":
		if len(cmd) != 3 {
			return "NULL", fmt.Errorf("INVALID COMMAND")
		}
		store.SetValue(cmd[1], cmd[2])
		resp = fmt.Sprintf("SET(%v):%v", cmd[1], cmd[2])
	case "GET":
		if len(cmd) != 2 {
			return "NULL", fmt.Errorf("INVALID COMMAND")
		}
		resp = fmt.Sprintf("GET(%v):%v", cmd[1], store.GetValue(cmd[1]))
	case "DEL":
		if len(cmd) != 2 {
			return "NULL", fmt.Errorf("INVALID COMMAND")
		}
		resp = fmt.Sprintf("DEL(%v)", cmd[1])
		store.DeleteKey(cmd[1])
	default:
		return "NULL", fmt.Errorf("unknown command type")
	}

	return resp, nil
}
