package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func handleCommand(raw_command []byte, flags Flags) ([]string, string) {
	command, err := parseRESP(raw_command)

	if err != nil {
		fmt.Println("Error parsing input: ", err.Error())
	}

	output := ""

	switch strings.ToLower(command[0]) {
	// TODO: Command that logs resp command
	case "command":
		output = "+OK\r\n"
	case "ping":
		output = "+PONG\r\n"
	case "echo":
		output = fmt.Sprintf("+%s\r\n", strings.Join(command[1:], " "))
	case "set":
		mutex.Lock()
		defer mutex.Unlock()

		if flags.is_master {
			for _, port := range replica_ports {
				err := handleHandshake(fmt.Sprintf("0.0.0.0:%d", port), []string{string(raw_command)}, false)
				if err != nil {
					fmt.Println("Error replicating commands: ", err.Error())
				}
			}
		}

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
			if !flags.is_master {
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
				replica_ports = append(replica_ports, replicaPort)
				output = "+OK\r\n"
			} else if command[1] == "capa" && command[2] == "psync2" {
				output = "+OK\r\n"
			}
		}
	case "psync":
		if len(command) == 3 {
			// replid, replOffset := command[1], command[2]

			emptyRDBFileBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
			emptyRDBFileBytes, err := base64.StdEncoding.DecodeString(emptyRDBFileBase64)
			if err != nil {
				fmt.Println("Failed to decode the empty rdb file to bytes")
				os.Exit(1)
			}
			output = fmt.Sprintf("+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n$%d\r\n%s", len(emptyRDBFileBytes), emptyRDBFileBytes)
		}
	default:
		fmt.Printf("The command you gave: %q, isn't a valid one yet\r\n", command[0])
		os.Exit(1)
	}
	return command, output
}
