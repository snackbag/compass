package main

import (
	"fmt"
	"github.com/snackbag/compass"
)

// see result on localhost:3000/static/test.txt or hi.png
func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())
	err := server.Run()
	if err != nil {
		server.Logger.Error(fmt.Sprintf("Failed to start server: %v", err))
	}
}
