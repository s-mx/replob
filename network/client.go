package network

import (
	"net"
	"encoding/gob"
	cont "github.com/s-mx/replob/containers"
	"sync"
	"log"
)

type ClientService struct {
	id             int
	service        string
	channelMessage chan cont.Message
	waitGroup      sync.WaitGroup
}

func NewClientService(id int, service string) *ClientService {
	return &ClientService{
		id:id,
		service:service,
		channelMessage:make(chan cont.Message, 10), // FIXME: use flags instead
	}
}

func (service *ClientService) start() {
	defer service.waitGroup.Done()

	for {
		var message cont.Message
		var more bool
		select {
		case message, more = <-service.channelMessage:
		default:
			continue
		}

		if more == false {
			log.Printf("INFO client server[%d]: stopping working\n", service.id)
			return
		}

		// FIXME: reuse connection
		// FIXME: implement reconnection and appropriate error handling
		conn, err := net.Dial("tcp", service.service)
		checkError(err)
		err = gob.NewEncoder(conn).Encode(message)
	}
}

func (service *ClientService) Start() {
	service.waitGroup.Add(1)
	go service.start()
}

func (service *ClientService) Stop() {
	close(service.channelMessage)
	service.waitGroup.Wait()
}

func NewClient(service string) net.Conn {
	conn, err := net.Dial("tcp", service)
	checkError(err)
	return conn
}
