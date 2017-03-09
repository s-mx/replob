package network

import (
	"log"
	"encoding/gob"
	"sync"
	"net"
	cont "github.com/s-mx/replob/containers"
	"time"
)

type ServerService struct {
	id             int
	nameServer     string
	service        string
	channelMessage chan cont.Message
	channelStop    chan interface{}
	waitGroup      sync.WaitGroup
}

func NewServerService(id int, config Configuration) *ServerService {
	return &ServerService{
		id:id,
		nameServer:config.nameServer,
		service:config.serviceServer,
		channelMessage:make(chan cont.Message, 10), // TODO: use flags
	}
}

func (service *ServerService) handleConnection(conn *net.TCPConn) {
	// FIXME: resuse connection in for {} loop
	defer conn.Close()
	defer service.waitGroup.Done()

	var message cont.Message
	// TODO: realize normal serialization
	err := gob.NewDecoder(conn).Decode(&message)
	checkError(err)
	service.channelMessage<-message
}

func (service *ServerService) Serve(listener *net.TCPListener) {
	defer service.waitGroup.Done()
	for {
		var more bool
		select {
		case _, more = <-service.channelStop:
		default:
		}

		if more == false {
			log.Printf("INFO server[%d]: stopping listening\n", service.id)
			listener.Close()
			return
		}

		listener.SetDeadline(time.Now().Add(500 * time.Microsecond))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			log.Printf("WARNING: %s", err)
		}

		conn.RemoteAddr()
		log.Printf("INFO server[%d]: %s\n", service.id, conn.RemoteAddr().String())
		service.waitGroup.Add(1)
		go service.handleConnection(conn)
	}
}

func (service *ServerService) Start() {
	laddr, err := net.ResolveTCPAddr("tcp", service.service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", laddr)
	log.Printf("INFO: server %s just started\n", service.nameServer)
	service.waitGroup.Add(1)
	go service.Serve(listener)
}

func (service *ServerService) Stop() {
	close(service.channelStop)
	service.waitGroup.Wait()
}

func (service* ServerService) logStart() {
	log.Printf("INFO: server %s just started\n", service.nameServer)
}

func checkError(err error) {
	if err != nil {
		// TODO: сделать конкретнее
		log.Panic(err.Error())
	}
}
