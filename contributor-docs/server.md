# Server

**File:** `server.go`

## ServerConfiguration

```go
type ServerConfiguration struct {
    Port       uint16
    AssetDir   string
    StaticUrl  string
    CompassDir string
}
```

| Field        | Purpose                                      | Default      |
|--------------|----------------------------------------------|--------------|
| `Port`       | TCP port to listen on                        | `3000`       |
| `AssetDir`   | Root directory for static assets             | `"assets"`   |
| `StaticUrl`  | URL prefix that triggers static file serving | `"/static"`  |
| `CompassDir` | Directory for internal Compass state         | `".compass"` |

`NewStandardConfiguration()` returns a value with these defaults pre-filled. It is the
recommended starting point. Callers can then override individual fields.

`CheckValidity()` returns a semicolon-separated string of all problems found, or an empty
string if the configuration is valid. It is called at the start of `Run()`. If it returns
anything non-empty, `Run()` returns an error immediately without starting the listener.
This means misconfigured servers fail loudly at startup rather than silently at request time.

Note that `CheckValidity` does not check whether directories actually exist on disk. It
only validates that the string values are structurally sensible (non-empty, correct prefix
for URLs). Filesystem checks happen lazily when requests arrive.

## Server

```go
type Server struct {
    Config          ServerConfiguration
    Logger          Logger
    AlertHandler    func (err error)
    NotFoundHandler func (request Request) Response
    
    routes map[int][]*Route
}
```

All public fields are intended to be replaced by callers after `NewServer()` and before
`Run()`. The unexported `routes` field is managed exclusively by `AddRoute` and `FindRoute`.

`AlertHandler` is called whenever an internal error occurs during request handling.
Thta is, whenever a handler returns an `InternalError` response, or whenever writing to
the client fails. The default is a no-op. In production, this is the right place to hook
in error reporting (Sentry, logging services, alerting, etc.).

`NotFoundHandler` is called when no route matches the incoming request path. It
receives the `Request` (with `Route` set to `nil`) and returns a `Response` like any
other handler. The default returns a plain HTML 404 page. See [requests.md](requests.md)
for how `handleNotFound` uses this.

## Run loop

`Run()` does three things:

1. Validates configuration via `CheckValidity()`.
2. Registers a single catch-all handler on `"/"` via `http.HandleFunc`.
3. Calls `http.ListenAndServe`.

The single catch-all handler inspects each incoming request:

- If the path starts with `Config.StaticUrl`, it delegates to `writeStatic`.
- Otherwise, it attaches the matched route (or nil) and delegates to `handleRequest`.

Any error from either path is passed to `writeError`, which logs it, calls `AlertHandler`,
and sends a generic code 500 to the client. Startup errors (invalid config, port already in use)
are returned directly from `Run()`.

`MustRun()` wraps `Run()` and calls `log.Fatalf` on error. Suitable for main functions that
have no meaningful error recovery.

## Static file serving

`writeStatic` resolves the request path relative to `<AssetDir>/static/`. The `StaticUrl`
prefix is stripped from the path before joining, so a request to `/static/css/main.css`
with default config looks for `assets/static/css/main.css` on disk.

`filepath.Clean` is applied to the incoming path before stripping the prefix. This prevents
path traversal attacks (e.g. `/static/../../secret`) from escaping the asset directory.

If the target file does not exist (`os.Stat` fails), `handleNotFound` is called. The same
not-found path as missing routes. If the file exists but cannot be opened, the error is
returned and handled by `writeError`.

On success, `http.ServeContent` is used rather than a raw write. This handles `Range`
requests, `Last-Modified` headers, and conditional `304` responses automatically.

## Internal write helpers

`write(w, r, data, status)` is the low-level byte writer. It logs the request, writes
the status code, and writes the body. It should be used only by the framework itself, not
by handler code.

`writeError(w, r, err)` handles unrecoverable errors during request handling. It logs
the error internally, triggers `AlertHandler`, and sends a generic code 500 body to the client.
Importantly, it does not expose the internal error message to the client. That message is
for the operator, not the user.

`writeResponse(w, r, resp)` is the shared path for writing any standard `Response`
(non-redirect, non-serve). It writes headers, the content type, and delegates body writing
to `write`. Both `handleRequest` and `handleNotFound` use this. Any future handler that
produces a standard response should use it too.

## Adding new customisation points

Follow the pattern of `NotFoundHandler` and `AlertHandler`:

1. Add a `func` field to `Server`.
2. Assign a sensible default in `NewServer()`.
3. Document the field and its signature here.

Avoid interface fields for single-function customisation points. A plain `func` is simpler
to replace and easier to understand at a glance.