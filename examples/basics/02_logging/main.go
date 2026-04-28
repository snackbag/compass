package main

import "github.com/snackbag/compass/v2"

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.Logger.Info("I'm simply logging some things")
	server.Logger.Warn("I warn you about stuff")
	server.Logger.Error("I am a big red scary text")
}
