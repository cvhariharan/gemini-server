package gemini

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
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

func FileServer(path string) Handler {
	return Handlerfunc(func(w *Response, r *Request) {
		filePath := r.URL.Path
		if !strings.HasPrefix(filePath, "/") {
			filePath = "/" + filePath
			r.URL.Path = filePath
		}
		f, err := os.Open(filepath.Join(path, filePath))
		if err != nil {
			log.Println(err)
			w.SetStatus(StatusPermanentFailure, "File not found")
			w.SendStatus()
		}
		defer f.Close()

		head := make([]byte, 261)
		f.Read(head)
		f.Seek(0, 0)
		kind, _ := filetype.Match(head)
		w.SetStatus(StatusSuccess, kind.MIME.Value)
		w.SendStatus()
		_, err = io.Copy(w.Body, f)
		if err != nil {
			log.Println(err)
			w.SetStatus(StatusPermanentFailure, "Could not read file")
			w.SendStatus()
		}
	})
}
