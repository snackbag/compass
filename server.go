package compass

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ServerConfiguration struct {
	Port       uint16 `json:"port"`
	AssetDir   string `json:"asset_dir"`
	StaticUrl  string `json:"static_url"`
	CompassDir string `json:"compass_dir"`
}

type Server struct {
	Config       ServerConfiguration
	Logger       Logger
	AlertHandler func(err error)

	routes map[int][]*Route // int = length
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

		routes: make(map[int][]*Route),
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
		return s.handleNotFound(w, request)
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
