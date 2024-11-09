package compass

import (
	"net/http"
	"net/url"
)

type Request struct {
	Method    string
	IP        string
	URL       url.URL
	UserAgent string
}

func NewRequest(r http.Request) Request {
	return Request{
		Method:    r.Method,
		IP:        r.RemoteAddr,
		URL:       *r.URL,
		UserAgent: r.UserAgent(),
	}
}

type Response struct {
	IsRedirect bool
	Code       int
	Content    string
}

func Redirect(target string) Response {
	return Response{IsRedirect: true, Code: 307, Content: target}
}

func RedirectWithCode(target string, code int) Response {
	return Response{IsRedirect: true, Code: code, Content: target}
}

func Text(content string) Response {
	return Response{IsRedirect: false, Code: 200, Content: content}
}

func TextWithCode(content string, code int) Response {
	return Response{IsRedirect: false, Code: code, Content: content}
}

func handleRequest(w http.ResponseWriter, r http.Request, request Request, server Server, response Response) {
	if response.IsRedirect {
		http.Redirect(w, &r, response.Content, response.Code)
	} else {
		w.Write([]byte(response.Content))
	}

	server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, response.Code, r.UserAgent())
}
