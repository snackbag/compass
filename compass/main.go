package compass

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
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

	routes          []Route
	notFoundHandler func(request Request) string
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
		notFoundHandler: func(request Request) string {
			return fmt.Sprintf(
				"<h1>Not Found</h1>"+
					"<p>The requested URL %s was not found on this server</p>"+
					"<hr/>"+
					"<i>Compass (%s) Server on port %d<i>",
				request.URL.Path, runtime.GOOS, 3000)
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
				response := route.handler(Request{
					Method:    r.Method,
					IP:        r.RemoteAddr,
					URL:       *r.URL,
					UserAgent: r.UserAgent(),
				})

				w.Write([]byte(response))

				server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, http.StatusOK, r.UserAgent())
				handled = true
				break
			}
		}

		if !handled {
			response := server.notFoundHandler(Request{
				Method:    r.Method,
				IP:        r.RemoteAddr,
				URL:       *r.URL,
				UserAgent: r.UserAgent(),
			})

			w.Write([]byte(response))
			server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, http.StatusNotFound, r.UserAgent())
		}
	})

	// Handle static
	http.HandleFunc(server.StaticRoute+"/", func(w http.ResponseWriter, r *http.Request) {
		filePath := server.StaticDirectory + r.URL.Path[len(server.StaticRoute):]
		server.Logger.Info(filePath)

		file, err := os.Open(filePath)
		if err != nil {
			http.NotFound(w, r)
			server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, http.StatusNotFound, r.UserAgent())
			return
		}
		defer file.Close()

		fileStat, err := file.Stat()
		if err != nil {
			http.NotFound(w, r)
			server.Logger.Request(r.Method, r.RemoteAddr, r.URL.Path, http.StatusNotFound, r.UserAgent())
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

func (server *Server) AddRoute(path string, handler func(request Request) string) {
	server.routes = append(server.routes, Route{
		path:    path,
		handler: handler,
	})
}

func (server *Server) SetNotFoundHandler(handler func(request Request) string) {
	server.notFoundHandler = handler
}
