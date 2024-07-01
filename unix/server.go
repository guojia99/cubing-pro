package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func handleConnection(conn net.Conn, dataReceived *int64) {
	defer conn.Close()

	buf := make([]byte, 1024*1024*1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		*dataReceived += int64(n)
	}
}

func main() {
	socketPath := "/tmp/unix_socket"
	if err := os.RemoveAll(socketPath); err != nil {
		fmt.Println("Failed to remove existing socket file:", err)
		return
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Failed to listen on unix socket:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on", socketPath)

	var dataReceived int64
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Printf("Data received: %.5f MiB/sec\n", float64(dataReceived)/1024/1024)
			dataReceived = 0
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			return
		}
		go handleConnection(conn, &dataReceived)
	}
}
