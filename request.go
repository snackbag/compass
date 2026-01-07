package compass

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	if resp.internalError {
		return errors.New(string(resp.Body))
	}

	for key, value := range resp.Headers {
		if strings.HasPrefix(key, "--COMPASS") {
			continue
		}

		w.Header().Set(key, value)
	}

	if resp.ContentType != nil {
		switch *resp.ContentType {
		case "--COMPASS-redirect":
			{
				http.Redirect(w, r.Http, string(resp.Body), resp.StatusCode)
				s.Logger.Request(r.Http, resp.StatusCode)
				return nil
			}
		case "--COMPASS-serve":
			{
				rs := bytes.NewReader(resp.Body)
				http.ServeContent(w, r.Http, resp.Headers["-Compass-File-Name"], time.Now(), rs)
				s.Logger.Request(r.Http, resp.StatusCode)
				return nil
			}
		}

		if *resp.ContentType == "--COMPASS-redirect" {
			http.Redirect(w, r.Http, string(resp.Body), resp.StatusCode)
			return nil
		}

		w.Header().Set("Content-Type", *resp.ContentType)
	} else {
		w.Header().Set("Content-Type", r.Http.Header.Get("Content-Type"))
	}
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
	return s.write(w, r.Http, []byte(fmt.Sprintf("<html><h1>Not Found</h1><p>The requested route %s was not found on this server.</p></html>", r.URL.Path)), http.StatusNotFound)
}

// GetRouteParam gets the requested content of the requested id. If nothing was found, it will return an empty string.
func (r *Request) GetRouteParam(id string) string {
	if len(id) < 1 {
		return ""
	}

	index, ok := r.Route.partIdMap[id]
	if !ok {
		return ""
	}

	split := splitUrlPath(r.URL.Path)
	if index > len(split)-1 {
		return ""
	}

	part := r.Route.parts[index]

	value := split[index]
	value = strings.TrimPrefix(value, part.prefix)
	value = strings.TrimSuffix(value, part.suffix)

	return value
}
