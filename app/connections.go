package main

import (
	"fmt"
	"io"
	"net"
)

func handleConnection(connection net.Conn, flags Flags) {
	defer connection.Close()
	for {
		bytesValue := make([]byte, 1024)
		numberOfBytes, err := connection.Read(bytesValue)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading commands: ", err.Error())
			}
			continue
		}

		raw_command := bytesValue[:numberOfBytes]

		command, output := handleCommand(raw_command, flags)

		_, err = connection.Write([]byte(output))

		if err != nil {
			fmt.Println("Error writing output: ", err.Error())
		}

		fmt.Printf("Received %d bytes: %q\n", numberOfBytes, command)
	}
}
