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

type Flag struct {
	Host    string
	Port    string
	Timeout time.Duration
}

type Client struct {
	Flag      Flag
	Dialer    *net.Dialer
	Conn      net.Conn
	Context   context.Context
	CancelCtx context.CancelFunc
}

func NewFlag() Flag {
	host := flag.String("h", "127.0.0.1", "network")
	port := flag.String("p", "8080", "port")
	timeout := flag.Int("timeout", 10, "timeout for connecting to the server")

	flag.Parse()

	flag := Flag{
		Host:    *host,
		Port:    *port,
		Timeout: time.Duration(*timeout) * time.Second,
	}

	return flag
}

func NewClient(flag Flag) (*Client, error) {
	client := Client{
		Flag: flag,
	}

	address := fmt.Sprintf("%s:%s", flag.Host, flag.Port)
	log.Printf("connecting to '%s'", address)
	err := client.Dial()
	if err != nil {
		return nil, err
	}
	log.Printf("connected '%s'", address)

	return &client, nil
}

func (c *Client) Dial() error {
	dialer := &net.Dialer{}
	address := fmt.Sprintf("%s:%s", c.Flag.Host, c.Flag.Port)

	c.Context, c.CancelCtx = context.WithTimeout(context.Background(), c.Flag.Timeout)

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
			c.CancelCtx()
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
	flag := NewFlag()
	client, err := NewClient(flag)
	if err != nil {
		log.Println(err)
		return
	}

	client.Start()
}
