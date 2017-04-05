package network

import (
	"log"
	"encoding/gob"
	"sync"
	"net"
	cont "github.com/s-mx/replob/containers"
	"time"
	"io"
	"sync/atomic"
)

const (
	STOPPED = iota
	RUNNING
)

type ServerService struct {
	id					int
	service				string
	channelMessage		chan cont.Message
	waitGroup			sync.WaitGroup

	numberClients		int

	isRunning			*int32
}

func NewServerService(id int, config *Configuration) *ServerService {
	tmpInt := int32(0)
	return &ServerService{
		id:id,
		service:config.serviceServer[id],
		channelMessage:make(chan cont.Message, 10), // TODO: use flags
		numberClients:0,
		isRunning:&tmpInt,
	}
}

func (service *ServerService) handleConnection(id int, conn *net.TCPConn) {
	defer func() {
		log.Printf("Client [%d]: stopping working", service.id)
		conn.Close()
		service.waitGroup.Done()
	}()

	for {
		if atomic.LoadInt32(service.isRunning) == STOPPED {
			log.Printf("Handler [%d]: stop working", service.id)
			return
		}

		conn.SetDeadline(time.Now().Add(500 * time.Microsecond))

		var message cont.Message
		var err error
		err = gob.NewDecoder(conn).Decode(&message)
		if err == io.EOF {
			log.Printf("Server [%d]: EOF has occured", service.id)
			return
		}

		if Err, ok := err.(net.Error); ok {
			if Err.Timeout() {
				continue
			}

			log.Printf("Network Error: %s", err.Error())
			return
		}

		checkError(err)

		log.Printf("Server [%d]: Message received", service.id)
		select {
		case service.channelMessage<-message:
			break
		default:
			log.Printf("Server [%d]: The message is lost", service.id)
		}
	}
}

func (service *ServerService) goHandler(id int, conn *net.TCPConn) {
	service.waitGroup.Add(1)
	go service.handleConnection(id, conn)
}

func (service *ServerService) Serve(listener *net.TCPListener) {
	defer func(){
		service.waitGroup.Done()
		listener.Close()
	}()

	for {
		if atomic.LoadInt32(service.isRunning) == STOPPED {
			log.Printf("Server [%d]: waiting for handlers", service.id)
			return
		}

		listener.SetDeadline(time.Now().Add(5 * time.Second)) // FIXME: use flags
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			log.Printf("WARNING: %s", err)
			continue
		}

		conn.RemoteAddr()
		log.Printf("INFO server[%d]: %s\n", service.id, conn.RemoteAddr().String())
		service.goHandler(service.numberClients, conn)
		service.numberClients += 1
	}
}

func (service *ServerService) Start() {
	if atomic.LoadInt32(service.isRunning) == RUNNING {
		return
	}

	atomic.StoreInt32(service.isRunning, RUNNING)

	laddr, err := net.ResolveTCPAddr("tcp", service.service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", laddr)
	checkError(err)
	log.Printf("INFO: server [%d] just started\n", service.id)
	service.waitGroup.Add(1)
	go service.Serve(listener)
}

func (service *ServerService) Stop() {

	if atomic.LoadInt32(service.isRunning) == STOPPED {
		return
	}

	atomic.StoreInt32(service.isRunning, STOPPED)
	service.waitGroup.Wait()
	log.Printf("Server [%d]: stop working", service.id)
}

func checkError(err error) {
	if err != nil {
		// TODO: сделать конкретнее
		log.Panic(err.Error())
	}
}
