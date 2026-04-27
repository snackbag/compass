# Compass Contributor Docs

This directory contains internal documentation for contributors to the Compass framework.
It is not aimed at end users. It covers how the internals work, why certain decisions were
made, and what to keep in mind when extending or modifying the codebase.

## Contents

| File                               | Covers                                                     |
|------------------------------------|------------------------------------------------------------|
| [architecture.md](architecture.md) | How the pieces fit together, request lifecycle             |
| [server.md](server.md)             | Server, config, run loop, session management               |
| [route.md](route.md)               | Route registration, segment matching, parameter extraction |
| [request.md](request.md)           | The handler pipeline, 404/405, writeResponse               |
| [response.md](response.md)         | Response type, constructors, sentinel content types        |
| [cookie.md](cookie.md)             | Cookie, SameSite, Set-Cookie serialisation                 |
| [session.md](session.md)           | Session lifecycle, transactions, disk reload               |
| [logging.md](logging.md)           | Logger interface, SimpleLogger                             |
| [cors.md](cors.md)                 | CORSPolicy, Apply, WithCORS                                |

If you are reading this for the first time, start with [architecture.md](architecture.md).
It explains how all the pieces fit together before you go into the detail of any individual file.

## Before adding something

Compass is intentionally minimal. Before adding something, ask:

- Can the user solve this themselves in two lines of handler code?
- Does it need a new dependency?
- Does it make the common case harder to read?

If yes to any of these, it probably doesn't belong in the framework.