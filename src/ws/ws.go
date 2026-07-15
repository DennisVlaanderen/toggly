package ws

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
)

func HandleRawTCP(conn net.Conn) error {
	defer conn.Close()

	log.Printf("new raw tcp client connected: %s", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	for {
		payload, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if _, err := conn.Write(payload); err != nil {
			return err
		}
	}
}

func StartTCPServer(ctx context.Context, addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("raw tcp server listening on %s", addr)

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("accept error: %v", err)
			continue
		}

		go func(client net.Conn) {
			if err := HandleRawTCP(client); err != nil && err != io.EOF {
				log.Printf("raw tcp handler error: %v", err)
			}
		}(conn)
	}
}
