package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

var data = make(map[string]string)
var mutex sync.Mutex

func main() {
	flags := parseFlags(os.Args)

	isSlave := flags.master_host != ""
	if isSlave {
		err := handleHandshake(flags)
		if err != nil {
			fmt.Printf("Failed to ping master %s:%d\r\n", flags.master_host, flags.master_port)
			os.Exit(1)
		}
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
