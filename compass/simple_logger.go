package compass

import (
	"fmt"
	"time"
)

type SimpleLogger struct{}

func log(message string) {
	currentTime := time.Now().Format("[2006-01-02 15:04:05]")
	fmt.Printf("%s %s\033[0m\n", currentTime, message)
}

func (l *SimpleLogger) Info(message string) {
	log(fmt.Sprintf("\x1b[38;2;40;177;249mINFO\033[0m %s", message))
}

func (l *SimpleLogger) Warn(message string) {
	log(fmt.Sprintf("\033[1;33mWARN \033[0;33m%s", message))
}

func (l *SimpleLogger) Error(message string) {
	log(fmt.Sprintf("\033[1;31mERROR \033[0;31m%s", message))
}

func (l *SimpleLogger) Request(method string, ip string, route string, code int, useragent string) {
	log(fmt.Sprintf("\x1b[0;34m%s %d\033[0m - \033[0;35m%s %s\033[0m \033[0;37m\"%s\"", ip, code, method, route, useragent))
}
