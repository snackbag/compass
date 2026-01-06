package compass

import (
	"fmt"
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

func NewStandardConfiguration() ServerConfiguration {
	return ServerConfiguration{
		Port:       3000,
		AssetDir:   "assets",
		StaticUrl:  "/static",
		CompassDir: ".compass",
	}
}

// CheckValidity returns an empty string when valid, otherwise a list of human-readable reasons why the config is invalid
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

func NewServer(config ServerConfiguration) *Server {
	return &Server{
		Config:       config,
		Logger:       NewSimpleLogger(),
		AlertHandler: func(err error) {},

		routes: make(map[int][]*Route),
	}
}

// Run starts the server, returns an error when the server crashes during startup. All other errors are handled by the Server.AlertHandler
func (s *Server) Run() error {
	configValidity := s.Config.CheckValidity()
	if configValidity != "" {
		return fmt.Errorf("config invalid: %s", configValidity)
	}

	s.Logger.Info(fmt.Sprintf("Server is listening on :%d", s.Config.Port))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		request := NewRequestFromHttp(r)

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

func (s *Server) write(w http.ResponseWriter, r *http.Request, data []byte, status int) error {
	s.Logger.Request(r, status)

	w.WriteHeader(status)
	_, err := w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write byte data for status %d: %v", status, err)
	}

	return nil
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Request(r, http.StatusInternalServerError)
	s.Logger.Error(fmt.Sprintf("Soft capture: %v", err))
	s.AlertHandler(err)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("There was an internal server error. Try again later."))
}
