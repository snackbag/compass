package compass

import (
	"fmt"
	"net/http"
	"strings"
)

type ServerConfiguration struct {
	Port       uint16 `json:"port"`
	AssetDir   string `json:"asset_dir"`
	StaticUrl  string `json:"static_url"`
	CompassDir string `json:"compass_dir"`
}

type Server struct {
	Config ServerConfiguration
	Logger Logger
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
		Config: config,
		Logger: NewSimpleLogger(),
	}
}

// Run starts the server, returns an error when the server crashes
func (s *Server) Run() error {
	configValidity := s.Config.CheckValidity()
	if configValidity != "" {
		return fmt.Errorf("config invalid: %s", configValidity)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

	})

	return http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Port), nil)
}
