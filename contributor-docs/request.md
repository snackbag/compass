# Requests

**File:** `request.go`

## Request

```go
type Request struct {
    Method string
    URL    *url.URL
    Route  *Route

    Http *http.Request
}
```

`Http` is the underlying `*http.Request`. Use it for body reading, form parsing, TLS
state, context, anything the wrapper doesn't yet expose.

`Method` is the HTTP method lowercased. `"GET"` becomes `"get"`.

`Route` is nil until `FindRoute` runs during dispatch. By the time a handler is called,
`Route` is always set. `handleRequest` returns early through `NotFoundHandler` before
reaching any handler if `Route` is nil.

## NewRequestFromHttp

Called once per request. Wraps the standard http request and lowercases the method. `Route` is
assigned separately after this returns.

## handleRequest

The main dispatch function, in order:

1. No route delegates to `NotFoundHandler`, writes the response, returns. The `Request` passed to
`NotFoundHandler` has `Route == nil`. Don't try to read route params in a not-found handler.

2. Method not allowed if the route doesn't include the request's method in `AllowedMethods`, delegates 
to `MethodNotAllowedHandler` and returns.

3. Call the handler

```go
resp := r.Route.handler(r)
```

4. Internal error if `resp.internalError` is true, the body is returned as a Go
`error`. The caller passes it to `writeError`, which logs it and sends a generic 500. The
message never reaches the client.

5. Special content types are checked in a switch:
- `--COMPASS-redirect`: calls `http.Redirect` with the body as the target URL.
- `--COMPASS-serve`: calls `http.ServeContent` with a `bytes.Reader` from the body. The 
filename hint comes from the `-Compass-File-Name` internal header.

Both paths write cookies first and log the request after.

6. Everything else goes through `writeResponse`.

## writeResponse

```go
func (s *Server) writeResponse(w http.ResponseWriter, r Request, resp Response) error
```

1. Calls `writeCookies` to write `Set-Cookie` headers.
2. Writes `resp.Headers`, skipping any key prefixed with `--COMPASS`.
3. Sets `Content-Type` from `resp.ContentType` if non-nil; otherwise defaults to `"text/html; charset=utf-8"`.
4. Calls `write` to write the status code and body.

## Internal headers

Headers prefixed `--COMPASS` are never sent to the client. The one internal header that
doesn't follow this prefix is `-Compass-File-Name`, which is only present when the content
type is `--COMPASS-serve`, so it's never reached by `writeResponse` anyway. This is very bad
practice, and we should really change this.

## Adding a new response type

1. Pick a sentinel content type string starting with `--COMPASS`.
2. Add a constructor in `response.go`.
3. Add a case in the `handleRequest` switch.
4. Add any internal headers to the table above.

Avoid adding special cases to `writeResponse`. Keep it as the clean general path.