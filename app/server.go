package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	c, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	for {
		b := make([]byte, 1024)
		_, err := c.Read(b)
		fmt.Println(string(b))
		if err != nil {
			fmt.Println("Error reading commands: ", err.Error())
			c.Close()
		}
		_, err = c.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error responding to commands: ", err.Error())
			os.Exit(1)
		}
	}
}
