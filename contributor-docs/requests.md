# Requests

**File:** `request.go`

## The Request type

```go
type Request struct {
    Method string
    URL    *url.URL
    Route  *Route
    
    Http *http.Request
}
```

`Request` is a lightweight wrapper, not a replacement. The full `*http.Request` is always
available via `Http` for anything the wrapper does not expose: body reading, form parsing,
cookies, TLS state, context, etc.

`Method` is the HTTP method normalised to lowercase (`"get"`, `"post"`, ...). This is a
convenience so handler comparisons are case-insensitive without callers needing to think
about it.

`Route` is `nil` until the server calls `FindRoute` during dispatch. Handlers can assume
`Route` is non-nil. `handleRequest` redirects to `handleNotFound` before calling the
handler if `Route` is nil.

## NewRequestFromHttp

Called by the server's catch-all handler for every incoming request. It wraps the standard
library request and normalises the method. Route assignment happens after this call, so
`Route` is always nil at construction time.

## The handler pipeline: handleRequest

`handleRequest` is the core dispatch function. Its steps in order:

**1. No route = not found**
If `r.Route` is nil, delegate immediately to `handleNotFound` and return. Nothing else
in the function runs.

**2. Call the handler**

```go
resp := r.Route.handler(r)
```

The handler receives the `Request` by value and returns a `Response` by value. It never
touches the `http.ResponseWriter`.

**3. Check for internal error**
If `resp.internalError` is true, return the error message as a Go `error`. The caller
(`Run`'s HTTP handler) passes this to `writeError`, which logs it and sends a generic code 500.
The error message from the handler is not sent to the client.

**4. Dispatch on content type**
Special internal content types are checked in a switch:

- `--COMPASS-redirect`: calls `http.Redirect` with the body as the target URL and the
  status code as the redirect type (303 or 307 for temporary, 308 for permanent).
- `--COMPASS-serve`: creates a `bytes.Reader` from the body and calls `http.ServeContent`,
  using the `-Compass-File-Name` header as the filename hint for MIME detection and
  `Content-Disposition`. The `-Compass-File-Name` header is an internal header and is
  filtered from the client response by the `--COMPASS` prefix check in `writeResponse`.

**5. Standard response to writeResponse**
Any response that is not a redirect or serve falls through to `writeResponse`. This
includes all `Text`, `Json*`, `Download*`, and `Raw` responses.

## handleNotFound

```go
func (s *Server) handleNotFound(w http.ResponseWriter, r Request) error
```

Calls `s.NotFoundHandler(r)` and passes the result directly to `writeResponse`. The not-found
handler has the same signature as any route handler: it receives a `Request` and returns a
`Response`. It can use any response constructor, set custom headers, return JSON, etc.

`NotFoundHandler` is called with a `Request` whose `Route` field is nil. Handlers that
call `GetRouteParam` or otherwise access the route will get empty strings back. This is
safe, but not useful. Not-found handlers should not attempt to extract route parameters.

## writeResponse

```go
func (s *Server) writeResponse(w http.ResponseWriter, r Request, resp Response) error
```

The shared terminal write path for all standard responses. It:

1. Iterates `resp.Headers`, skipping any key with prefix `"--COMPASS"` (internal headers).
2. Sets `Content-Type` from `resp.ContentType` if present, or mirrors the incoming
   request's `Content-Type` if not.
3. Delegates body writing to `write`, which also logs the request and writes the status code.

All callers of `write` that produce user-visible responses should go through `writeResponse`
rather than calling `write` directly. This ensures header and content-type logic is
consistent and maintained in one place.

## Internal header conventions

Headers prefixed with `--COMPASS` are reserved for internal framework use and are never
forwarded to the client. Currently used:

| Header               | Used by              | Purpose                               |
|----------------------|----------------------|---------------------------------------|
| `-Compass-File-Name` | `ServeBytesWithCode` | Filename hint for `http.ServeContent` |

Note that `-Compass-File-Name` starts with `-Compass`, not `--COMPASS`, so it is not
filtered by the prefix check in `writeResponse`. It is filtered implicitly because it is
only present when the content type is `--COMPASS-serve`, which is handled before
`writeResponse` is reached. This is subtle, but keep it in mind if the dispatch logic changes.

## Extending the pipeline

If you need a new kind of special response (e.g. server-sent events, WebSocket upgrades),
the pattern is:

1. Add a new sentinel content type string constant (e.g. `--COMPASS-sse`).
2. Add a constructor in `response.go` that sets it.
3. Add a case in the `handleRequest` switch to handle it.
4. Document the new internal header conventions here if any are needed.

Avoid adding logic to `writeResponse` for special cases. The function should stay as the
clean general path.