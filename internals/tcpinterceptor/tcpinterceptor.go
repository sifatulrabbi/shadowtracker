package tcpinterceptor

import (
	"fmt"
	"log"
	"net"
)

type InterceptedConn struct {
	*net.TCPConn
	HTTPForwardPort string
	HTTPS           bool
}

func NewTCPListener(targetPort string, forwardPort string) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%s", targetPort))
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Println(err)
				continue
			}
			go handleNewConn(&InterceptedConn{conn, forwardPort, false}, errCh)
		}
	}()
	err = <-errCh
	close(errCh)
	return err
}

func handleNewConn(conn *InterceptedConn, errCh chan error) {
	defer conn.TCPConn.Close()
	buf := make([]byte, 1024)
	if _, err := conn.TCPConn.Read(buf); err != nil {
		errCh <- err
		return
	}

	fmt.Println(string(buf))
}
