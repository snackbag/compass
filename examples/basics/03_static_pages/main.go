package main

import (
	"github.com/snackbag/compass/v2"
)

// see result on localhost:3000/static/test.txt or hi.png
func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())
	server.MustRun()
}
