# Logging

**File:** `logger.go`

## The Logger interface

```go
type Logger interface {
    Info(message string)
    Warn(message string)
    Error(message string)
    
    Request(r *http.Request, code int)
}
```

The interface is intentionally minimal. `Info`, `Warn`, and `Error` cover structured log
levels for framework messages. `Request` is a dedicated method for HTTP request logging
because it requires both the request context and a response code: two things that are
awkward to fold into a general-purpose log call.

The `Server` holds a `Logger` field. Replacing the logger requires one assignment:

```go
server.Logger = myCustomLogger
```

Any type that satisfies the interface works. The framework never type-asserts the logger.

## SimpleLogger

```go
type SimpleLogger struct {
    PrefixMaxLength int
}
```

The default implementation. Writes coloured, timestamped lines to stdout. Suitable for
development; not suitable for structured log aggregation pipelines.

### Output format

**Level messages** (Info, Warn, Error):

```
[2006-01-02 15:04:05] INFO  Server is listening on :3000
```

Components in order: timestamp, coloured prefix, message. The prefix is padded or trimmed
to `PrefixMaxLength` characters so log lines stay horizontally aligned regardless of prefix
length.

**Request messages:**

```
127.0.0.1:54321 200 - GET /api/users "Mozilla/5.0 ..."
```

Components: remote address, coloured status code, method, path, user agent. The status
code colour follows HTTP conventions: green for 2xx, yellow for 3xx, red for 4xx/5xx.

### ANSI colour codes used

| Level / Range | Colour                         |
|---------------|--------------------------------|
| Info          | Blue (`\x1b[38;2;40;177;249m`) |
| Warn          | Bold yellow (`\033[1;33m`)     |
| Error         | Bold red (`\033[1;31m`)        |
| 2xx           | Bold green (`\033[1;32m`)      |
| 3xx           | Bold yellow (`\033[1;33m`)     |
| 4xx/5xx       | Bold red (`\033[1;31m`)        |
| Other codes   | Bold white (`\033[1;37m`)      |

These are written directly to stdout via `fmt.Printf`. There is no TTY detection. If the
output is piped to a file or log aggregator, the ANSI escape codes will be present as
literal bytes. A production logger implementation should omit them or use a library that
detects TTY.

### PrefixMaxLength

Controls the width of the level prefix column. The default is `5` (set by
`NewSimpleLogger`), which fits "INFO", "WARN", and "ERROR" with one space of padding on
the shorter ones.

If you add a new log level with a longer prefix (e.g. "DEBUG" is also 5, but
"CRITICAL" is 8), increase this value in `NewSimpleLogger` to keep alignment.

## Writing a custom logger

Implement the `Logger` interface. The most common reason to do this in production is to
write structured JSON logs, suppress ANSI codes, or forward to an external sink. Example
skeleton:

```go
type JsonLogger struct{}

func (j *JsonLogger) Info(message string) {
    j.emit("info", message)
}

func (j *JsonLogger) Warn(message string) {
    j.emit("warn", message)
}

func (j *JsonLogger) Error(message string) {
    j.emit("error", message)
}

func (j *JsonLogger) Request(r *http.Request, code int) {
    // encode as structured JSON line
}

func (j *JsonLogger) emit(level, message string) {
    // write JSON to stdout or a file
}
```

Assign it before calling `Run()`:

```go
server.Logger = &JsonLogger{}
```

## Where the logger is called

| Call site                  | Method    | When                     |
|----------------------------|-----------|--------------------------|
| `Run`                      | `Info`    | Server start message     |
| `AddRoute`                 | `Error`   | Empty route skipped      |
| `writeError`               | `Error`   | Internal server error    |
| `writeStatic` (success)    | `Request` | Static file served       |
| `write`                    | `Request` | Any response via `write` |
| `handleRequest` (redirect) | `Request` | After redirect           |
| `handleRequest` (serve)    | `Request` | After ServeContent       |

`write` covers most ordinary responses. The redirect and serve cases log explicitly because
they bypass `write`.