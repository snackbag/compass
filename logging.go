package compass

import (
	"fmt"
	"time"
)

type Logger interface {
	Info(message string)
	Warn(message string)
	Error(message string)
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

func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{PrefixMaxLength: 5}
}
