# Responses

**File:** `response.go`

## Response

```go
type Response struct {
    internalError bool
    cookies       []Cookie

    ContentType *string
    Body        []byte
    StatusCode  int
    Headers     map[string]string
}
```

Handlers build a `Response` and return it. They never write anything themselves.

`internalError` can only be set via `InternalError()`. When true, the pipeline treats the
body as an operator-facing error message and sends a generic 500 to the client instead.

`ContentType` is a pointer. nil means "not set", which causes `writeResponse` to fall back
to `"text/html; charset=utf-8"`. An empty string and nil are different things here.

`Headers` is always a non-nil map. You can add headers directly after calling any constructor:

```go
resp := compass.Text("ok")
resp.Headers["X-Foo"] = "bar"
return resp
```

## Cookie methods

`SetCookie(c Cookie)` appends a cookie. Multiple calls are fine, because each becomes its own
`Set-Cookie` header.

`RemoveCookie(name string)` appends a cookie with `Max-Age: 0`, which tells the browser to
delete it.

`SetSession(session *Session)` calls `SetCookie` and sets the `_compassId` cookie.

## Constructor hierarchy

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

| Value                  | Used by                     | Handled in             |
|------------------------|-----------------------------|------------------------|
| `"--COMPASS-redirect"` | `Redirect`, `PermaRedirect` | `handleRequest` switch |
| `"--COMPASS-serve"`    | `ServeBytesWithCode`        | `handleRequest` switch |

These start with `"--COMPASS"` so they can't collide with real MIME types. They never
reach the client.

## Download vs Serve

**Download** (`DownloadBytes`, `DownloadFile`) sets `Content-Disposition: attachment`,
which tells the browser to save the file. The filename is encoded in both ASCII (fallback)
and UTF-8 via RFC 5987 `filename*` so non-ASCII names work in modern browsers and degrade
to `_` substitution in old ones. If the whole name sanitises to empty, `"download"` is used.

**Serve** (`ServeBytes`, `ServeFile`) uses `http.ServeContent` via the `--COMPASS-serve`
sentinel. The browser decides whether to display inline or download based on the MIME type
of the filename. `ServeContent` also handles `Range` requests and `304 Not Modified`.
`DownloadBytes` does not, so if range support matters, use the Serve family.

## InternalError

`InternalError(message, code)` marks the response so the pipeline sends a generic 500 and
passes the message to `AlertHandler`.

## Adding a constructor

1. Write `XxxWithCode(... , code int) Response`.
2. Write `Xxx(...)` that calls it with `200`.
3. If it needs special pipeline handling, add a sentinel and a case in `handleRequest`.
4. Add it to the hierarchy above.