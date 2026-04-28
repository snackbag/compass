package main

import (
	"github.com/snackbag/compass/v2"
	"os"
	"path/filepath"
)

func handleDownload(request compass.Request) compass.Response {
	filename, ok := request.GetRouteParam("file")
	if !ok {
		return compass.TextWithCode("You need to provide a file name", 400)
	}

	path := filepath.Join("uploads", filename)
	if _, err := os.Stat(path); err != nil {
		return compass.TextWithCode("There is no such file", 404)
	}

	return compass.DownloadFile(filename, path)
}
