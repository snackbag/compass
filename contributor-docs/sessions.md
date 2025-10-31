# Sessions

Compass handles sessions for you.

A session is built up like this:

- true ID; an integer stored only on the server as primary key, autoincrement.
- UUID; the visual session ID. Stored in the server's database as `UNIQUE`, and on the client as the `_compassId`
  cookie. It is what the server uses to distinguish sessions per request.
- Value; a JSON string stored only on the server.
- Expiry; the time a session takes until it expires. Stored on the server as unix timestamp, while the client's cookie
  is set to never disappear.

The server uses SQLite and has an index for the true ID and UUID.

## Useful files
- [/session.go](/session.go) - session creation, lookup & checking
- [/db.go](/db.go) - database setup
