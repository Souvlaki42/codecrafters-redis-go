package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func handleConnection(c net.Conn) {
	defer c.Close()
	for {
		b := make([]byte, 1024)
		n, err := c.Read(b)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading commands: ", err.Error())
			}
			continue
		}
		fmt.Printf("Received %d bytes: %q\n", n, b[:n])
		_, err = c.Write([]byte("+PONG\r\n"))
		if err != nil {
			fmt.Println("Error writing output: ", err.Error())
		}
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(c)
	}
}
