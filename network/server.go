package network

import (
	"log"
	"encoding/gob"
	"sync"
	"net"
	cont "github.com/s-mx/replob/containers"
	"time"
	"io"
	"bytes"
)

type TwoWayChannel struct {
	forward, backward	chan interface{}
}

func NewTwoWayChannel() TwoWayChannel {
	return TwoWayChannel{
		forward:make(chan interface{}),
		backward:make(chan interface{}),
	}
}

type ServerService struct {
	id				int
	service			string
	channelMessage	chan cont.Message
	channelStop		chan interface{}
	waitGroup		sync.WaitGroup

	channelsToClients	map[int]TwoWayChannel

	isRunning		bool
	mutexRunning	sync.Mutex
}

func NewServerService(id int, config *Configuration) *ServerService {
	return &ServerService{
		id:id,
		service:config.serviceServer[id],
		channelMessage:make(chan cont.Message, 10), // TODO: use flags
		channelStop:make(chan interface{}),
	}
}

func (service *ServerService) handleConnection(id int, channels TwoWayChannel, conn *net.TCPConn) {
	defer conn.Close()
	defer close(channels.backward) // Проверить
	defer service.waitGroup.Done()

	for {
		select {
		case <-channels.forward:
			return
		default:
		}

		conn.SetDeadline(time.Now().Add(500 * time.Microsecond))

		var message cont.Message
		// TODO: realize normal serialization
		var buffer bytes.Buffer
		_, err := conn.Read(buffer.Bytes())
		_ = gob.NewDecoder(&buffer).Decode(&message)
		if err == io.EOF {
			return
		}

		if Err, ok := err.(net.Error); ok {
			if Err.Timeout() {
				continue
			}

			log.Printf("Network Error")
			channels.backward<-0
			return
		}

		service.channelMessage <- message
	}
}

func (service *ServerService) Serve(listener *net.TCPListener) {
	defer service.waitGroup.Done()

	numberClient := 0
	for {
		select {
		case _ = <-service.channelStop: // Нужно остановить хендлеры
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
		service.waitGroup.Add(1)
		channel := NewTwoWayChannel()
		service.channelsToClients[numberClient] = channel
		go service.handleConnection(numberClient, channel, conn)
		numberClient += 1
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

	service.isRunning = false
	close(service.channelStop)
	service.waitGroup.Wait()
}

func checkError(err error) {
	if err != nil {
		// TODO: сделать конкретнее
		log.Panic(err.Error())
	}
}
