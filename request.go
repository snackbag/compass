package compass

import (
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	Method string
	URL    *url.URL

	Http *http.Request
}

func NewRequestFromHttp(r *http.Request) Request {
	return Request{
		Method: strings.ToLower(r.Method),
		URL:    r.URL,

		Http: r,
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r Request) error {
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return err
	}

	s.Logger.Request(r.Http, 200)
	return nil
}

// TODO add customizability
func (s *Server) handleNotFound(w http.ResponseWriter, r Request) error {
	return s.write(w, r.Http, []byte("Your requested page does not exist"), http.StatusNotFound)
}
