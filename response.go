package compass

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var asciiFallback = regexp.MustCompile(`[^A-Za-z0-9 ._-]+`)

type Response struct {
	internalError bool

	ContentType *string
	Body        []byte
	StatusCode  int
	Headers     map[string]string
}

func Raw(contentType *string, body []byte, code int) Response {
	return Response{
		internalError: false,

		ContentType: contentType,
		Body:        body,
		StatusCode:  code,
		Headers:     make(map[string]string),
	}
}

func InternalError(message string, code int) Response {
	return Response{
		internalError: true,

		ContentType: nil,
		Body:        []byte(message),
		StatusCode:  code,
		Headers:     make(map[string]string),
	}
}

func Text(text string) Response {
	return TextWithCode(text, 200)
}

func TextWithCode(text string, code int) Response {
	return Raw(nil, []byte(text), code)
}

func JsonString(content string) Response {
	return JsonStringWithCode(content, 200)
}

func JsonStringWithCode(content string, code int) Response {
	typ := "application/json"
	return Raw(&typ, []byte(content), code)
}

func JsonMarshal(obj any) Response {
	return JsonMarshalWithCode(obj, 200)
}

func JsonMarshalWithCode(obj any, code int) Response {
	content, err := json.Marshal(obj)
	if err != nil {
		return InternalError(fmt.Sprintf("failed to marshal json object: %s", err), 500)
	}

	return JsonStringWithCode(string(content), code)
}

func DownloadBytes(filename string, data []byte) Response {
	return DownloadBytesWithCode(filename, data, 200)
}

func DownloadBytesWithCode(filename string, data []byte, code int) Response {
	ascii := strings.TrimSpace(filename)
	ascii = asciiFallback.ReplaceAllString(ascii, "_")

	if ascii == "" {
		ascii = "download"
	}

	resp := Raw(nil, data, code)
	resp.Headers["Content-Disposition"] = `attachment; filename="` + ascii + `"; filename*=UTF-8''` + url.PathEscape(filename)
	return resp
}

func DownloadFile(filename string, path string) Response {
	return DownloadFileWithCode(filename, path, 200)
}

func DownloadFileWithCode(filename string, path string, code int) Response {
	data, err := os.ReadFile(path)
	if err != nil {
		return InternalError(fmt.Sprintf("failed to prepare file data for download: %s", err), 500)
	}

	return DownloadBytesWithCode(filename, data, code)
}
