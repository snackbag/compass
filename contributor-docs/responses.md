# Responses

**File:** `response.go`

## The Response type

```go
type Response struct {
   internalError bool
   
   ContentType *string
   Body        []byte
   StatusCode  int
   Headers     map[string]string
}

```

`Response` is a plain value type. Handlers construct one and return it. They do not write
to any writer. The framework's pipeline (in `handleRequest` / `writeResponse`) owns the
actual write.

`internalError` is unexported. It can only be set via `InternalError()`. When true, the
pipeline treats the body as an error message for the operator, not as a client response.

`ContentType` is a pointer to allow the nil case to mean "not set", which the pipeline
uses to mirror the incoming request's content type. The empty string `""` and a nil
pointer have distinct meanings.

`Headers` is always initialised as a non-nil `map[string]string` by all constructors.
Handlers can safely add headers without a nil check:

```go
resp := compass.Text("ok")
resp.Headers["X-Request-Id"] = "abc123"
return resp
```

## Constructor hierarchy

The constructors are layered. Lower-level ones delegate upward:

```
Raw(contentType, body, code)
    └── TextWithCode(text, code)
            └── Text(text)
    └── JsonStringWithCode(content, code)
            └── JsonString(content)
    └── DownloadBytesWithCode(filename, data, code)
            └── DownloadBytes(filename, data)
    └── ServeBytesWithCode(data, name, code)
            └── ServeBytes(data, name)

JsonMarshalWithCode(obj, code)   - calls JsonStringWithCode after marshalling
    └── JsonMarshal(obj)

DownloadFileWithCode(filename, path, code)   - reads file, calls DownloadBytesWithCode
    └── DownloadFile(filename, path)

ServeFileWithCode(path, name, code)   - reads file, calls ServeBytesWithCode
    └── ServeFile(path, name)

InternalError(message, code)   - sets internalError=true, does not use Raw

Redirect(target, retainMethod)   - uses Raw with sentinel content type
PermaRedirect(target)            - uses Raw with sentinel content type
```

When adding a new constructor, follow this pattern: implement the `WithCode` variant
as the real function and make the plain variant call it with `200`. This keeps the
default-200 convenience while avoiding duplicated logic.

## Sentinel content types

Two content type strings are reserved for internal framework use:

| Sentinel               | Used by                     | Handled in             |
|------------------------|-----------------------------|------------------------|
| `"--COMPASS-redirect"` | `Redirect`, `PermaRedirect` | `handleRequest` switch |
| `"--COMPASS-serve"`    | `ServeBytesWithCode`        | `handleRequest` switch |

These strings start with `"--COMPASS"` to make collisions with real MIME types impossible.
The pipeline checks for them before writing any headers, so they never reach the client.

## Download vs Serve

There are two families of file-related constructors with distinct behaviours:

**Download (`DownloadBytes`, `DownloadFile`)** sets `Content-Disposition: attachment`.
The browser is instructed to save the file to disk. The filename is encoded in both ASCII
(for compatibility) and UTF-8 via `RFC 5987` `filename*` syntax, so filenames with
non-ASCII characters work correctly in modern clients while degrading gracefully in older
ones.

The ASCII fallback replaces non-ASCII characters with `_` using the `asciiFallback` regex,
compiled once at package level. If the sanitised name is empty (e.g. the filename was
entirely non-ASCII), `"download"` is used instead.

**Serve (`ServeBytes`, `ServeFile`)** uses `http.ServeContent` via the `--COMPASS-serve`
sentinel. The browser decides whether to display inline or offer a download based on the
MIME type it detects from the filename and content. Images, PDFs, and text files typically
display inline; binary files are downloaded.

`ServeContent` also handles `Range` requests and `304 Not Modified` automatically.
`DownloadBytes` does not. If range support matters, use the Serve family.

## InternalError

```go
func InternalError(message string, code int) Response
```

Creates a response that signals a server-side failure. When the pipeline sees
`resp.internalError == true`, it calls `writeError` with the message as the error. The
message is logged and passed to `AlertHandler`, but never sent to the client.

The `code` field on an `InternalError` response is ignored by the pipeline (the client
always gets a generic code 500 via `writeError`). It exists for informational purposes only.
For example, `JsonMarshalWithCode` sets it to 500 so the code is meaningful if the
response is ever inspected in tests.

## Adding a new constructor

1. Implement `XxxWithCode(... , code int) Response` as the core variant.
2. Implement `Xxx(...) Response` calling `XxxWithCode` with `200`.
3. If the new type needs special handling in the pipeline, add a sentinel type and a case
   in `handleRequest`. See [requests.md](requests.md) for the extension pattern.
4. Add the constructor to the hierarchy diagram above.