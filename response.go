package compass

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// Raw creates a basic Response with full control over its contents.
//
// It sets the body, status code, and optional content type. Headers
// are initialized as an empty map and can be modified after creation.
func Raw(contentType *string, body []byte, code int) Response {
	return Response{
		internalError: false,

		ContentType: contentType,
		Body:        body,
		StatusCode:  code,
		Headers:     make(map[string]string),
	}
}

// InternalError creates a Response that signals an internal server error.
//
// This does not directly write a response to the client. Instead, it marks
// the response as an internal error so the server can handle it separately.
func InternalError(message string, code int) Response {
	return Response{
		internalError: true,

		ContentType: nil,
		Body:        []byte(message),
		StatusCode:  code,
		Headers:     make(map[string]string),
	}
}

// Text creates a plain text response with status code 200.
func Text(text string) Response {
	return TextWithCode(text, 200)
}

// TextWithCode creates a plain text response with a custom status code.
func TextWithCode(text string, code int) Response {
	return Raw(nil, []byte(text), code)
}

// JsonString creates a JSON response from a raw string with status code 200.
//
// The string is assumed to already be valid JSON.
func JsonString(content string) Response {
	return JsonStringWithCode(content, 200)
}

// JsonStringWithCode creates a JSON response from a raw string
// with a custom status code.
//
// The string is assumed to already be valid JSON.
func JsonStringWithCode(content string, code int) Response {
	typ := "application/json"
	return Raw(&typ, []byte(content), code)
}

// JsonMarshal converts an object to JSON and returns it as a response.
//
// If marshalling fails, an InternalError response is returned instead.
func JsonMarshal(obj any) Response {
	return JsonMarshalWithCode(obj, 200)
}

// JsonMarshalWithCode converts an object to JSON and returns it
// with a custom status code.
//
// If marshalling fails, an InternalError response is returned instead.
func JsonMarshalWithCode(obj any, code int) Response {
	content, err := json.Marshal(obj)
	if err != nil {
		return InternalError(fmt.Sprintf("failed to marshal json object: %s", err), 500)
	}

	return JsonStringWithCode(string(content), code)
}

// DownloadBytes creates a response that forces the client to download data
// as a file with status code 200.
//
// The filename is sanitized for ASCII compatibility while preserving the
// original name for UTF-8 capable clients.
func DownloadBytes(filename string, data []byte) Response {
	return DownloadBytesWithCode(filename, data, 200)
}

// DownloadBytesWithCode creates a response that forces the client to download data
// as a file with a custom status code.
//
// The filename is sanitized for ASCII compatibility while preserving the
// original name for UTF-8 capable clients.
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

// DownloadFile reads a file from disk and returns it as a download response.
//
// If the file cannot be read, an internal error response is returned.
func DownloadFile(filename string, path string) Response {
	return DownloadFileWithCode(filename, path, 200)
}

// DownloadFileWithCode reads a file from disk and returns it as a download
// response with a custom status code.
//
// If the file cannot be read, an internal error response is returned.
func DownloadFileWithCode(filename string, path string, code int) Response {
	data, err := os.ReadFile(path)
	if err != nil {
		return InternalError(fmt.Sprintf("failed to prepare file data for download: %s", err), 500)
	}

	return DownloadBytesWithCode(filename, data, code)
}

// Redirect creates a redirect response to the given target.
//
// If retainMethod is true, a 307 redirect is used. Otherwise, a 303
// redirect is used, which converts the request to a GET.
func Redirect(target string, retainMethod bool) Response {
	typ := "--COMPASS-redirect"
	status := http.StatusSeeOther
	if retainMethod {
		status = http.StatusTemporaryRedirect
	}

	return Raw(&typ, []byte(target), status)
}

// PermaRedirect creates a permanent redirect (HTTP 308) to the target URL.
// The method which a request uses is retained.
func PermaRedirect(target string) Response {
	typ := "--COMPASS-redirect"
	return Raw(&typ, []byte(target), http.StatusPermanentRedirect)
}

// ServeBytes creates a response that serves data as a file-like resource
// with status code 200.
//
// Unlike "download" responses, this does not force a download. The client
// (e.g. browser) may choose to display the content inline if it supports it
// (such as images, text, or PDFs).
func ServeBytes(data []byte, name string) Response {
	return ServeBytesWithCode(data, name, 200)
}

// ServeBytesWithCode creates a response that serves data as a file-like
// resource with a custom status code.
//
// Unlike "download" responses, this does not force a download. The client
// (e.g. browser) may choose to display the content inline if it supports it
// (such as images, text, or PDFs).
func ServeBytesWithCode(data []byte, name string, code int) Response {
	typ := "--COMPASS-serve"
	raw := Raw(&typ, data, code)
	raw.Headers["-Compass-File-Name"] = name
	return raw
}

// ServeFile reads a file from disk and serves it as a file-like response.
//
// This behaves the same as ServeBytes, but reads the file from disk first.
// If the file cannot be read, an internal error response is returned.
func ServeFile(path string, name string) Response {
	return ServeFileWithCode(path, name, 200)
}

// ServeFileWithCode reads a file from disk and serves it as a file-like
// response with a custom status code.
//
// This behaves the same as ServeBytes, but reads the file from disk first.
// If the file cannot be read, an internal error response is returned.
func ServeFileWithCode(path string, name string, code int) Response {
	data, err := os.ReadFile(path)
	if err != nil {
		return InternalError(fmt.Sprintf("failed to prepare file data for serve: %s", err), 500)
	}

	return ServeBytesWithCode(data, name, code)
}
