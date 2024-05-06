package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var data = make(map[string]string)
var mutex sync.Mutex

func parseRESP(data []byte) ([]string, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	return RESP(reader)
}

func handleCommand(raw_command []byte) ([]string, string) {
	command, err := parseRESP(raw_command)

	if err != nil {
		fmt.Println("Error parsing input: ", err.Error())
	}

	output := ""

	switch strings.ToLower(command[0]) {
	case "command":
		output = "+\r\n"
	case "ping":
		output = "+PONG\r\n"
	case "echo":
		output = fmt.Sprintf("+%s\r\n", strings.Join(command[1:], " "))
	case "set":
		mutex.Lock()
		defer mutex.Unlock()

		data[command[1]] = command[2]
		if len(command) == 5 && strings.ToLower(command[3]) == "px" {
			expr, err := strconv.ParseUint(command[4], 10, 64)
			if err != nil {
				fmt.Println("Error parsing expiration time: ", err.Error())
			}
			expirationTime := time.Now().Add(time.Duration(expr) * time.Millisecond)

			go func(k string, exp time.Time) {
				time.Sleep(time.Until(exp))
				mutex.Lock()
				defer mutex.Unlock()
				data[k] = ""
			}(command[1], expirationTime)
		}
		output = "+OK\r\n"
	case "get":
		item := data[command[1]]
		if item == "" {
			output = "$-1\r\n"
		} else {
			output = fmt.Sprintf("$%d\r\n%s\r\n", len([]byte(item)), item)
		}
	default:
		fmt.Printf("The command you gave: %q, isn't a valid one yet\r\n", command[0])
		os.Exit(1)
	}
	return command, output
}

func handleConnection(connection net.Conn) {
	defer connection.Close()
	for {
		bytes := make([]byte, 1024)
		numberOfBytes, err := connection.Read(bytes)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading commands: ", err.Error())
			}
			continue
		}

		raw_command := bytes[:numberOfBytes]

		command, output := handleCommand(raw_command)

		_, err = connection.Write([]byte(output))

		if err != nil {
			fmt.Println("Error writing output: ", err.Error())
		}

		fmt.Printf("Received %d bytes: %q\n", numberOfBytes, command)
	}
}

func main() {
	var port uint
	flag.UintVar(&port, "port", 6379, "Specify the port number")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	fmt.Printf("Server binded to port %d...\r\n", port)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\r\n", port)
		os.Exit(1)
	}
	defer listener.Close()
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection)
	}
}
