package network

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"net"
	"os"
	"io"
	cont "github.com/s-mx/replob/containers"
)

func ReadBuffer(conn net.Conn) (bytes.Buffer, error) {
	defer conn.Close()

	result := bytes.Buffer{}
	var buf [512]byte // нужно еще уточнить число 512
	for {
		n, err := conn.Read(buf[0:])
		result.Write(buf[0:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return result, err
		}
	}

	return result, nil
}

func HandleClient(conn net.Conn) {
	buffer, err := ReadBuffer(conn)
	checkError(err)

	var message cont.Message
	decoder := gob.NewDecoder(&buffer)
	err = decoder.Decode(&message)
	checkError(err)

	fmt.Println(message.TypeMessage)
}

func RunLocalServer(service string) {
	listener, err := net.Listen("tcp", service)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		fmt.Println("Listen")
		// Кажется, нужно делать это ассинхроно
		HandleClient(conn)
	}
}

func NewLocalServer(service string) {
	go RunLocalServer(service)
}

func checkError(err error) {
	if err != nil {
		os.Exit(1)
	}
}