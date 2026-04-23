# Architecture

## Overview

Compass is a thin wrapper around Go's standard `net/http` package. It does not replace
the standard library — it organises it. The goal is to remove the boilerplate that every
project ends up writing anyway (route matching, response construction, static file serving)
while staying close enough to `net/http` that the underlying request and response writer
are always accessible.

There are no third-party dependencies. Every file in the package is either framework logic
or standard library usage.

## Package structure

```
compass/
  server.go     - Server struct, configuration, the HTTP listener, static file serving
  routing.go    - Route registration and path matching
  request.go    - Request type, the handler pipeline, 404 dispatch
  response.go   - Response type and all constructors
  logger.go     - Logger interface and SimpleLogger
  cors.go       - CORSPolicy, Apply, WithCORS
```

Each file has a focused responsibility. There is intentionally no god-object or global state.

## Request lifecycle

```
net/http.ListenAndServe
    └── http.HandleFunc("/", ...)
            ├── NewRequestFromHttp(r)        - wrap *http.Request into Request
            ├── FindRoute(r.URL.Path)        - attach matching *Route (or nil)
            │
            ├── [static path?]
            │       └── writeStatic(...)     - serve from assets/static/
            │
            └── handleRequest(w, request)
                    ├── [no route?] -> handleNotFound -> NotFoundHandler(r) -> writeResponse
                    ├── [internalError?] -> writeError
                    ├── [redirect?] -> http.Redirect
                    ├── [serve?] -> http.ServeContent
                    └── writeResponse(w, r, resp)
```

The handler registered on a `Route` receives a `Request` and returns a `Response`. It never
touches `http.ResponseWriter` directly. The framework owns that boundary.

## Core types

### Server

The central struct. Holds configuration, the logger, the alert handler, the
not-found handler, and the route registry. All state lives here. There is no global
mutable state anywhere in the package.

### Route

Owns a parsed representation of a URL pattern and a handler function. Routes are grouped
in a `map[int][]*Route` keyed by segment count, which avoids scanning every route on
every request. See [routing.md](routing.md) for detail.

### Request

A lightweight wrapper over `*http.Request`. Adds the matched `*Route` reference and
normalises the HTTP method to lowercase. The original `*http.Request` is always
accessible via `Request.Http` for anything the wrapper doesn't expose.

### Response

A value type (not a pointer) containing a body, status code, optional content type, and
a header map. Constructors return ready-to-use values. The handler pipeline inspects the
response and writes it to the client. The handler itself never writes anything directly.

## Design principles

**Handlers return values, not writers.**
Handlers take a `Request` and return a `Response`. They never receive an
`http.ResponseWriter`. This makes handlers trivially testable. Call the function,
inspect the return value.

**The framework owns the write path.**
Everything between "handler returns" and "bytes on the wire" is framework code.
This is where headers are written, content types are set, and special response types
(redirects, file serves) are dispatched. Centralising this prevents subtle bugs from
callers writing headers in the wrong order.

**Sentinel content types for special responses.**
Redirects and served files are represented as normal `Response` values with internal
content type strings (`--COMPASS-redirect`, `--COMPASS-serve`). This means the handler
still returns a plain `Response` with no special interface; the framework detects and
handles them in `handleRequest`. Users never see these strings, they use `Redirect()` and
`ServeFile()` like any other constructor.

**Errors escalate, never panic.**
`InternalError` creates a `Response` that carries an error message but does not write
anything to the client. The pipeline detects it and passes it to `writeError`, which
triggers the `AlertHandler`. Nothing in the framework panics.

**Swappable behaviour via func fields.**
`AlertHandler` and `NotFoundHandler` are plain `func` fields on `Server`. Replacing
them requires one assignment. No interfaces, no registration methods, no builder pattern.
This is the model to follow for any future customisation points.