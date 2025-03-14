package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

var data = make(map[string]string)
var mutex sync.Mutex
var replicas = make(chan *net.TCPConn)

func main() {
	flags := parseFlags()

	if !flags.is_master {
		con, err := handleHandshake(flags)

		replicas <- con

		fmt.Printf("Hapenned?: %d\r\n", len(replicas))

		if err != nil {
			fmt.Printf("Failed to ping master %s:%d:%s\r\n", flags.master_host, flags.master_port, err.Error())
			os.Exit(1)
		}

		defer con.Close()
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", flags.port))
	fmt.Printf("Server binded to port %d...\r\n", flags.port)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\r\n", flags.port)
		os.Exit(1)
	}
	defer listener.Close()
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection, flags)
	}
}
