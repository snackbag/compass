package compass

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

type Any struct {
	Value interface{}
}

func (any *Any) ToString() string {
	return fmt.Sprintf("%v", any.Value)
}

type Route struct {
	Pattern        RoutePattern
	AllowedMethods []string
	Handler        func(request Request) Response
}

type RoutePattern struct {
	raw          string
	segments     []string
	paramIndices map[int]string // maps segment index to parameter name
	prefixParams map[int]string // maps segment index to prefix-based parameter name
	suffixParams map[int]struct {
		name   string
		suffix string
	}
}

func parseRoutePattern(pattern string) RoutePattern {
	segments := strings.Split(strings.Trim(pattern, "/"), "/")
	paramIndices := make(map[int]string)
	prefixParams := make(map[int]string)
	suffixParams := make(map[int]struct {
		name   string
		suffix string
	})

	for i, segment := range segments {
		if strings.HasPrefix(segment, "<") && strings.HasSuffix(segment, ">") {
			// Standard parameter like <name>
			paramName := segment[1 : len(segment)-1]
			paramIndices[i] = paramName
		} else if strings.Contains(segment, "<") && strings.Contains(segment, ">") {
			// Parameter with prefix/suffix
			start := strings.Index(segment, "<")
			end := strings.Index(segment, ">")

			if start == 0 {
				// Has suffix only
				paramName := segment[1:end]
				suffix := segment[end+1:]
				suffixParams[i] = struct {
					name   string
					suffix string
				}{paramName, suffix}
			} else if end == len(segment)-1 {
				// Has prefix only
				prefix := segment[:start]
				paramName := segment[start+1 : end]
				if strings.HasPrefix(prefix, "@") {
					prefixParams[i] = paramName
				}
			}
		}
	}

	return RoutePattern{
		raw:          pattern,
		segments:     segments,
		paramIndices: paramIndices,
		prefixParams: prefixParams,
		suffixParams: suffixParams,
	}
}

func matchRoute(pattern RoutePattern, path string) (bool, map[string]string) {
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathSegments) != len(pattern.segments) {
		return false, nil
	}

	params := make(map[string]string)

	for i, segment := range pattern.segments {
		pathSegment := pathSegments[i]

		// Check standard parameters
		if paramName, ok := pattern.paramIndices[i]; ok {
			params[paramName] = pathSegment
			continue
		}

		// Check prefix parameters
		if paramName, ok := pattern.prefixParams[i]; ok {
			if !strings.HasPrefix(pathSegment, "@") {
				return false, nil
			}
			params[paramName] = pathSegment[1:]
			continue
		}

		// Check suffix parameters
		if param, ok := pattern.suffixParams[i]; ok {
			if !strings.HasSuffix(pathSegment, param.suffix) {
				return false, nil
			}
			params[param.name] = strings.TrimSuffix(pathSegment, param.suffix)
			continue
		}

		// Check static segments
		if segment != pathSegment {
			return false, nil
		}
	}

	return true, params
}

type Server struct {
	Port               int
	Logger             Logger
	StaticDirectory    string
	StaticRoute        string
	TemplatesDirectory string
	SessionDirectory   string
	sessionSecret      *string

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
		SessionDirectory:   fmt.Sprintf(".compass%csessions", filepath.Separator),
		sessionSecret:      nil,
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

func (server *Server) SetSessionSecret(secret string) {
	if server.sessionSecret != nil {
		panic("Cannot set session secret on a server that already has a set secret")
	}

	server.sessionSecret = &secret
}

func (server *Server) Start() {
	if server.sessionSecret == nil {
		server.Logger.Error("CRITICAL: Session secret is not set. Please set this to something secure.")
	}

	if _, err := os.Stat(server.StaticDirectory); os.IsNotExist(err) {
		server.Logger.Warn(fmt.Sprintf("static directory '%s' does not exist.", server.StaticDirectory))
	}

	if _, err := os.Stat(server.TemplatesDirectory); os.IsNotExist(err) {
		server.Logger.Warn(fmt.Sprintf("templates directory '%s' does not exist.", server.TemplatesDirectory))
	}

	if _, err := os.Stat(server.SessionDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(server.SessionDirectory, 0755)
		if err != nil {
			panic(err)
		}

		server.Logger.Info(fmt.Sprintf("Created session directory at %s", server.SessionDirectory))
	}

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handled := false
		for i := range server.routes {
			route := &server.routes[i]
			if matches, params := matchRoute(route.Pattern, r.URL.Path); matches {
				request := NewRequest(*r, server)
				request.routeParams = params
				handleRequest(w, *r, request, *server, route.Handler(request), route)
				handled = true
				break
			}
		}

		if !handled {
			request := NewRequest(*r, server)
			handleRequest(w, *r, request, *server, server.notFoundHandler(request), nil)
		}
	})

	// Handle static
	http.HandleFunc(server.StaticRoute+"/", func(w http.ResponseWriter, r *http.Request) {
		filePath := server.StaticDirectory + r.URL.Path[len(server.StaticRoute):]

		file, err := os.Open(filePath)
		if err != nil {
			request := NewRequest(*r, server)
			handleRequest(w, *r, request, *server, server.notFoundHandler(request), nil)
			return
		}
		defer file.Close()

		fileStat, err := file.Stat()
		if err != nil {
			request := NewRequest(*r, server)
			handleRequest(w, *r, request, *server, server.notFoundHandler(request), nil)
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

func (server *Server) AddRoute(pattern string, handler func(request Request) Response) *Route {
	route := &Route{
		Pattern:        parseRoutePattern(pattern),
		Handler:        handler,
		AllowedMethods: []string{"GET"},
	}
	server.routes = append(server.routes, *route)
	return route
}

func (server *Server) SetAllowedMethod(route *Route, method string, allowed bool) {
	for i := range server.routes {
		if server.routes[i].Pattern.raw == route.Pattern.raw {
			if allowed {
				if !slices.Contains(server.routes[i].AllowedMethods, method) {
					server.routes[i].AllowedMethods = append(server.routes[i].AllowedMethods, method)
				}
			} else {
				server.routes[i].AllowedMethods = slices.DeleteFunc(server.routes[i].AllowedMethods, func(m string) bool {
					return m == method
				})
			}
			route.AllowedMethods = server.routes[i].AllowedMethods
			break
		}
	}
}

func (server *Server) SetNotFoundHandler(handler func(request Request) Response) {
	server.notFoundHandler = handler
}
