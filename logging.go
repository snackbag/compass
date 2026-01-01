package compass

import (
	"fmt"
	"time"
)

type Logger interface {
	Info(message string)
	Warn(message string)
	Error(message string)

	Request(method string, ip string, route string, code int, useragent string)
}

type SimpleLogger struct {
	PrefixMaxLength int
}

func (s *SimpleLogger) log(color string, prefix string, message string) {
	currentTime := time.Now().Format("[2006-01-02 15:04:05]")

	if len(prefix) > s.PrefixMaxLength {
		prefix = prefix[:s.PrefixMaxLength]
	}
	prefix = fmt.Sprintf("%-*s", s.PrefixMaxLength, prefix)

	fmt.Printf("%s %s%s\033[0m %s\033[0m\n", currentTime, color, prefix, message)
}

func (s *SimpleLogger) Info(message string) {
	s.log("\x1b[38;2;40;177;249m", "INFO", message)
}

func (s *SimpleLogger) Warn(message string) {
	s.log("\033[1;33m", "WARN", message)
}

func (s *SimpleLogger) Error(message string) {
	s.log("\033[1;31m", "ERROR", message)
}

func (s *SimpleLogger) Request(method string, ip string, route string, code int, useragent string) {
	var colorCode string
	switch {
	case code >= 200 && code < 300:
		colorCode = "\033[1;32m"
	case code >= 300 && code < 400:
		colorCode = "\033[1;33m"
	case code >= 400 && code < 600:
		colorCode = "\033[1;31m"
	default:
		colorCode = "\033[1;37m"
	}

	fmt.Printf(
		"\x1b[0;34m%s %s%d\033[0m - \033[0;35m%s %s\033[0m \033[0;37m\"%s\"",
		ip, colorCode, code, method, route, useragent,
	)
}

func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{PrefixMaxLength: 5}
}
