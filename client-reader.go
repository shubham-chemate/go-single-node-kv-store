package main

import (
	"bufio"
	"fmt"
	"io"
)

func readSizeInput(reader *bufio.Reader) (string, error) {
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
	return resp, nil
}

func readTextInput(reader *bufio.Reader, n int) (string, error) {
	input := make([]byte, n+2)
	_, err := io.ReadFull(reader, input)
	if err != nil {
		return "", fmt.Errorf("error reading the string of given size")
	}
	resp := string(input)
	return resp, nil
}
