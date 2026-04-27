# Logging

**File:** `logging.go`

## Logger interface

```go
type Logger interface {
    Info(message string)
    Warn(message string)
    Error(message string)

    Request(r *http.Request, code int)
}
```

`Request` is a separate method because it needs both the request and the response code. It's awkward to 
fold into a generic log call.

Swap the logger with one assignment:

```go
server.Logger = myLogger
```

The framework never type-asserts the logger, so any implementation works.

## SimpleLogger

```go
type SimpleLogger struct {
    PrefixMaxLength int
}
```

Writes coloured, timestamped lines to stdout. Fine for development. Not suitable for
anything that needs structured output (log aggregators, JSON pipelines, etc.).

### Output format

Level messages:
```
[2006-01-02 15:04:05] INFO  Server is listening on :3000
```

Request messages:
```
127.0.0.1:54321 200 - GET /api/users "Mozilla/5.0 ..."
```

### Colours

| Level / range | Colour                         |
|---------------|--------------------------------|
| Info          | Blue (`\x1b[38;2;40;177;249m`) |
| Warn          | Bold yellow (`\033[1;33m`)     |
| Error         | Bold red (`\033[1;31m`)        |
| 2xx           | Bold green (`\033[1;32m`)      |
| 3xx           | Bold yellow (`\033[1;33m`)     |
| 4xx/5xx       | Bold red (`\033[1;31m`)        |
| Other         | Bold white (`\033[1;37m`)      |

ANSI codes go straight to stdout. There's no TTY detection. If you pipe output anywhere,
the escape codes will be there as raw bytes. We might implement a `ProductionLogger` in the future, which
is better for production use cases, but for now you have to implement your own.

### PrefixMaxLength

Controls the width of the level prefix column. Default is `5`, which fits INFO, WARN, and
ERROR with consistent alignment. Increase it if you add a prefix longer than 5 characters.

## Writing a custom logger

```go
type JsonLogger struct{}

func (j *JsonLogger) Info(message string)               { j.emit("info", message) }
func (j *JsonLogger) Warn(message string)               { j.emit("warn", message) }
func (j *JsonLogger) Error(message string)              { j.emit("error", message) }
func (j *JsonLogger) Request(r *http.Request, code int) { /* write structured line */ }
func (j *JsonLogger) emit(level, message string)        { /* write to stdout or file */ }
```

Assign it before calling `Run()`:

```go
server.Logger = &JsonLogger{}
```

## Where the logger is called

| Call site                  | Method    | When                     |
|----------------------------|-----------|--------------------------|
| `Run`                      | `Info`    | server start             |
| `AddRoute`                 | `Error`   | empty route skipped      |
| `writeError`               | `Error`   | internal server error    |
| `writeStatic` (success)    | `Request` | static file served       |
| `write`                    | `Request` | any normal response      |
| `handleRequest` (redirect) | `Request` | after redirect           |
| `handleRequest` (serve)    | `Request` | after ServeContent       |