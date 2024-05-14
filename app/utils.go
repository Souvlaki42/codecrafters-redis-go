package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func StringExistsAndIndex(slice []string, target string, case_sensitive bool) (bool, int) {
	var t, s string
	for i, rs := range slice {
		if !case_sensitive {
			t = strings.ToLower(target)
			s = strings.ToLower(rs)
		} else {
			t = target
			s = rs
		}
		if s == t {
			return true, i
		}
	}
	return false, -1
}

type Flags struct {
	port        uint64
	master_host string
	master_port uint64
	is_master   bool
}

func (f *Flags) String() string {
	return fmt.Sprintf("Port: %d\r\nMaster Host: %s\r\nMaster Port: %d\r\nMaster: %t\r\n", f.port, f.master_host, f.master_port, f.is_master)
}

func parseFlags() Flags {
	f := Flags{port: 6379, master_host: "localhost", master_port: 6379, is_master: true}

	exists, index := StringExistsAndIndex(os.Args, "--port", false)
	if len(os.Args) >= index+2 && exists {
		port, err := strconv.ParseUint(os.Args[index+1], 10, 64)
		if err != nil {
			fmt.Println("You are supposed to provide server's port as a positive integer.")
			os.Exit(1)
		}
		f.port = port
	}

	exists, index = StringExistsAndIndex(os.Args, "--replicaof", false)
	if len(os.Args) >= index+3 && exists {
		port, err := strconv.ParseUint(os.Args[index+2], 10, 64)
		if err != nil {
			fmt.Println("You are supposed to provide master's port as a positive integer.")
			os.Exit(1)
		}
		f.is_master = false
		f.master_host = os.Args[index+1]
		f.master_port = port
	}

	return f
}
