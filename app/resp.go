package main

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

func RESP(reader *bufio.Reader) ([]string, error) {
	firstByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch firstByte {
	case '+':
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return []string{strings.TrimSpace(line)}, nil

	case '-':
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return nil, errors.New(strings.TrimSpace(line))

	case ':':
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return []string{strings.TrimSpace(line)}, nil

	case '$':
		lengthLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(strings.TrimSpace(lengthLine))
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return []string{""}, nil
		}
		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			return nil, err
		}
		reader.ReadString('\n') // Read the trailing newline
		return []string{string(data)}, nil

	case '*':
		lengthLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(strings.TrimSpace(lengthLine))
		if err != nil {
			return nil, err
		}
		var array []string
		for i := 0; i < length; i++ {
			element, err := RESP(reader)
			if err != nil {
				return nil, err
			}
			array = append(array, element...)
		}
		return array, nil

	default:
		return nil, errors.New("invalid message type")
	}
}
