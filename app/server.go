package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func handleCommands(commands []string, data map[string]string) (string, []string, string) {
	command, args, output := strings.ToLower(commands[0]), commands[1:], ""
	switch command {
	case "command":
	case "ping":
		output = "+PONG\r\n"
	case "echo":
		output = fmt.Sprintf("+%s\r\n", strings.Join(args, " "))
	case "set":
		data[args[0]] = args[1]
		output = "+OK\r\n"
	case "get":
		item := data[args[0]]
		output = fmt.Sprintf("$3\r\n%s\r\n", item)
	default:
		fmt.Printf("Command %q is not yet acceptable\r\n", command)
		os.Exit(1)
	}
	return command, args, output
}

func handleConnection(connection net.Conn, data map[string]string) {
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
		commands, err := parseRESP(raw_command)

		if err != nil {
			fmt.Println("Error parsing RESP:", err.Error())
		}

		command, args, output := handleCommands(commands, data)

		_, err = connection.Write([]byte(output))
		if err != nil {
			fmt.Println("Error writing output: ", err.Error())
		}

		fmt.Printf("Received %d bytes: %q, %q\n", numberOfBytes, command, args)
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	var data map[string]string
	defer listener.Close()
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection, data)
	}
}
