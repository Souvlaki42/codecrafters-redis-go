package main

import (
	"fmt"
	"net"
	"os"
)

func checkForErrors(err error, message string, printErr bool) {
	if err != nil {
		if printErr {
			fmt.Println(message, err.Error())
		} else {
			fmt.Println(message)
		}
		os.Exit(1)
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	checkForErrors(err, "Failed to bind to port 6379", false)
	c, err := l.Accept()
	checkForErrors(err, "Error accepting connection: ", true)
	_, err = c.Write([]byte("+PONG\r\n"))
	checkForErrors(err, "Error responding to commands: ", true)
	c.Close()
}
