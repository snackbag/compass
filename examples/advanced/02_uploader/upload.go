package main

import (
	"github.com/snackbag/compass/v2"
	"io"
	"os"
	"path/filepath"
)

const MaxUploadSize = 128 << 20 // 128MB; << 20 is the magic that converts a number into bytes

func handleUpload(request compass.Request) compass.Response {
	if request.Method == "get" {
		return compass.ServeFile("template/upload.html", "index.html")
	}

	err := request.Http.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		return compass.InternalError(err.Error())
	}

	file, header, err := request.Http.FormFile("file")
	if err != nil {
		return compass.Text("Please attach a file!")
	}

	defer file.Close()

	// IMPORTANT! Sanitize your file names! Never trust the client
	filename := filepath.Base(header.Filename)
	filename = filepath.Clean(filename)

	path := filepath.Join("uploads", filename)
	os.MkdirAll(filepath.Dir(path), 0755)
	upload, err := os.Create(path)
	if err != nil {
		return compass.InternalError(err.Error())
	}
	defer upload.Close()

	limited := io.LimitReader(file, MaxUploadSize+1)
	n, err := io.Copy(upload, limited)
	if n > MaxUploadSize {
		os.Remove(path)
		return compass.Text("File too large!")
	}

	return compass.Redirect("/", false)
}
