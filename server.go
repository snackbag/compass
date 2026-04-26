package compass

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ServerConfiguration struct {
	Port       uint16 `json:"port"`
	AssetDir   string `json:"asset_dir"`
	StaticUrl  string `json:"static_url"`
	CompassDir string `json:"compass_dir"`

	SessionExpiryTime   int `json:"session_expiry_time"`
	SessionTickInterval int `json:"session_tick_interval"`
}

type Server struct {
	Config       ServerConfiguration
	Logger       Logger
	AlertHandler func(err error)

	NotFoundHandler         func(request Request) Response
	MethodNotAllowedHandler func(request Request) Response

	routes   map[int][]*Route // int = length
	sessions map[string]*Session
}

// NewStandardConfiguration returns a default ServerConfiguration.
//
// The returned configuration contains sensible defaults for local
// development, including a default port and directory names.
func NewStandardConfiguration() ServerConfiguration {
	return ServerConfiguration{
		Port:       3000,
		AssetDir:   "assets",
		StaticUrl:  "/static",
		CompassDir: ".compass",

		SessionExpiryTime:   3 * 24 * 60 * 60 * 1000, // 72 hours
		SessionTickInterval: 5 * 60 * 1000,           // 5 minutes
	}
}

// CheckValidity validates the configuration fields.
//
// It returns an empty string if the configuration is valid.
// Otherwise, it returns a semicolon-separated string describing
// all detected issues.
func (c ServerConfiguration) CheckValidity() string {
	rv := ""

	if len(c.AssetDir) < 1 {
		rv += "asset directory value is empty;"
	}

	if len(c.StaticUrl) < 1 {
		rv += "static url is empty;"
	} else if !strings.HasPrefix(c.StaticUrl, "/") {
		rv += "static url must start with /;"
	}

	if len(c.CompassDir) < 1 {
		rv += "compass directory value is empty;"
	}

	if c.SessionExpiryTime < 1 {
		rv += "session expiry time must be above zero;"
	}

	if c.SessionTickInterval < 1 {
		rv += "session tick interval must be above zero;"
	}

	return strings.TrimSuffix(rv, ";")
}

// NewServer creates a new Server instance using the given configuration.
//
// The server is initialized with a default logger, a no-op alert handler,
// and an empty route registry. The configuration is not validated here.
func NewServer(config ServerConfiguration) *Server {
	return &Server{
		Config:       config,
		Logger:       NewSimpleLogger(),
		AlertHandler: func(err error) {},

		NotFoundHandler: func(r Request) Response {
			return TextWithCode(fmt.Sprintf("<html><h1>Not Found</h1><p>The requested route %s was not found on this server.</p></html>", r.URL.Path), http.StatusNotFound)
		},
		MethodNotAllowedHandler: func(r Request) Response {
			return TextWithCode("<html><h1>Method not allowed</h1><p>The method is not allowed for the requested URL.</p></html>", http.StatusMethodNotAllowed)
		},

		routes:   make(map[int][]*Route),
		sessions: make(map[string]*Session),
	}
}

// doManageSessionLifetimes starts a ticker for 5 minutes that checks the
// LastAccess of each loaded session. If the session is found to be expired,
// it is destroyed.
//
// This method is called after config validation in Run.
func (s *Server) doManageSessionLifetimes() {
	for range time.Tick(5 * time.Minute) {
		for _, session := range s.sessions {
			if time.Now().UnixMilli()-session.LastAccess > int64(s.Config.SessionExpiryTime) {
				session.Destroy()
				continue
			}
		}
	}
}

