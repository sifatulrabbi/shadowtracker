package tcpinterceptor

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
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

	bufReader := bufio.NewReader(bytes.NewReader(buf))
	parsedReq, err := http.ReadRequest(bufReader)
	if err != nil {
		errCh <- err
		return
	}
	if parsedReq == nil {
		errCh <- fmt.Errorf("failed to parse the incoming http request\n")
		return
	}

	newReq, err := http.NewRequest(parsedReq.Method, fmt.Sprintf("127.0.0.1:%s%s", conn.HTTPForwardPort, parsedReq.RequestURI), parsedReq.Body)
	if err != nil {
		errCh <- err
		return
	}
	for _, c := range parsedReq.Cookies() {
		newReq.AddCookie(c)
	}
	for k, values := range parsedReq.Header {
		for _, v := range values {
			newReq.Header.Add(k, v)
		}
	}

	res, err := http.DefaultClient.Do(newReq)
	if err != nil {
		errCh <- err
		return
	}
}
