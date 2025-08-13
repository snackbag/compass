package compass

type Server struct {
	Port   int
	Logger Logger
}

func NewServer() *Server {
	return &Server{
		Port:   3000,
		Logger: NewSimpleLogger(),
	}
}
