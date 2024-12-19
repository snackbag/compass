package compass

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"time"
)

type Request struct {
	Method      string
	IP          string
	URL         url.URL
	UserAgent   string
	FullRequest http.Request
	Cookies     []*http.Cookie
	Server      *Server

	routeParams map[string]string
	r           http.Request
}

func (request *Request) GetParam(name string) string {
	return request.routeParams[name]
}

func NewRequest(r http.Request, server *Server) Request {
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
		Cookies:     r.Cookies(),
		routeParams: make(map[string]string),
		r:           r,
		Server:      server,
	}
}

func (request *Request) GetCookie(name string) *http.Cookie {
	cookie, err := request.r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil
		}

		panic(err)
	}

	return cookie
}

func (request *Request) GetSession() *Session {
	possibleCookie := request.GetCookie("_compassId")
	var session *Session

	if possibleCookie == nil {
		session = NewSession(request.Server)
	} else {
		recv := GetSessionById(request.Server, possibleCookie.Value)
		if recv == nil {
			return nil
		}

		session = recv
	}

	session.ResetTransaction()
	return session
}

type Response struct {
	IsRedirect bool
	Code       int
	Content    string

	cookies        map[string]http.Cookie
	removedCookies []string
}

func (resp *Response) SetCookie(cookie http.Cookie) {
	resp.cookies[cookie.Name] = cookie
}

func (resp *Response) RemoveCookie(cookie http.Cookie) {
	delete(resp.cookies, cookie.Name)
	resp.removedCookies = append(resp.removedCookies)
}

func NewResponse(isRedirect bool, code int, content string) Response {
	return Response{IsRedirect: isRedirect, Code: code, Content: content, cookies: make(map[string]http.Cookie)}
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
	for _, cookie := range response.removedCookies {
		http.SetCookie(w, &http.Cookie{Name: cookie, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
	}

	for _, cookie := range response.cookies {
		http.SetCookie(w, &cookie)
	}

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
