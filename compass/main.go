package compass

import (
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	Port int
}

func (server *Server) Start() {
	http.HandleFunc("/", getRoot)

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
