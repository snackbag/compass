package compass

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Session represents an active user session backed by a JSON file on disk.
//
// Values are stored as raw JSON internally, so any JSON-serialisable type
// can be stored. Use SessionGet to retrieve typed values, and BeginTx to
// make changes.
//
// Sessions are safe for concurrent reads. Writes must go through a
// SessionTransaction and be finalised with Commit.
type Session struct {
	server *Server

	id         uuid.UUID
	LastAccess int64

	rwMutex      sync.RWMutex
	lastModified int64 // mtime of the file at last load, in UnixNano
	destroyed    bool
	data         map[string]json.RawMessage
}

// ID returns the session identifier as a UUID string.
//
// This value is stored in the client's _compassId cookie and used
// as the session's filename on disk.
func (s *Session) ID() string {
	return s.id.String()
}

// Destroy marks the session for deletion.
//
// The session file is removed from disk and the client cookie is expired
// when the response is written. Calling Commit on any transaction belonging
// to a destroyed session is ignored.
//
// Returns an error if we failed to dump the session data
func (s *Session) Destroy() error {
	s.destroyed = true
	return s.dump()
}

// MustDestroy executes Destroy, but any error that is thrown is forwarded to the AlertHandler.
func (s *Session) MustDestroy() {
	err := s.Destroy()
	if err != nil {
		err = fmt.Errorf("failed to destroy session: %w", err)
		s.server.Logger.Error(err.Error())
		s.server.AlertHandler(err)
	}
}

// filePath returns the absolute path of this session's JSON file on disk.
func (s *Session) filePath() string {
	return filepath.Join(s.server.Config.CompassDir, "session", s.ID()+".json")
}

// checkReload stats the session file and calls reloadFromDisk if the file
// has been modified since the last load.
//
// The stat itself is done outside any lock. If a reload is needed, the write
// lock is acquired inside reloadFromDisk. This means the common no
// change case pays only the cost of a stat and a read lock check.
func (s *Session) checkReload() {
	info, err := os.Stat(s.filePath())
	if err != nil {
		return
	}

	mtime := info.ModTime().UnixNano()

	s.rwMutex.RLock()
	needsReload := mtime != s.lastModified
	s.rwMutex.RUnlock()

	if needsReload {
		s.reloadFromDisk()
	}
}

// reloadFromDisk reads the session file from disk and replaces the in-memory
// data with its contents.
//
// If the file cannot be read or parsed, the existing in-memory data is left
// unchanged. The write lock is held for the duration of the reload.
//
// This is called automatically by SessionGet when the file's mtime has
// changed since the last load, allowing external modifications, such as
// those made by another server instance, to be picked up transparently.
func (s *Session) reloadFromDisk() {
	unlocked := false // cursed

	s.rwMutex.Lock()
	defer func() { // very cursed stuff here
		if unlocked {
			return
		}
		unlocked = true
		s.rwMutex.Unlock()
	}()

	info, err := os.Stat(s.filePath())
	if err != nil {
		return
	}

	raw, err := os.ReadFile(s.filePath())
	if err != nil {
		return
	}

	var newData map[string]json.RawMessage
	if err := json.Unmarshal(raw, &newData); err != nil {
		return
	}

	s.data = newData
	s.lastModified = info.ModTime().UnixNano()

	s.rwMutex.Unlock()
	unlocked = true // the most curserest
	s.LastAccess = SessionGetOrDefault(s, "--COMPASS-Last-Access", s.lastModified)
	s.destroyed = SessionGetOrDefault(s, "--COMPASS-Destroyed", false)
}

// dump serialises the session's current data map to disk.
//
// It must be called with the write lock already held. After a successful
// write, lastModified is updated to match the file's new mtime so that
// the next checkReload does not trigger a redundant reload.
func (s *Session) dump() error {
	wdat := s.data
	b, _ := json.Marshal(time.Now().UnixMilli())
	wdat["--COMPASS-Last-Access"] = b
	b, _ = json.Marshal(s.destroyed)
	wdat["--COMPASS-Destroyed"] = b

	data, err := json.Marshal(wdat)
	if err != nil {
		return fmt.Errorf("failed to marshal during commit of session %s: %w", s.ID(), err)
	}

	path := s.filePath()
	if err = os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write during commit of session %s: %w", s.ID(), err)
	}

	// Update lastModified so checkReload does not immediately reload just written file!!!!!!!!!!!!!!!!
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat after commit of session %s: %w", s.ID(), err)
	}
	s.lastModified = info.ModTime().UnixNano()

	return nil
}

// SessionGet retrieves and deserializes a value from the session by key.
//
// The file is reloaded from disk first if its mtime has changed since the
// last load. If the key does not exist, or deserialisation fails, a non-nil
// error is returned and the zero value of T is returned.
//
// Example:
//
//	user, err := compass.SessionGet[User](session, "user")
func SessionGet[T any](s *Session, key string) (T, error) {
	s.checkReload()

	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	var zero T
	raw, ok := s.data[key]
	if !ok {
		return zero, fmt.Errorf("session %s does not have key %q", s.ID(), key)
	}

	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal key %q of session %s: %w", key, s.ID(), err)
	}

	s.LastAccess = time.Now().UnixMilli()
	return result, nil
}

// SessionGetOrDefault is a wrapper around SessionGet, which attempts to get
// the specified key, but upon any error, it silently returns the fallback value.
func SessionGetOrDefault[T any](s *Session, key string, fallback T) T {
	result, err := SessionGet[T](s, key)
	if err != nil {
		return fallback
	}

	return result
}

//
// Transactions
//

// SessionTransaction holds a set of pending changes to a Session.
//
// Changes are not applied to the session or written to disk until Commit
// is called. A nil value in the changes map signals that the key should
// be deleted from the session.
//
// Create a transaction with Session.BeginTx.
type SessionTransaction struct {
	session *Session
	changes map[string]*json.RawMessage
}

// BeginTx creates a new SessionTransaction for this session.
//
// The transaction starts empty. Use Set and Delete to stage changes,
// then call Commit to apply them atomically.
func (s *Session) BeginTx() *SessionTransaction {
	return &SessionTransaction{
		session: s,
		changes: make(map[string]*json.RawMessage),
	}
}

func (s *Session) cookie() Cookie {
	return Cookie{
		Name:     "_compassId",
		Value:    s.ID(),
		HttpOnly: true,
		SameSite: SameSiteLax,
		Path:     "/",
	}
}

// Commit applies all staged changes to the session and writes it to disk.
//
// The write lock is held for the duration of the apply and the disk write.
// If the session has been destroyed, Commit is ignored and returns nil.
func (tx *SessionTransaction) Commit() error {
	if tx.session.destroyed {
		return nil
	}

	tx.session.rwMutex.Lock()
	defer tx.session.rwMutex.Unlock()

	for k, v := range tx.changes {
		if v == nil {
			delete(tx.session.data, k)
			continue
		}
		tx.session.data[k] = *v
	}

	return tx.session.dump()
}

// Set stages a value to be written to the session under the given key.
//
// The value is marshalled to JSON immediately. If marshalling fails, an
// error is returned and the key is not staged. The change is not applied
// to the session until Commit.
func (tx *SessionTransaction) Set(key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %q: %w", key, err)
	}
	raw := json.RawMessage(b)
	tx.changes[key] = &raw
	return nil
}

// Delete stages the removal of a key from the session.
//
// If the key does not exist in the session, this is ignored at Commit time.
// If Set was previously called for the same key in this transaction, Delete
// takes precedence.
func (tx *SessionTransaction) Delete(key string) {
	tx.changes[key] = nil
}