// Run starts the HTTP server.
//
// It first validates the configuration and returns an error if invalid.
// Incoming requests are routed based on their path. Requests matching
// the configured StaticUrl are served from the asset directory, while
// all other requests are handled by registered routes.
//
// Errors that occur during request handling are passed to writeError
// and in continuation the AlertHandler. Only startup failures are returned.
func (s *Server) Run() error {
	configValidity := s.Config.CheckValidity()
	if configValidity != "" {
		return fmt.Errorf("config invalid: %s", configValidity)
	}

	go s.doManageSessionLifetimes()
	s.Logger.Info(fmt.Sprintf("Server is listening on :%d", s.Config.Port))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		request := NewRequestFromHttp(r)
		request.Route = s.FindRoute(r.URL.Path)

		if strings.HasPrefix(request.URL.Path, s.Config.StaticUrl) {
			err := s.writeStatic(w, request, s.Config.AssetDir, strings.TrimPrefix(filepath.Clean(request.URL.Path), s.Config.StaticUrl))
			if err != nil {
				s.writeError(w, r, err)
			}

			return
		}

		err := s.handleRequest(w, request)
		if err != nil {
			s.writeError(w, r, err)
		}
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Port), nil)
}

// MustRun starts the server and exits the program if startup fails.
//
// This is a convenience wrapper around Run that calls log.Fatalf
// when an error occurs.
func (s *Server) MustRun() {
	err := s.Run()
	if err != nil {
		log.Fatalf("failed to start: %s", err)
	}
}

// writeStatic serves a static file from the asset directory.
//
// The target path is resolved relative to "<assetDir>/static".
// If the file does not exist, a 404 response is returned.
// On success, the file is served using http.ServeContent.
func (s *Server) writeStatic(w http.ResponseWriter, request Request, assetDir string, target string) error {
	path := filepath.Join(assetDir, "static", target)

	_, err := os.Stat(path)
	if err != nil {
		return s.writeResponse(w, request, s.NotFoundHandler(request))
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	s.Logger.Request(request.Http, http.StatusOK)
	http.ServeContent(w, request.Http, target, stat.ModTime(), file)
	return nil
}

// write writes raw byte data to the response with a given status code.
//
// It also logs the request. If writing fails, an error is returned.
func (s *Server) write(w http.ResponseWriter, r *http.Request, data []byte, status int) error {
	s.Logger.Request(r, status)

	w.WriteHeader(status)
	_, err := w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write byte data for status %d: %v", status, err)
	}

	return nil
}

// writeError handles internal server errors.
//
// It logs the error, triggers the AlertHandler, and sends a generic
// 500 response to the client. The actual error details are not exposed
// in the response body.
func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Request(r, http.StatusInternalServerError)
	s.Logger.Error(fmt.Sprintf("Soft capture: %v", err))
	s.AlertHandler(err)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("There was an internal server error. Try again later."))
}

// CreateSession creates a new session, writes it to disk, and registers
// it in the server's session map.
//
// The session directory is created if it does not already exist.
// The caller is responsible for attaching the session cookie to the
// response using response.SetSession
//
//		session, err := server.CreateSession()
//		if err != nil { ... }
//
//	 resp := compass.Redirect("/", false)
//	 resp.SetSession(session)
//	 return resp
func (s *Server) CreateSession() (*Session, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session id: %w", err)
	}

	dir := filepath.Join(s.Config.CompassDir, "session")
	if err = os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	session := &Session{
		server:     s,
		id:         id,
		LastAccess: time.Now().UnixMilli(),
		rwMutex:    sync.RWMutex{},
		data:       make(map[string]json.RawMessage),
	}

	session.rwMutex.Lock()
	err = session.dump()
	session.rwMutex.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to write initial session file: %w", err)
	}

	s.sessions[session.ID()] = session
	return session, nil
}

// splitUrlPath splits a URL path into its individual non-empty segments.
//
// It separates the input string by "/" and filters out any empty parts,
// which may occur due to leading, trailing, or repeated slashes.
// If the resulting slice is empty, a single empty string is returned
// to ensure the result always contains at least one element.
//
// Examples:
//
//	"/a/b/c"   -> ["a", "b", "c"]
//	"///a//b/" -> ["a", "b"]
//	"/"        -> [""]
func splitUrlPath(path string) []string {
	raw := strings.Split(path, "/")
	split := make([]string, 0)

	for _, part := range raw {
		if part == "" {
			continue
		}

		split = append(split, part)
	}

	if len(split) < 1 {
		split = append(split, "")
	}

	return split
}
