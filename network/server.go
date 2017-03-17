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

type ServerService struct {
	id					int
	service				string
	channelMessage		chan cont.Message
	channelStop			chan interface{}
	waitGroup			sync.WaitGroup
	waitGroupHandlers	sync.WaitGroup

	numberClients		int
	flagClose			*int32

	isRunning			bool
	mutexRunning		sync.Mutex
}

func NewServerService(id int, config *Configuration) *ServerService {
	tmpFlagClose := int32(0)
	return &ServerService{
		id:id,
		service:config.serviceServer[id],
		channelMessage:make(chan cont.Message, 10), // TODO: use flags
		channelStop:make(chan interface{}),
		numberClients:0,
		flagClose:&tmpFlagClose,
	}
}

func (service *ServerService) handleConnection(id int, conn *net.TCPConn) {
	defer func() {
		log.Printf("Client [%d]: stopping working", service.id)
		conn.Close()
		service.waitGroupHandlers.Done()
		service.waitGroup.Done()
	}()

	for {
		if atomic.LoadInt32(service.flagClose) != 0 {
			log.Printf("Handler [%d]: stop working", service.id)
			return
		}

		conn.SetDeadline(time.Now().Add(500 * time.Microsecond))

		var message cont.Message
		var err error
		err = gob.NewDecoder(conn).Decode(&message) // FIXME: consider error checking here
		if err == io.EOF {
			// FIXME: add info log
			log.Printf("EOF")
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

		select {
		case service.channelMessage<-message:
			log.Printf("Server [%d]: Message sent", service.id)
			break
		// FIXME: consider removing timeout
		case <-time.After(time.Second): // FIXME: use flags instead
			log.Printf("Server [%d]: The message is lost", service.id)
			break
		}
	}
}

func (service *ServerService) goHandler(id int, conn *net.TCPConn) {
	service.waitGroup.Add(1)
	service.waitGroupHandlers.Add(1)
	go service.handleConnection(id, conn)
}

func (service *ServerService) Serve(listener *net.TCPListener) {
	defer func(){
		service.waitGroup.Done()
		listener.Close()
	}()

	for {
		if atomic.LoadInt32(service.flagClose) != 0 {
			log.Printf("Server [%d]: waiting for handlers", service.id)
			service.waitGroupHandlers.Wait()
			return
		}

		listener.SetDeadline(time.Now().Add(500 * time.Microsecond)) // FIXME: use flags
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			log.Printf("WARNING: %s", err)
			// FIXME: add handling
		}

		conn.RemoteAddr()
		log.Printf("INFO server[%d]: %s\n", service.id, conn.RemoteAddr().String())
		service.goHandler(service.numberClients, conn)
		service.numberClients += 1
	}
}

func (service *ServerService) Start() {
	defer service.mutexRunning.Unlock()
	service.mutexRunning.Lock()

	if service.isRunning {
		return
	}

	service.isRunning = true

	laddr, err := net.ResolveTCPAddr("tcp", service.service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", laddr)
	checkError(err)
	log.Printf("INFO: server [%d] just started\n", service.id)
	service.waitGroup.Add(1)
	go service.Serve(listener)
}

func (service *ServerService) Stop() {
	defer service.mutexRunning.Unlock()
	service.mutexRunning.Lock()

	if service.isRunning == false {
		return
	}

	// FIXME: Combine isRunning and flagClose as single atomic variable
	service.isRunning = false
	atomic.StoreInt32(service.flagClose, 1)
	service.waitGroup.Wait()
	log.Printf("Server [%d]: stop working", service.id)
}

func checkError(err error) {
	if err != nil {
		// TODO: сделать конкретнее
		log.Panic(err.Error())
	}
}
