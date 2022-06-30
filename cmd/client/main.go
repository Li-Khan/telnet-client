package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Client struct {
	Host      string
	Port      string
	Timeout   time.Duration
	Dialer    *net.Dialer
	Conn      net.Conn
	Context   context.Context
	CancelCtx context.CancelFunc
}

func NewClient() *Client {
	host := flag.String("h", "127.0.0.1", "network")
	port := flag.String("p", "8080", "port")
	timeout := flag.Int("timeout", 10, "timeout for connecting to the server")

	flag.Parse()

	client := Client{
		Host:    *host,
		Port:    *port,
		Timeout: time.Duration(*timeout) * time.Second,
	}

	client.DialAndConnection()
	return &client
}

func (c *Client) DialAndConnection() error {
	dialer := &net.Dialer{}
	address := fmt.Sprintf("%s:%s", c.Host, c.Port)

	c.Context, c.CancelCtx = context.WithTimeout(context.Background(), c.Timeout)

	conn, err := dialer.DialContext(c.Context, "tcp", address)
	if err != nil {
		return err
	}

	c.Dialer = dialer
	c.Conn = conn

	return nil
}

func (c *Client) ReadRoutine() {
	scanner := bufio.NewScanner(c.Conn)

OUTER:
	for {
		select {
		case <-c.Context.Done():
			break OUTER
		default:
			if !scanner.Scan() {
				c.CancelCtx()
				break OUTER
			}
			text := scanner.Text()
			log.Printf("from server: %s\n", text)
		}
	}
	log.Printf("finished ReadRoutine")
}

func (c *Client) WriteRoutine() {
	scanner := bufio.NewScanner(os.Stdin)
OUTER:
	for {
		select {
		case <-c.Context.Done():
			c.CancelCtx()
			break OUTER
		default:
			if !scanner.Scan() {
				c.CancelCtx()
				break OUTER
			}
			text := scanner.Text()
			log.Printf("to server: %s\n", text)

			fmt.Fprintf(c.Conn, "%s\n", text)
		}
	}
	log.Printf("finished WriteRoutine")
}

func (c *Client) Start() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		c.ReadRoutine()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		c.WriteRoutine()
		wg.Done()
	}()

	wg.Wait()
	_ = c.Conn.Close()
}

func main() {
	client := NewClient()
	client.Start()
}
