package main

import "github.com/snackbag/compass"

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())
	server.MustRun()
}
