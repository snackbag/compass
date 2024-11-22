package compass

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
)

type Any struct {
	Value interface{}
}

func (any *Any) ToString() string {
	return fmt.Sprintf("%v", any.Value)
}

type Route struct {
	path           string
	allowedMethods []string
	handler        func(request Request) Response
}

type Server struct {
	Port               int
	Logger             Logger
	StaticDirectory    string
	StaticRoute        string
	TemplatesDirectory string

	routes          []Route
	notFoundHandler func(request Request) Response
}

type Logger interface {
	Info(message string)
	Warn(message string)
	Error(message string)

	Request(method string, ip string, route string, code int, useragent string)
}

func NewServer() Server {
	return Server{
		Port:               3000,
		Logger:             NewLogger(),
		StaticDirectory:    "static",
		StaticRoute:        "/static",
		TemplatesDirectory: "templates",
		routes:             []Route{},
		notFoundHandler: func(request Request) Response {
			return TextWithCode(fmt.Sprintf(
				"<h1>Not Found</h1>"+
					"<p>The requested URL %s was not found on this server</p>"+
					"<hr/>"+
					"<i>Compass (%s) Server on port %d<i>",
				request.URL.Path, runtime.GOOS, 3000), 404)
		},
	}
}

func NewLogger() Logger {
	return &SimpleLogger{}
}

func (server *Server) Start() {
	if _, err := os.Stat(server.StaticDirectory); os.IsNotExist(err) {
		server.Logger.Warn(fmt.Sprintf("static directory '%s' does not exist.", server.StaticDirectory))
	}

	if _, err := os.Stat(server.TemplatesDirectory); os.IsNotExist(err) {
		server.Logger.Warn(fmt.Sprintf("templates directory '%s' does not exist.", server.TemplatesDirectory))
	}

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handled := false
		for _, route := range server.routes {
			if r.URL.Path == route.path {
				request := NewRequest(*r)
				handleRequest(w, *r, request, *server, route.handler(request))
				handled = true
				break
			}
		}

		if !handled {
			request := NewRequest(*r)
			handleRequest(w, *r, request, *server, server.notFoundHandler(request))
		}
	})

	// Handle static
	http.HandleFunc(server.StaticRoute+"/", func(w http.ResponseWriter, r *http.Request) {
		filePath := server.StaticDirectory + r.URL.Path[len(server.StaticRoute):]

		file, err := os.Open(filePath)
		if err != nil {
			request := NewRequest(*r)
			handleRequest(w, *r, request, *server, server.notFoundHandler(request))
			return
		}
		defer file.Close()

		fileStat, err := file.Stat()
		if err != nil {
			request := NewRequest(*r)
			handleRequest(w, *r, request, *server, server.notFoundHandler(request))
			return
		}

		http.ServeContent(w, r, filePath, fileStat.ModTime(), file)
		server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, http.StatusOK, r.UserAgent())
	})

	server.Logger.Info(fmt.Sprintf("Server is listening on :%d", server.Port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", server.Port), nil)
	if err != nil {
		panic(err)
	}
}

func (server *Server) AddRoute(path string, handler func(request Request) Response) Route {
	route := Route{
		path:           path,
		handler:        handler,
		allowedMethods: []string{"GET"},
	}
	server.routes = append(server.routes, route)

	return route
}

func (server *Server) SetAllowedMethod(route Route, method string, allowed bool) {
	if allowed {
		for _, m := range route.allowedMethods {
			if m == method {
				return
			}
		}
		route.allowedMethods = append(route.allowedMethods, method)
	} else {
		newMethods := make([]string, 0, len(route.allowedMethods))
		for _, m := range route.allowedMethods {
			if m != method {
				newMethods = append(newMethods, m)
			}
		}
		route.allowedMethods = newMethods
	}
}

func (server *Server) SetNotFoundHandler(handler func(request Request) Response) {
	server.notFoundHandler = handler
}
