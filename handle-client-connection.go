package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func handleClient(conn net.Conn, wg *sync.WaitGroup, clientsLimiter chan struct{}) {
	defer func() { <-clientsLimiter }()
	defer wg.Done()
	defer conn.Close()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s] panic when processing client, message: %v", conn.RemoteAddr(), r)
		}
	}()

	clientAddress := conn.RemoteAddr().String()

	logger := slog.With("client_address", clientAddress)

	conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

	logger.Info("client connected")

	reader := bufio.NewReader(conn)

	for {
		// *3\r\n$3\r\nSET\r\n$3\r\npin\r\n$6\r\n414103\r\n

		message, err := readSizeInput(reader)
		if err != nil {
			logger.Info("client disconnected", "msg", err.Error())
			resp := "-ERR connection closed\r\n"
			conn.Write([]byte(resp))
			return
		}
		conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

		if !strings.HasPrefix(message, "*") || !strings.HasSuffix(message, "\r\n") {
			logger.Error("unknown format for size of array")
			resp := "-ERR UNKNOWN FORMAT received: " + message + "\r\n"
			conn.Write([]byte(resp))
			return
		}

		message = strings.TrimSpace(message[1:])

		arraySize, err := strconv.Atoi(message)
		if err != nil {
			logger.Error("invalid value for array size", "received", message, "msg", err.Error())
			resp := "-ERR PARSING THE ARRAY SIZE\r\n"
			conn.Write([]byte(resp))
			return
		}

		// logger.Info("reading array", "size", arraySize)

		inputCommand := []string{}
		for range arraySize {
			message, err := readSizeInput(reader)
			if err != nil {
				logger.Info("client disconnected", "msg", err.Error())
				resp := "-ERR CONNECTION CLOSED\r\n"
				conn.Write([]byte(resp))
				return
			}
			conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

			if !strings.HasPrefix(message, "$") || !strings.HasSuffix(message, "\r\n") {
				logger.Error("invalid format of string size", "received", message)
				resp := "-ERR UNKNOWN FORMAT received: " + message + "\r\n"
				conn.Write([]byte(resp))
				return
			}

			message = strings.TrimSpace(message[1:])
			cmdSize, err := strconv.Atoi(message)
			if err != nil {
				logger.Error("invalid value of string size", "received", message, "msg", err.Error())
				resp := "-ERR PARSING THE COMMAND STRING SIZE\r\n"
				conn.Write([]byte(resp))
				return
			}
			// put limit on cmdSize, thinking about 64K Bytes
			if cmdSize > MAX_KEY_VAL_SIZE {
				logger.Error("string size exceeded max size allowed", "allowed", MAX_KEY_VAL_SIZE, "given", cmdSize)
				resp := "-ERR command input exceeded max size allowed\r\n"
				conn.Write([]byte(resp))
				return
			}
			if cmdSize < 1 {
				logger.Error("empty command size not allowed")
				resp := "-ERR empty command size\r\n"
				conn.Write([]byte(resp))
				return
			}

			message, err = readTextInput(reader, cmdSize)
			if err != nil {
				logger.Info("client disconnected", "msg", err.Error())
				resp := "-ERR CONNECTION CLOSED\r\n"
				conn.Write([]byte(resp))
				return
			}
			conn.SetReadDeadline(time.Now().Add(READ_DEADLINE_TIME * time.Second))

			message = strings.TrimSpace(message)
			inputCommand = append(inputCommand, message)
		}

		resp, err := ProcessCommand(clientAddress, inputCommand)
		if err != nil {
			logger.Info("error occurred while processing the command", "command", inputCommand)
			resp := "-ERR PROCESSING THE COMMAND, message: " + err.Error() + "\r\n"
			conn.Write([]byte(resp))
			return
		}
		conn.Write([]byte(resp))
	}
}
