# Compass Contributor Docs

This directory contains internal documentation for contributors to the Compass framework.
It is not aimed at end users. It covers how the internals work, why certain decisions were
made, and what to keep in mind when extending or modifying the codebase.

## Contents

| File                               | What it covers                                             |
|------------------------------------|------------------------------------------------------------|
| [architecture.md](architecture.md) | High-level overview, data flow, design principles          |
| [server.md](server.md)             | `Server`, `ServerConfiguration`, the run loop              |
| [routing.md](routing.md)           | Route registration, segment matching, parameter extraction |
| [requests.md](requests.md)         | The request pipeline from HTTP to handler                  |
| [responses.md](responses.md)       | All response constructors, internal sentinel types         |
| [logging.md](logging.md)           | The `Logger` interface and `SimpleLogger`                  |
| [cors.md](cors.md)                 | `CORSPolicy`, `Apply`, and `WithCORS`                      |

## Quick orientation

If you are reading this for the first time, start with [architecture.md](architecture.md).
It explains how all the pieces fit together before you go into the detail of any individual file.

## Code style

Compass is intentionally minimal. Before adding something, ask:

- Can the user solve this themselves in two lines of handler code?
- Does this require a new dependency?
- Does this make the common case harder to read?

If any answer is yes, the addition probably does not belong in the framework core.
Keep public APIs small. Unexported helpers are fine and encouraged.