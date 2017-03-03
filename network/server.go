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
	id				int
	nameServer		string
	service			string
	channelMessage	chan cont.Message
	channelStop		chan int
	semaphore		sync.WaitGroup
}

func NewServerService(id int, config Configuration) *ServerService {
	return &ServerService{
		id:id,
		nameServer:config.nameServer,
		service:config.serviceServer,
		channelMessage:make(chan cont.Message, 1),
	}
}

func (service *ServerService) handleMessage(conn *net.TCPConn) {
	defer conn.Close()
	defer service.semaphore.Done()

	var message cont.Message
	// TODO: realize normal serialization
	err := gob.NewDecoder(conn).Decode(&message)
	checkError(err)
	service.channelMessage<-message
}

func (service *ServerService) Serve(listener *net.TCPListener) {
	defer service.semaphore.Done()
	for {
		select {
		case <-service.channelStop:
			log.Printf("INFO server[%d]: stopping listening\n", service.id)
			listener.Close()
			return
		default:
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
		service.semaphore.Add(1)
		go service.handleMessage(conn)
	}
}

func (service *ServerService) Start() {
	laddr, err := net.ResolveTCPAddr("tcp", service.service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", laddr)
	service.logStart()
	service.semaphore.Add(1)
	go service.Serve(listener)
}

func (service *ServerService) Stop() {
	service.channelStop<-0
	service.semaphore.Wait()
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
