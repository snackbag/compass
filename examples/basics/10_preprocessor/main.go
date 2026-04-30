package main

import "github.com/snackbag/compass/v2"

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.Preprocessor = func(request compass.Request) *compass.Response {
		if request.URL.Query().Get("secret") == "peace-love-and-plants" {
			return nil
		}

		resp := compass.Text("You must know the secret to enter...")
		return &resp
	}

	server.AddRoute("/", func(request compass.Request) compass.Response {
		return compass.Text("You're awesome")
	})

	server.MustRun()
}
