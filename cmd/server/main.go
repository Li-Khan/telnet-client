package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	Mutex sync.Mutex
	Conn  map[net.Conn]string
}

const address string = "127.0.0.1:8080"

func (s *Server) handleConnection(conn net.Conn) {
	s.Mutex.Lock()
	s.Conn[conn] = conn.RemoteAddr().String()
	s.Mutex.Unlock()

	defer func() {
		s.Mutex.Lock()
		delete(s.Conn, conn)
		s.Mutex.Unlock()

		_ = conn.Close()
	}()

	fmt.Fprintf(conn, "Welcome, %s\n", conn.RemoteAddr())
	s.printMessage("Connected", conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		log.Printf("%s: %s", conn.RemoteAddr(), text)
		if text == "exit" {
			break
		}

		s.printMessage(text, conn)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	log.Printf("Closing connection with %s", conn.RemoteAddr())
	s.printMessage("Disconnected", conn)
}

func (s *Server) printMessage(msg string, conn net.Conn) {
	for c := range s.Conn {
		if c != conn {
			fmt.Fprintf(c, "%s: %s\n", conn.RemoteAddr(), msg)
		}
	}
}

func start(l net.Listener) error {
	server := Server{
		Conn: make(map[net.Conn]string),
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go server.handleConnection(conn)
	}
}

func main() {
	listener, err := net.Listen("tcp", address)
	defer func() {
		_ = listener.Close()
	}()
	if err != nil {
		log.Println(err)
	}

	log.Printf("starting server on: %s", address)
	err = start(listener)
	if err != nil {
		log.Println(err)
		return
	}
}
