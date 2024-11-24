package compass

import (
	"net/http"
	"net/url"
	"slices"
)

type Request struct {
	Method      string
	IP          string
	URL         url.URL
	UserAgent   string
	FullRequest http.Request
}

func NewRequest(r http.Request) Request {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}

	return Request{
		Method:      r.Method,
		IP:          r.RemoteAddr,
		URL:         *r.URL,
		UserAgent:   r.UserAgent(),
		FullRequest: r,
	}
}

type Response struct {
	IsRedirect bool
	Code       int
	Content    string

	cookies []http.Cookie
}

func NewResponse(isRedirect bool, code int, content string) Response {
	return Response{IsRedirect: isRedirect, Code: code, Content: content, cookies: make([]http.Cookie, 0)}
}

func Redirect(target string) Response {
	return NewResponse(true, 307, target)
}

func RedirectWithCode(target string, code int) Response {
	return NewResponse(true, code, target)
}

func Text(content string) Response {
	return NewResponse(false, 200, content)
}

func TextWithCode(content string, code int) Response {
	return NewResponse(false, code, content)
}

func handleRequest(w http.ResponseWriter, r http.Request, request Request, server Server, response Response, route *Route) {
	if route != nil {
		if !slices.Contains(route.AllowedMethods, request.Method) {
			w.WriteHeader(405)
			w.Write([]byte("405 - Method not allowed"))
			server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, 405, r.UserAgent())
			return
		}
	}

	if response.IsRedirect {
		http.Redirect(w, &r, response.Content, response.Code)
	} else {
		w.WriteHeader(response.Code)
		w.Write([]byte(response.Content))
	}

	server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, response.Code, r.UserAgent())
}
