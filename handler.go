package gemini

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
)

type Request struct {
	URL *url.URL
}

type Response struct {
	Body       net.Conn
	Meta       string
	StatusCode int
	StatusText string
}

func (r *Response) Write(b []byte) (int, error) {
	err := r.SendStatus()
	if err != nil {
		log.Println(err)
		return -1, err
	}
	return r.Body.Write(b)
}

func (r *Response) SetStatus(statusCode int, statusText string) {
	r.StatusCode = statusCode
	r.StatusText = statusText
}

func (r *Response) SendStatus() error {
	if r.StatusText == "" {
		r.StatusCode = StatusSuccess
		r.StatusText = "text/gemini"
	}
	_, err := r.Body.Write([]byte(fmt.Sprintf("%d %s\r\n", r.StatusCode, r.StatusText)))
	return err
}

type Handler interface {
	ServeGemini(w *Response, r *Request)
}

type Handlerfunc func(*Response, *Request)

func (h Handlerfunc) ServeGemini(w *Response, r *Request) {
	h(w, r)
}

type Path struct {
	handler Handler
	path    string
}

type SimpleHandler struct {
	pathHandler []Path
}

var DefaultHandler = new(SimpleHandler)

func (s *SimpleHandler) ServeGemini(w *Response, r *Request) {
	u := r.URL.Path
	if u == "" {
		u = "/"
	}

	for _, h := range s.pathHandler {
		if strings.HasPrefix(u, h.path) {
			h.handler.ServeGemini(w, r)
			return
		}
	}
}

func HandleFunc(p string, h Handlerfunc) {
	DefaultHandler.pathHandler = append(DefaultHandler.pathHandler, Path{handler: h, path: p})
}

func ListenAndServeTLS(addr string, certFile, keyFile string) error {
	s := Server{Addr: addr, Handler: DefaultHandler}
	return s.ListenAndServeTLS(certFile, keyFile)
}
