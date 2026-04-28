package main

import (
	"github.com/snackbag/compass"
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

	_, err = io.Copy(upload, file)
	if err != nil {
		return compass.InternalError(err.Error())
	}

	return compass.Redirect("/", false)
}
