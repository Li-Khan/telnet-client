package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

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

func main() {
	l, err := net.Listen("tcp", ":8080")
	defer func() {
		_ = l.Close()
	}()
	if err != nil {
		log.Println(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleConnection(conn)
	}
}
