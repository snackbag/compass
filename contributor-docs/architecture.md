# Architecture

## Overview

Compass wraps Go's standard `net/http` package. It handles route matching, response
writing, session storage, and static file serving so you don't have to write the boilerplate
yourself. The underlying `net/http` types are always accessible if you need them.

One third-party dependency: `github.com/google/uuid` for session IDs.

## Package structure

```
compass/
  server.go   - Server struct, config, the HTTP listener, static file serving
  route.go    - route registration and path matching
  request.go  - Request type, the handler pipeline (incl. 404/405)
  response.go - Response type and constructors
  cookie.go   - Cookie type, SameSite constants, Set-Cookie serialisation
  session.go  - Session and SessionTransaction
  logging.go  - Logger interface and SimpleLogger
  cors.go     - CORSPolicy, Apply, WithCORS
```

## Request lifecycle

```
net/http.ListenAndServe
    └── http.HandleFunc("/", ...)
            ├── NewRequestFromHttp(r)               - wrap *http.Request
            ├── FindRoute(r.URL.Path)               - attach matching *Route (or nil)
            │
            ├── [static path?]
            │       └── writeStatic(...)            - serve from assets/static/
            │
            └── handleRequest(w, request)
                    ├── [no route?]                 -> NotFoundHandler -> writeResponse
                    ├── [method not allowed?]       -> MethodNotAllowedHandler -> writeResponse
                    ├── execute preprocessor        -> not nil? -> writeResponse   
                    │       ├── [not nil?]          - write preprocessor response
                    │       └── execute route func  -> MethodNotAllowedHandler -> writeResponse
                    ├── [internalError?]            -> writeError
                    ├── [redirect?]                 -> http.Redirect
                    ├── [serve?]                    -> http.ServeContent
                    └── writeResponse(w, r, resp)
                            ├── writeCookies
                            ├── write headers
                            ├── set Content-Type
                            └── write body + status
```

The handler gets a `Request` and returns a `Response`. It never touches
`http.ResponseWriter`. Compass handles all the actual writing.

## Types

**Server** holds config, the logger, the alert handler, the not-found and method-not-allowed
handlers, the route map, and the session map.

**Route** holds the parsed URL pattern, the handler function, and the allowed HTTP methods.

**Request** wraps `*http.Request`, adds the matched `*Route`, and normalises the method to
lowercase. The original `*http.Request` is always at `Request.Http`.

**Response** is a plain struct: body, status code, optional content type, headers map, and
a list of cookies. Handlers build one and return it. They never write anything directly.

**Cookie** maps to the attributes of a `Set-Cookie` header.

**Session** stores key-value data as a JSON file on disk. Reads and writes go through
`SessionGet` and `SessionTransaction`. See [session.md](session.md).

## Design notes

**Handlers return values, not writers.** A handler is just a function you can call in a
test and inspect the return value. No mocking needed.

**Compass owns the write path.** Cookies, headers, content type, and body are all written
in one place (`writeResponse`). This keeps header ordering consistent and means handlers
can't accidentally write a header after the status code.

**Special responses use sentinel content types.** Redirects and file serves are just normal
`Response` values with an internal content type string (`--COMPASS-redirect`,
`--COMPASS-serve`). `handleRequest` detects them and dispatches accordingly. Handlers just
call `Redirect()` or `ServeFile()` like anything else.

**Errors don't panic.** `InternalError()` returns a `Response` that the pipeline forwards
to `writeError`. `writeError` logs it, calls `AlertHandler`, and sends a generic 500. The
internal message never reaches the client.

**Customisation should be simple.** `NotFoundHandler`, `MethodNotAllowedHandler`, and `AlertHandler`
are `func` fields on `Server`. Swap them out before calling `Run()`.