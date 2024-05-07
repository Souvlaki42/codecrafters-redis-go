package main

import (
	"bufio"
	"bytes"
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

type Flags struct {
	port        uint64
	master_host string
	master_port uint64
}

func (f *Flags) String() string {
	return fmt.Sprintf("Port: %d\r\nMaster Host: %s\r\nMaster Port: %d", f.port, f.master_host, f.master_port)
}

func parseRESP(data []byte) ([]string, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	return RESP(reader)
}

func handleCommand(raw_command []byte, flags Flags) ([]string, string) {
	command, err := parseRESP(raw_command)

	if err != nil {
		fmt.Println("Error parsing input: ", err.Error())
	}

	isSlave := flags.master_host != ""
	output := ""

	switch strings.ToLower(command[0]) {
	case "command":
		output = "+OK\r\n"
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
	case "info":
		if strings.ToLower(command[1]) == "replication" {
			item := ""
			if isSlave {
				item = "# Replication\r\nrole:slave\r\n"
			} else {
				item = "# Replication\r\nrole:master\r\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\r\nmaster_repl_offset:0\r\n"
			}
			output = fmt.Sprintf("$%d\r\n%s\r\n", len([]byte(item)), item)
		}
	case "replconf":
		if len(command) == 3 {
			if command[1] == "listening-port" {
				replicaPort, _ := strconv.ParseUint(command[2], 10, 64)
				fmt.Printf("Replica binded to port %d...\r\n", replicaPort)
				output = "+OK\r\n"
			} else if command[1] == "capa" && command[2] == "psync2" {
				output = "+OK\r\n"
			}
		}
	case "psync":
		if len(command) == 3 {
			// replid, replOffset := command[1], command[2]
			output = "+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n"
		}
	default:
		fmt.Printf("The command you gave: %q, isn't a valid one yet\r\n", command[0])
		os.Exit(1)
	}
	return command, output
}

func handleConnection(connection net.Conn, flags Flags) {
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

		command, output := handleCommand(raw_command, flags)

		_, err = connection.Write([]byte(output))

		if err != nil {
			fmt.Println("Error writing output: ", err.Error())
		}

		fmt.Printf("Received %d bytes: %q\n", numberOfBytes, command)
	}
}

func parseFlags(args []string) Flags {
	i := 0
	f := Flags{port: 6379, master_host: "", master_port: 6379}
	for i <= len(args)-1 {
		switch strings.ToLower(args[i]) {
		case "--port":
			i += 1
			p, _ := strconv.ParseUint(args[i], 10, 64)
			f.port = p
			i += 1
		case "--replicaof":
			i += 1
			f.master_host = args[i]
			i += 1
			p, _ := strconv.ParseUint(args[i], 10, 64)
			f.master_port = p
			i += 1
		default:
			i += 1
		}
	}

	return f
}

func handleHandshake(flags Flags) error {
	tcpaddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", flags.master_host, flags.master_port))
	if err != nil {
		return err
	}

	connection, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		return err
	}

	defer connection.Close()

	handshakeParts := [4]string{"*1\r\n$4\r\nPING\r\n", fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n%d\r\n", flags.port), "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n", "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"}

	for _, part := range handshakeParts {
		_, err = connection.Write([]byte(part))
		if err != nil {
			return err
		}

		buf := make([]byte, 1024)
		_, err = connection.Read(buf)
		if err != nil {
			return err
		}
	}

	fmt.Printf("%s\r\n", flags.String())
	return nil
}

func main() {
	flags := parseFlags(os.Args)

	isSlave := flags.master_host != ""
	if isSlave {
		err := handleHandshake(flags)
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
