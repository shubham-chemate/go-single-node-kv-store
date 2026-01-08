package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	READ_DEADLINE_TIME = 60
	MAX_CLIENT_CONN    = 12000
	MAX_KEY_VAL_SIZE   = 1000
	CLEANER_FREQUENCY  = 40
)

var store *kvstore

// concurrent tcp-server
// graceful shutdown
// read deadline
// panic recovery
// reading data
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store = &kvstore{mp: make(map[string]Entry)}
	go store.StartStoreCleaner()

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		slog.Error(err.Error())
		panic("error connecting to OS for tcp listening")
	}

	slog.Info("successfully connected to OS for tcp port 8080!")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	clientsLimiter := make(chan struct{}, MAX_CLIENT_CONN)

	go handleClients(listener, &wg, clientsLimiter)

	<-quit

	slog.Info("Shutdown signal received, shutting down program!")
	listener.Close()
	close(quit)

	slog.Info("waiting for active connections to close!")
	wg.Wait()

	slog.Info("all connections closed.")
}

func handleClients(listener net.Listener, wg *sync.WaitGroup, clientsLimiter chan struct{}) {
	slog.Info("Press CTRL+C to stop gracefully")

	for {
		conn, err := listener.Accept()
		if err != nil {
			// our goroutine was blocked on listener.Accept()
			// this specific error mostly occurs due to listener.Close()
			// we want to ignore that error and exit the goroutine
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			} else {
				slog.Error("error while accepting new client connection: ", "ERROR", err.Error())
				continue
			}
		}

		clientsLimiter <- struct{}{}

		wg.Add(1)
		go handleClientConnection(conn, wg, clientsLimiter)
	}
}

func handleClientConnection(conn net.Conn, wg *sync.WaitGroup, clientsLimiter chan struct{}) {
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
	// scanner := bufio.NewScanner(conn)
	// buffer := make([]byte, 1024)

	for {
		// *3\r\n$3\r\nSET\r\n$3\r\npin\r\n$6\r\n414103\r\n

		message, err := readFromClient(clientAddress, reader)
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

		logger.Info("reading array", "size", arraySize)

		inputCommand := []string{}
		for range arraySize {
			message, err := readFromClient(clientAddress, reader)
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

			message, err = readFromClient(clientAddress, reader)
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

		// message := scanner.Text()

		// n, err := conn.Read(buffer)
		// if err != nil {
		// 	if err == io.EOF {
		// 		fmt.Printf("[%s] got EOF (nothing to read): %s\n", clientAddress, err)
		// 	} else {
		// 		fmt.Printf("[%s] error while reading from connection: %s\n", clientAddress, err)
		// 	}
		// 	return
		// }

		// message := string(buffer[:n])

		// message = strings.TrimSpace(message)
		// fmt.Printf("[%s] received message: %s\n", clientAddress, message)

		// time.Sleep(3 * time.Second)

		// resp := fmt.Sprintf("ACK: %s\n", message)
		conn.Write([]byte(resp))
	}

	// if err := scanner.Err(); err != nil {
	// 	fmt.Println("error in scanner:", err)
	// }
}
