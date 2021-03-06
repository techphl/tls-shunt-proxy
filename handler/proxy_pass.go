package handler

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type ProxyPassHandler struct {
	target string
}

func NewProxyPassHandler(target string) *ProxyPassHandler {
	return &ProxyPassHandler{target: target}
}

func (h *ProxyPassHandler) Handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	var err error

	var dstConn net.Conn
	if strings.HasPrefix(h.target, "unix:") {
		dstConn, err = net.Dial("unix", h.target[5:])
	} else {
		dstConn, err = net.Dial("tcp", h.target)
	}
	if err != nil {
		log.Printf("fail to connect to %s :%v\n", h.target, err)
		return
	}
	defer func() { _ = dstConn.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	go func(srcConn net.Conn, dstConn net.Conn) {
		_, err := io.Copy(dstConn, srcConn)
		if err != nil && err != io.EOF {
			log.Printf("failed to send to %s:%v\n", h.target, err)
		}
		wg.Done()
	}(conn, dstConn)
	go func(srcConn net.Conn, dstConn net.Conn) {
		_, err := io.Copy(srcConn, dstConn)
		if err != nil && err != io.EOF {
			log.Printf("failed to read from %s: %v\n", h.target, err)
		}
		wg.Done()
	}(conn, dstConn)

	wg.Wait()
}
