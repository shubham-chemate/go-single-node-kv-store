package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
)

func parseCommandArray(conn net.Conn, reader *bufio.Reader, logger *slog.Logger, arraySize int) ([]string, error) {
	inputCommand := make([]string, 0)

	for range arraySize {
		resetReadDeadline(conn)

		cmdSize, err := parseStringSize(reader, logger)
		if err != nil {
			return nil, err
		}

		if err := validateCommandSize(cmdSize, logger); err != nil {
			return nil, err
		}

		resetReadDeadline(conn)

		text, err := readAndParseText(reader, logger, cmdSize)
		if err != nil {
			return nil, err
		}

		inputCommand = append(inputCommand, text)
	}

	return inputCommand, nil
}

func readAndParseText(reader *bufio.Reader, logger *slog.Logger, cmdSize int) (string, error) {
	message, err := readTextInput(reader, cmdSize)
	if err != nil {
		logger.Info("client disconnected", "msg", err.Error())
		return "", fmt.Errorf(errConnectionClosed)
	}

	return strings.TrimSpace(message), nil
}

func validateCommandSize(cmdSize int, logger *slog.Logger) error {
	if cmdSize > MAX_KEY_VAL_SIZE {
		logger.Error("string size exceeded max size allowed", "allowed", MAX_KEY_VAL_SIZE, "given", cmdSize)
		return fmt.Errorf(errMaxSizeExceeded)
	}
	if cmdSize < 1 {
		logger.Error("empty command size not allowed")
		return fmt.Errorf(errEmptyCommandSize)
	}
	return nil
}

func parseStringSize(reader *bufio.Reader, logger *slog.Logger) (int, error) {
	message, err := readSizeInput(reader)
	if err != nil {
		logger.Info("client disconnected", "msg", err.Error())
		return 0, fmt.Errorf(errConnectionClosed)
	}

	if !strings.HasPrefix(message, "$") || !strings.HasSuffix(message, "\r\n") {
		logger.Error("invalid format of string size", "received", message)
		return 0, fmt.Errorf(errUnknownFormat, message)
	}

	message = strings.TrimSpace(message[1:])
	cmdSize, err := strconv.Atoi(message)
	if err != nil {
		logger.Error("invalid value of string size", "received", message, "msg", err.Error())
		return 0, fmt.Errorf(errStringSizeParsing)
	}

	return cmdSize, nil
}

func parseArraySize(reader *bufio.Reader, logger *slog.Logger) (int, error) {
	message, err := readSizeInput(reader)
	if err != nil {
		logger.Info("client disconnected", "msg", err.Error())
		return 0, fmt.Errorf(errConnectionClosed)
	}

	if !strings.HasPrefix(message, "*") || !strings.HasSuffix(message, "\r\n") {
		logger.Error("unknown format for size of array")
		return 0, fmt.Errorf(errUnknownFormat, message)
	}

	message = strings.TrimSpace(message[1:])

	arraySize, err := strconv.Atoi(message)
	if err != nil {
		logger.Error("invalid value for array size", "received", message, "msg", err.Error())
		return 0, fmt.Errorf(errArraySizeParsing)
	}

	return arraySize, nil
}
