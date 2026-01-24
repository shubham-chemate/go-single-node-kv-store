package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	errUnknownFormat     = "-ERR UNKNOWN FORMAT RECEIVED: %s\r\n"
	errArraySizeParsing  = "-ERR PARSING THE ARRAY SIZE\r\n"
	errConnectionClosed  = "-ERR CONNECTION CLOSED\r\n"
	errStringSizeParsing = "-ERR PARSING THE COMMAND STRING SIZE\r\n"
	errMaxSizeExceeded   = "-ERR COMMAND INPUT EXCEEDED MAX SIZE ALLOWED\r\n"
	errEmptyCommandSize  = "-ERR EMPTY COMMAND SIZE\r\n"
	errProcessingCommand = "-ERR PROCESSING THE COMMAND, message: %s\r\n"
)

func handleClient(conn net.Conn, wg *sync.WaitGroup, clientsLimiter chan struct{}) {
	defer func() { <-clientsLimiter }()
	defer wg.Done()
	defer conn.Close()
	defer recoverFromPanic(conn)

	clientAddress := conn.RemoteAddr().String()
	logger := slog.With("client_address", clientAddress)
	logger.Info("client connected")

	resetReadDeadline(conn)
	reader := bufio.NewReader(conn)

	for {
		// *3\r\n$3\r\nSET\r\n$3\r\npin\r\n$6\r\n414103\r\n
		if err := processRequest(conn, reader, logger); err != nil {
			return
		}
	}
}

func processRequest(conn net.Conn, reader *bufio.Reader, logger *slog.Logger) error {

	arraySize, err := parseArraySize(reader, logger)
	if err != nil {
		writeError(conn, err.Error())
		return err
	}

	resetReadDeadline(conn)

	inputCommand, err := parseCommandArray(conn, reader, logger, arraySize)
	if err != nil {
		writeError(conn, err.Error())
		return err
	}

	resetReadDeadline(conn)

	resp, err := ProcessCommand(inputCommand)
	if err != nil {
		logger.Info("error occurred while processing the command", "command", inputCommand)
		writeError(conn, fmt.Sprintf(errProcessingCommand, err.Error()))
		return err
	}

	conn.Write([]byte(resp))
	return nil
}

func writeError(conn net.Conn, errMessage string) {
	conn.Write([]byte(errMessage))
}

func recoverFromPanic(conn net.Conn) {
	if r := recover(); r != nil {
		fmt.Printf("[%s] panic when processing client, message: %v", conn.RemoteAddr(), r)
	}
}

func resetReadDeadline(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))
}
