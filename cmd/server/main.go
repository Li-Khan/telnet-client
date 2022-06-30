package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const address string = "127.0.0.1:8080"

func handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	fmt.Fprintf(conn, "Welcome, %s\n", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		log.Printf("%s: %s", conn.RemoteAddr(), text)
		if text == "exit" {
			break
		}

		fmt.Fprintf(conn, "I have recieved '%s'\n", text)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	log.Printf("Closing connection with %s", conn.RemoteAddr())
}

func start(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn)
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
