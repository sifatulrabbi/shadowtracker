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
	"time"

	"github.com/sifatulrabbi/shadowtracker/internals/logswriter"
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
	logger          *logswriter.Logger
}

func NewTCPListener(targetPort string, forwardPort string, logger *logswriter.Logger) error {
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
			go handleNewConn(&InterceptedConn{conn, forwardPort, false, logger})
		}
	}()
	err = <-errCh
	close(errCh)
	return err
}

func handleNewConn(conn *InterceptedConn) {
	var (
		buf           = make([]byte, MB)
		forwardingSrv = fmt.Sprintf("http://127.0.0.1:%s", conn.HTTPForwardPort)
		start         = time.Now()
		localLog      = logswriter.Log{ServerIP: forwardingSrv, CreatedAt: start}
	)

	defer conn.TCPConn.Close()
	defer conn.logger.WriteHTTPLog(&localLog, start)

	if _, err := conn.TCPConn.Read(buf); err != nil {
		writeToTCP(conn.TCPConn, []byte("failed to read incoming data"), []byte(err.Error()))
		localLog.Err = err
		return
	}
	localLog.BodySize = len(buf)

	parsedReq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if err != nil || parsedReq == nil {
		writeToTCP(conn.TCPConn, []byte("failed parse the HTTP request"), []byte(err.Error()))
		localLog.Err = err
		return
	}
	localLog.HTTPMethod = parsedReq.Method
	localLog.ReqPath = parsedReq.URL.RequestURI()
	localLog.HTTPVersion = fmt.Sprintf("%d.%d", parsedReq.ProtoMajor, parsedReq.ProtoMinor)
	localLog.ClientIP = parsedReq.Header.Get("x-real-ip")
	if localLog.ClientIP == "" {
		localLog.ClientIP = parsedReq.Header.Get("x-forwared-by")
	}
	if localLog.ClientIP == "" {
		localLog.ClientIP = parsedReq.RemoteAddr
	}

	res, err := forwardHttpReq(parsedReq, forwardingSrv, &localLog)
	if err != nil {
		writeToTCP(
			conn.TCPConn,
			[]byte(fmt.Sprintf("unable to forward the request to %s", forwardingSrv)),
			[]byte(err.Error()),
		)
		localLog.Err = err
		return
	}
	writeToTCP(conn.TCPConn, res)
}

func forwardHttpReq(parsedReq *http.Request, srvUrl string, localLog *logswriter.Log) ([]byte, error) {
	var resBuf bytes.Buffer
	newReq, err := http.NewRequest(
		parsedReq.Method,
		fmt.Sprintf("%s%s", srvUrl, parsedReq.RequestURI),
		parsedReq.Body,
	)
	if err != nil {
		localLog.Err = err
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
		localLog.Err = err
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
			localLog.Err = err
			log.Println("Unable to read response body:", err)
		}
		if _, err = resBuf.Write(b); err != nil {
			localLog.Err = err
			return nil, err
		}
	}

	localLog.ResponseSize = resBuf.Len()
	localLog.ResponseCode = res.StatusCode

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
