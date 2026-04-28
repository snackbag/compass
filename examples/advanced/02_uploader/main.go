package main

import (
	"github.com/snackbag/compass/v2"
)

var server *compass.Server

func main() {
	server = compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/", handleIndex)
	server.AddRoute("/upload", handleUpload).AllowedMethods = []string{"get", "post"}
	server.AddRoute("/download/<file>", handleDownload)

	server.MustRun()
}
