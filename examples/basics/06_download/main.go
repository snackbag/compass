package main

import "github.com/snackbag/compass/v2"

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/", func(request compass.Request) compass.Response {
		return compass.DownloadBytes("example.txt", []byte("helo ereveryione!!"))
	})

	server.AddRoute("/file", func(request compass.Request) compass.Response {
		return compass.DownloadFile("file.png", "example.png")
	})

	server.MustRun()
}
