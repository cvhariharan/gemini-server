package gemini

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"net/url"
)

const (
	StatusInput              = 10
	StatusSuccess            = 20
	StatusRedirect           = 30
	StatusTemporaryFailure   = 40
	StatusPermanentFailure   = 50
	StatusClientCertRequired = 60
)

type Server struct {
	Addr    string
	Handler Handler
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", s.Addr, config)
	if err != nil {
		return err
	}
	defer listener.Close()

	return s.serveGemini(listener)
}

func (s *Server) serveGemini(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.geminiHandler(conn)
	}
}

func (s *Server) geminiHandler(conn net.Conn) {
	req := Request{}
	resp := Response{Body: conn}

	defer conn.Close()

	reader := bufio.NewReaderSize(conn, 1024)
	request, more, err := reader.ReadLine()
	if more {
		resp.SetStatus(StatusPermanentFailure, "Request size more than 1024 bytes")
		resp.SendStatus()
		return
	} else if err != nil {
		resp.SetStatus(StatusTemporaryFailure, "Unknown error while reading request")
		resp.SendStatus()
		return
	}

	u, err := url.Parse(string(request))
	if err != nil {
		resp.SetStatus(StatusPermanentFailure, "Unknown error while reading request")
		resp.SendStatus()
		log.Println(err)
		return
	}

	if u.Scheme == "" {
		u.Scheme = "gemini"
	}

	if u.Scheme != "gemini" {
		resp.SetStatus(StatusPermanentFailure, "Unsupported protocol")
		resp.SendStatus()
		return
	}
	req.URL = u

	s.Handler.ServeGemini(&resp, &req)
}
