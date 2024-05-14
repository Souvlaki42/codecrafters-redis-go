package main

import (
	"fmt"
	"net"
)

func handleHandshake(flags Flags) (*net.TCPConn, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", flags.master_host, flags.master_port))
	if err != nil {
		return nil, err
	}

	connection, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		return nil, err
	}

	commands := [4]string{
		"*1\r\n$4\r\nPING\r\n",
		fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%d\r\n", flags.port),
		"*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n",
		"*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n",
	}

	for _, command := range commands {
		_, err = connection.Write([]byte(command))
		if err != nil {
			return connection, err
		}

		buf := make([]byte, 1024)
		_, err = connection.Read(buf)
		if err != nil {
			return connection, err
		}
	}

	return connection, nil
}
