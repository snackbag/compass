package compass

type ServerConfiguration struct {
	Port       uint16 `json:"port"`
	AssetDir   string `json:"asset_dir"`
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
		CompassDir: ".compass",
	}
}

func NewServer(config ServerConfiguration) *Server {
	return &Server{
		Config: config,
		Logger: NewSimpleLogger(),
	}
}
