package network

import (
	"net"
	"bytes"
	"encoding/gob"
	cont "github.com/s-mx/replob/containers"
)

func NewClient(service string) net.Conn {
	conn, err := net.Dial("tcp", service)
	checkError(err)
	return conn
}

func SendMessage(service string, message cont.Message) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(message)
	checkError(err)

	conn := NewClient(service)
	_, err = conn.Write(buffer.Bytes())
	checkError(err)
	conn.Close()
}
