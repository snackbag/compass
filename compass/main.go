package compass

import (
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	Port   int
	Logger Logger
}

type Logger interface {
	Info(message string)
	Warn(message string)
	Error(message string)

	Request(method string, ip string, route string, code int, useragent string)
}

func NewServer() Server {
	return Server{Port: 3000, Logger: NewLogger()}
}

func NewLogger() Logger {
	return &SimpleLogger{}
}

func (server *Server) Start() {
	http.HandleFunc("/", getRoot)

	server.Logger.Info(fmt.Sprintf("Server is listening on :%d", server.Port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", server.Port), nil)
	if err != nil {
		panic(err)
	}
}

func getRoot(writer http.ResponseWriter, request *http.Request) {
	_, err := io.WriteString(writer, "Test")
	if err != nil {
		panic(err)
	}
}
