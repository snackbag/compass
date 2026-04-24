package compass

import (
	"bytes"
	"errors"
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

// NewRequestFromHttp constructs a Request from a standard http.Request.
//
// The HTTP method is normalized to lowercase. The Route field is not
// populated and must be assigned later during routing.
func NewRequestFromHttp(r *http.Request) Request {
	return Request{
		Method: strings.ToLower(r.Method),
		URL:    r.URL,

		Http: r,
	}
}

// writeResponse writes the headers, content type, status code, and body
// of a Response to the client.
//
// Headers prefixed with "--COMPASS" are skipped. This is the shared
// path for all standard responses.
func (s *Server) writeResponse(w http.ResponseWriter, r Request, resp Response) error {
	s.writeCookies(w, resp.cookies)

	for key, value := range resp.Headers {
		if strings.HasPrefix(key, "--COMPASS") {
			continue
		}
		w.Header().Set(key, value)
	}

	if resp.ContentType != nil {
		w.Header().Set("Content-Type", *resp.ContentType)
	} else {
		w.Header().Set("Content-Type", r.Http.Header.Get("Content-Type"))
	}

	return s.write(w, r.Http, resp.Body, resp.StatusCode)
}

// handleRequest processes an incoming Request and writes the response.
//
// If no route is matched, it delegates to handleNotFound. Otherwise,
// it executes the route handler and writes the resulting response,
// including headers, status code, and body.
//
// Special internal ContentType values control behavior:
//
//	"--COMPASS-redirect": performs an HTTP redirect
//	"--COMPASS-serve": serves content as a file
//
// Headers prefixed with "--COMPASS" are ignored. All successful
// responses are logged. If the handler signals an internal error,
// it is returned.
func (s *Server) handleRequest(w http.ResponseWriter, r Request) error {
	if r.Route == nil {
		return s.handleNotFound(w, r)
	}

	resp := r.Route.handler(r)
	if resp.internalError {
		return errors.New(string(resp.Body))
	}

	if resp.ContentType != nil {
		switch *resp.ContentType {
		case "--COMPASS-redirect":
			s.writeCookies(w, resp.cookies)
			http.Redirect(w, r.Http, string(resp.Body), resp.StatusCode)
			s.Logger.Request(r.Http, resp.StatusCode)
			return nil
		case "--COMPASS-serve":
			rs := bytes.NewReader(resp.Body)
			s.writeCookies(w, resp.cookies)
			http.ServeContent(w, r.Http, resp.Headers["-Compass-File-Name"], time.Now(), rs)
			s.Logger.Request(r.Http, resp.StatusCode)
			return nil
		}
	}

	return s.writeResponse(w, r, resp)
}

// handleNotFound writes a default 404 HTML response.
//
// The response contains a simple HTML page indicating that the
// requested route was not found. This implementation is currently
// not customizable.
func (s *Server) handleNotFound(w http.ResponseWriter, r Request) error {
	resp := s.NotFoundHandler(r)
	return s.writeResponse(w, r, resp)
}

// GetRouteParam returns the value of a named route parameter.
//
// The parameter is resolved using the route's internal mapping and
// extracted from the URL path. Any defined prefix or suffix on the
// route part is removed before returning the value.
//
// If the parameter does not exist, the route is not set, or the index
// is out of bounds, an empty string and false are returned.
func (r *Request) GetRouteParam(id string) (string, bool) {
	if len(id) < 1 {
		return "", false
	}

	index, ok := r.Route.partIdMap[id]
	if !ok {
		return "", false
	}

	split := splitUrlPath(r.URL.Path)
	if index > len(split)-1 {
		return "", false
	}

	part := r.Route.parts[index]

	value := split[index]
	value = strings.TrimPrefix(value, part.prefix)
	value = strings.TrimSuffix(value, part.suffix)

	return value, true
}

// GetCookie returns the value of the named cookie from the incoming request.
//
// The second return value is false if no cookie with that name was sent.
// Note that incoming cookies carry only name and value and no attributes
// like Path, Expires or HttpOnly.
func (r *Request) GetCookie(name string) (string, bool) {
	c, err := r.Http.Cookie(name)
	if err != nil {
		return "", false
	}
	return c.Value, true
}

// GetCookies returns all cookies sent with the request as a name-value map.
//
// If the same cookie name appears more than once, the last value wins.
// Note that incoming cookies carry only name and value and no attributes
// like Path, Expires or HttpOnly.
func (r *Request) GetCookies() map[string]string {
	result := make(map[string]string)
	for _, c := range r.Http.Cookies() {
		result[c.Name] = c.Value
	}
	return result
}
