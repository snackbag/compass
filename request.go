package compass

import (
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	Method string
	URL    *url.URL
	Route  *Route

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
	if r.Route == nil {
		return s.handleNotFound(w, r)
	}

	resp := r.Route.handler(r)

	w.WriteHeader(resp.StatusCode)
	_, err := w.Write(resp.Body)
	if err != nil {
		return err
	}

	s.Logger.Request(r.Http, resp.StatusCode)
	return nil
}

// TODO add customizability
func (s *Server) handleNotFound(w http.ResponseWriter, r Request) error {
	return s.write(w, r.Http, []byte("Your requested page does not exist"), http.StatusNotFound)
}
