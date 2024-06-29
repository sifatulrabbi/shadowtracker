package tcpinterceptor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
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
			go handleNewConn(&InterceptedConn{conn, forwardPort, false})
		}
	}()
	err = <-errCh
	close(errCh)
	return err
}

func handleNewConn(conn *InterceptedConn) {
	defer conn.TCPConn.Close()

	var (
		buf           = make([]byte, MB)
		forwardingSrv = fmt.Sprintf("http://127.0.0.1:%s", conn.HTTPForwardPort)
	)
	if _, err := conn.TCPConn.Read(buf); err != nil {
		writeToTCP(conn.TCPConn, []byte("failed to read incoming data"), []byte(err.Error()))
		return
	}

	parsedReq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if err != nil || parsedReq == nil {
		writeToTCP(conn.TCPConn, []byte("failed parse the HTTP request"), []byte(err.Error()))
		return
	}
	res, err := forwardHttpReq(parsedReq, forwardingSrv)
	if err != nil {
		writeToTCP(
			conn.TCPConn,
			[]byte(fmt.Sprintf("unable to forward the request to %s", forwardingSrv)),
			[]byte(err.Error()),
		)
		return
	}
	writeToTCP(conn.TCPConn, res)
}

func forwardHttpReq(parsedReq *http.Request, srvUrl string) ([]byte, error) {
	var resBuf bytes.Buffer
	newReq, err := http.NewRequest(
		parsedReq.Method,
		fmt.Sprintf("%s%s", srvUrl, parsedReq.RequestURI),
		parsedReq.Body,
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	resBuf.WriteString(fmt.Sprintf("HTTP/%d.%d %d %s\r\n",
		res.ProtoMajor, res.ProtoMinor, res.StatusCode, res.Status))
	for k, v := range res.Header {
		resBuf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ",")))
	}
	// TODO: add cookies to the response
	// for _, v := range res.Cookies() {
	// }
	resBuf.WriteString("\r\n")
	if res.Body != nil {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("Unable to read response body:", err)
		}
		if _, err = resBuf.Write(b); err != nil {
			return nil, err
		}
	}
	return resBuf.Bytes(), nil
}

func writeToTCP(conn *net.TCPConn, res ...[]byte) {
	for _, v := range res {
		if _, err := conn.Write(v); err != nil {
			log.Println("Unable to write response to TCP connection.", err)
			break
		}
	}
}
