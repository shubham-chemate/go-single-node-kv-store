package main

import (
	"bufio"
	"fmt"
	"log/slog"
)

func readFromClient(clientAddress string, reader *bufio.Reader) (string, error) {
	input := []byte{}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}

		input = append(input, b)
		if len(input) > MAX_KEY_VAL_SIZE {
			return "", fmt.Errorf("max allowed key value length exceeded")
		}

		if b == '\n' {
			break
		}
	}

	resp := string(input)

	slog.Info("reading bytes", "read", resp)

	return resp, nil
}
