package compass

import (
	"fmt"
	"net/http"
	"os"
)

type Route struct {
	path    string
	handler func(request Request) string
}

type Server struct {
	Port               int
	Logger             Logger
	StaticDirectory    string
	StaticRoute        string
	TemplatesDirectory string

	routes []Route
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

	server.Logger.Info(fmt.Sprintf("Server is listening on :%d", server.Port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", server.Port), nil)
	if err != nil {
		panic(err)
	}
}

func (server *Server) AddRoute(path string, handler func(request Request) string) {
	server.routes = append(server.routes, Route{
		path:    path,
		handler: handler,
	})
}
