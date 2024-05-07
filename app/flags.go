package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Flags struct {
	port        uint64
	master_host string
	master_port uint64
}

func (f *Flags) String() string {
	return fmt.Sprintf("Port: %d\r\nMaster Host: %s\r\nMaster Port: %d", f.port, f.master_host, f.master_port)
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
