package main

import (
	"net"
)

func handleHandshake(address string, commands []string) error {
	tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	connection, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		return err
	}

	defer connection.Close()

	for _, command := range commands {
		_, err = connection.Write([]byte(command))
		if err != nil {
			return err
		}

		buf := make([]byte, 1024)
		_, err = connection.Read(buf)
		if err != nil {
			return err
		}
	}

	return nil
}
