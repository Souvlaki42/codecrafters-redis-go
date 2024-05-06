package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

var data = make(map[string]string)

func parseRESP(input []byte) []string {
	slice := strings.Split(string(input), "\r\n")
	var result []string
	reg := regexp.MustCompile("[^a-zA-Z]+")
	for _, str := range slice {
		cleanStr := reg.ReplaceAllString(str, "")
		if cleanStr != "" {
			result = append(result, cleanStr)
		}
	}
	return result
}

// func handleCommand() (command []string, err error, output []byte) {

// }

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
		command := parseRESP(raw_command)

		commandName := strings.ToLower(command[0])
		commandArgs := command[1:]

		switch commandName {
		case "command":
			_, err = connection.Write([]byte("+\r\n"))
		case "ping":
			_, err = connection.Write([]byte("+PONG\r\n"))
		case "echo":
			_, err = connection.Write([]byte(fmt.Sprintf("+%s\r\n", strings.Join(commandArgs, " "))))
		case "set":
			data[commandArgs[0]] = commandArgs[1]
			_, err = connection.Write([]byte("+OK\r\n"))
		case "get":
			item := data[commandArgs[0]]
			_, err = connection.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len([]byte(item)), item)))
		default:
			fmt.Printf("The command you gave: %q, isn't a valid one yet\r\n", commandName)
			os.Exit(1)
		}

		if err != nil {
			fmt.Println("Error writing output: ", err.Error())
		}

		fmt.Printf("Received %d bytes: %q\n", numberOfBytes, command)
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	fmt.Println("Server binded to port 6379...")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
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
