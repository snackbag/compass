package compass

import (
	"fmt"
	"net/http"
	"time"
)

// Logger defines a minimal logging interface used by the server.
//
// It supports basic log levels (Info, Warn, Error) and a dedicated
// method for logging HTTP requests.
type Logger interface {
	Info(message string)
	Warn(message string)
	Error(message string)

	Request(r *http.Request, code int)
}

// SimpleLogger is a basic console logger implementation.
//
// It prints colored log messages to stdout and formats them with
// timestamps and fixed-width prefixes for alignment.
type SimpleLogger struct {
	PrefixMaxLength int
}

// log formats and prints a log message with a timestamp and prefix.
//
// The prefix is trimmed or padded to PrefixMaxLength to keep output
// aligned. ANSI color codes are used for styling.
func (s *SimpleLogger) log(color string, prefix string, message string) {
	currentTime := time.Now().Format("[2006-01-02 15:04:05]")

	if len(prefix) > s.PrefixMaxLength {
		prefix = prefix[:s.PrefixMaxLength]
	}
	prefix = fmt.Sprintf("%-*s", s.PrefixMaxLength, prefix)

	fmt.Printf("%s %s%s\033[0m %s\033[0m\n", currentTime, color, prefix, message)
}

// Info logs a message with the "INFO" level.
func (s *SimpleLogger) Info(message string) {
	s.log("\x1b[38;2;40;177;249m", "INFO", message)
}

// Warn logs a message with the "WARN" level.
func (s *SimpleLogger) Warn(message string) {
	s.log("\033[1;33m", "WARN", message)
}

// Error logs a message with the "ERROR" level.
func (s *SimpleLogger) Error(message string) {
	s.log("\033[1;31m", "ERROR", message)
}

// Request logs an HTTP request and its response status code.
//
// The status code is color-coded based on its range:
//
//	2xx = green, 3xx = yellow, 4xx/5xx = red.
//
// It also logs the remote address, method, path, and user agent.
func (s *SimpleLogger) Request(r *http.Request, code int) {
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
		"\x1b[0;34m%s %s%d\033[0m - \033[0;35m%s %s\033[0m \033[0;37m\"%s\"\033[0m\n",
		r.RemoteAddr, colorCode, code, r.Method, r.URL.Path, r.UserAgent(),
	)
}

// NewSimpleLogger creates a SimpleLogger with default settings.
//
// The default PrefixMaxLength is set to 5.
func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{PrefixMaxLength: 5}
}
