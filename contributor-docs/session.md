# Sessions

**File:** `session.go`

## Overview

Each session is a JSON file in `.compass/session/<uuid>.json`. The UUID is stored in the
client's `_compassId` cookie. Sessions survive restarts because the data is on disk, not
just in memory.

## Session struct

```go
type Session struct {
    server     *Server
    id         uuid.UUID
    LastAccess int64        // UnixMilli
    rwMutex    sync.RWMutex
    destroyed  bool
    data       map[string]json.RawMessage
}
```

`LastAccess` is exported so the reaper goroutine can read it. The framework updates it automatically
when `SessionGet` is invoked.

## Sessions loading

When the server starts, it walks all files in `.compass/session/` and attempts to parse all JSON files 
as sessions.

## Creating a session

Always use `Server.CreateSession()`. It generates a UUID, creates the session directory if
needed, writes the initial file, registers the session in `s.sessions`, and returns the
`*Session`. Then attach it to the response:

```go
session, err := server.CreateSession()
if err != nil { ... }

resp := compass.Redirect("/", false)
resp.SetSession(session) // sets the _compassId cookie
return resp
```

Without `SetSession`, the client never gets the cookie and the session can't be found on
the next request.

## Getting a session

```go
session, ok := r.GetSession(server)
if !ok {
    return compass.Redirect("/login", false)
}
```

`GetSession` reads the `_compassId` cookie and looks it up in `server.sessions`. Returns
false if the cookie is missing, the ID is unknown, or the session has been destroyed.

## SessionGet

```go
func SessionGet[T any](s *Session, key string) (T, error)
```

A top-level generic function because Go doesn't support generic methods.

Before reading, it calls `checkReload`. If the session file's mtime has changed since
the last load, it reloads from disk. This lets multiple processes sharing the same
`.compass/` directory pick up each other's changes.

Returns the zero value of `T` and an error if the key doesn't exist or unmarshalling fails.

There is also `SessionGetOrDefault` which takes a fallback and returns it instead of an error.

```go
user, err := compass.SessionGet[User](session, "user")
```

## Writing via SessionTransaction

Changes are staged in a transaction and applied on `Commit`. This avoids partially-written
sessions if the process crashes mid-write.

```go
tx := session.BeginTx()
tx.Set("user", user)       // marshals immediately
tx.Delete("temp_token")    // staged removal
tx.Commit()                // writes to disk
```

`Set` marshals the value immediately. If that fails, the key isn't staged.

`Delete` and `Set` for the same key in one transaction: `Delete` wins.

`Commit` on a destroyed session is ignored.

## Destroying a session

```go
session.Destroy()
```

Marks it as destroyed. The file and the `s.sessions` entry are cleaned up on the next
reaper tick. To expire the client's cookie at the same time:

```go
resp.RemoveCookie("_compassId")
```

## Disk reload

`checkReload` stats the file. If the mtime differs from `lastModified`, it calls
`reloadFromDisk` under a write lock. The stat is done outside the lock to keep the
common no-change case fast.

`reloadFromDisk` reads the file and replaces `s.data`. If reading or parsing fails, the
existing data is left as-is.

## dump

Must be called with the write lock held. Marshals `s.data`, writes the file, and updates
`lastModified` from the file's new mtime. The mtime update is important, since without it the
very next `checkReload` would see a changed mtime and reload data the process just wrote.

## Concurrency

Reads hold `rwMutex.RLock`. Writes hold `rwMutex.Lock`. Multiple concurrent reads are
fine; a write blocks everything else.

`s.sessions` on `Server` is not protected by a lock. This works because `CreateSession`
is only called from request handlers (serialised by `net/http`) and the reaper only
deletes sessions already marked destroyed. If Compass is ever extended with true concurrent
session creation outside `net/http`'s model, a mutex around `s.sessions` will be needed.