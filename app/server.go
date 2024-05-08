package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

var data = make(map[string]string)
var replica_ports = make([]uint64, 0)
var mutex sync.Mutex

func main() {
	flags := parseFlags()

	if !flags.is_master {
		err := handleHandshake(fmt.Sprintf("%s:%d", flags.master_host, flags.master_port), []string{
			"*1\r\n$4\r\nPING\r\n",
			"*2\r\n$4\r\necho\r\n$3\r\nhey",
			fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%d\r\n", flags.port),
			"*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n",
			"*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n",
		}, true)
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
