package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	READ_DEADLINE_TIME = 60
	MAX_CLIENT_CONN    = 12000
	MAX_KEY_VAL_SIZE   = 1000
	CLEANER_FREQUENCY  = 40 // in seconds
	NUMBER_OF_SHARDS   = 32
)

var store *ShardedKVStore

// concurrent tcp-server
// graceful shutdown
// read deadline
// panic recovery
// reading data
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	store = GetNewShardedKVStore(NUMBER_OF_SHARDS)

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		slog.Error(err.Error())
		panic("error connecting to OS for tcp listening")
	}

	slog.Info("successfully connected to OS for tcp port 8080!")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go handleClients(listener)

	<-quit

	slog.Info("Shutdown signal received, shutting down program!")
	listener.Close()
	close(quit)

	slog.Info("all connections closed.")
}

func handleClients(listener net.Listener) {
	slog.Info("Press CTRL+C to stop gracefully")

	clientsLimiter := make(chan struct{}, MAX_CLIENT_CONN)
	var wg sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			// our goroutine was blocked on listener.Accept()
			// this specific error mostly occurs due to listener.Close()
			// we want to ignore that error and exit the goroutine
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			} else {
				slog.Error("error while accepting new client connection: ", "ERROR", err.Error())
				continue
			}
		}

		clientsLimiter <- struct{}{}

		wg.Add(1)
		go handleClientConnection(conn, &wg, clientsLimiter)
	}

	slog.Info("waiting for active connections to close!")
	wg.Wait()
}
