# Server

**File:** `server.go`

## ServerConfiguration

```go
type ServerConfiguration struct {
    Port       uint16
    AssetDir   string
    StaticUrl  string
    CompassDir string
    
    SessionExpiryTime   int
    SessionTickInterval int
}

```

| Field                 | Default      | What it does                                        |
|-----------------------|--------------|-----------------------------------------------------|
| `Port`                | `3000`       | TCP port                                            |
| `AssetDir`            | `"assets"`   | Root directory for static files                     |
| `StaticUrl`           | `"/static"`  | URL prefix that triggers static file serving        |
| `CompassDir`          | `".compass"` | Where Compass stores internal state (sessions etc.) |
| `SessionExpiryTime`   | `259200000`  | How long (ms) a session can go untouched (72h)      |
| `SessionTickInterval` | `300000`     | How often (ms) the session reaper runs (5 min)      |

`NewStandardConfiguration()` returns a value with these defaults. Override individual
fields after calling it.

`CheckValidity()` returns a semicolon-separated list of problems, or an empty string if
everything looks fine. `Run()` calls this first and returns an error immediately if the
config is invalid. It only checks that values are structurally sensible (non-empty strings,
positive numbers, URL starts with `/`). It does not check whether directories exist.

## Server

```go
type Server struct {
    Config       ServerConfiguration
    Logger       Logger
    AlertHandler func(err error)

    NotFoundHandler         func(request Request) Response
    MethodNotAllowedHandler func(request Request) Response

    routes   map[int][]*Route
    sessions map[string]*Session
}
```

Public fields can be replaced after `NewServer()` and before `Run()`. The unexported
`routes` and `sessions` fields are managed by the framework.

`AlertHandler` is called when a handler returns an `InternalError` or when writing to the
client fails. The default does nothing. Hook into an error reporting service here.

`NotFoundHandler` is called when no route matches. The `Request` it receives has `Route`
set to nil. The default returns a plain HTML 404 page.

`MethodNotAllowedHandler` is called when a route matches the path but not the method. The
default returns a plain HTML 405 page.

## Run

`Run()` validates the config, starts the session reaper goroutine, registers a catch-all
handler on `"/"`, and calls `http.ListenAndServe`. Errors that happen during request
handling go to `writeError`. Only startup failures are returned from `Run()`.

`MustRun()` calls `Run()` and calls `log.Fatalf` if it fails.

The catch-all handler does two things: if the path starts with `Config.StaticUrl`, it
serves a static file. Otherwise, it routes to a handler.

## Session management

`doManageSessionLifetimes()` runs in a goroutine. Every `SessionTickInterval` milliseconds
it checks all in-memory sessions and destroys any whose `LastAccess` is older than
`SessionExpiryTime`. Destroyed sessions get their file deleted and are removed from
`s.sessions`.

This only acts on sessions loaded into memory in the current process.

`CreateSession()` generates a UUID, creates the `.compass/session/` directory if needed,
writes an empty JSON file, registers the session in `s.sessions`, and returns the
`*Session`. The caller must call `resp.SetSession(session)` to give the client the cookie:

```go
session, err := server.CreateSession()
if err != nil { ... }

resp := compass.Redirect("/", false)
resp.SetSession(session)
return resp
```

## Static file serving

`writeStatic` joins the request path onto `<AssetDir>/static/` after stripping the
`StaticUrl` prefix. So `/static/css/main.css` with default config maps to
`assets/static/css/main.css`.

`filepath.Clean` is applied before the prefix is stripped to block path traversal
(`/static/../../etc/passwd` style attempts).

If the file doesn't exist, `NotFoundHandler` is used. If it exists, `http.ServeContent`
handles it. This means `Range` requests, `Last-Modified`, and `304` responses work
automatically.

## Write helpers

`write(w, r, data, status)` writes the status code and body and logs the request. Used
only by the framework.

`writeError(w, r, err)` logs the error, calls `AlertHandler`, and sends a generic 500.
The error message is not sent to the client.

`writeResponse(w, r, resp)` is the normal write path for all standard responses. It calls
`writeCookies`, writes headers (skipping any with the `--COMPASS` prefix), sets
`Content-Type`, and calls `write`. All standard handler responses go through here.