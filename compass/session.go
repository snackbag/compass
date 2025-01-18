package compass

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type UUID [16]byte

func NewUUID() UUID {
	var uuid UUID
	_, err := rand.Read(uuid[:])
	if err != nil {
		panic(err)
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return uuid
}

func UUIDToString(uuid UUID) string {
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:])
	return string(buf)
}

func StringToUUID(str string) (UUID, error) {
	var uuid UUID
	if len(str) != 36 {
		return UUID{}, errors.New("invalid UUID length")
	}

	uuidHex := make([]byte, 32)
	copy(uuidHex[0:8], str[0:8])
	copy(uuidHex[8:12], str[9:13])
	copy(uuidHex[12:16], str[14:18])
	copy(uuidHex[16:20], str[19:23])
	copy(uuidHex[20:], str[24:])

	_, err := hex.Decode(uuid[:], uuidHex)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid UUID format: %w", err)
	}

	return uuid, nil
}

type Session struct {
	Server *Server
	ID     UUID

	transaction map[string]interface{}
	vars        map[string]interface{}
}

func NewSession(server *Server) *Session {
	s := &Session{
		Server:      server,
		ID:          NewUUID(),
		transaction: make(map[string]interface{}),
	}

	s.WriteInt64("_compassLastUpdate", time.Now().UnixMilli())
	s.Commit()
	return s
}

func (session *Session) Read(key string, dflt interface{}) interface{} {
	transVal, transOk := session.transaction[key]
	if transOk {
		return transVal
	}

	val, ok := session.vars[key]
	if !ok {
		return dflt
	}

	return val
}

func (session *Session) WriteString(key string, value string) {
	session.transaction[key] = value
}

func (session *Session) ReadString(key string, dflt string) string {
	return session.Read(key, dflt).(string)
}

func (session *Session) WriteInt(key string, value int) {
	session.transaction[key] = value
}

func (session *Session) ReadInt(key string, dflt int) int {
	return session.Read(key, dflt).(int)
}

func (session *Session) WriteInt32(key string, value int32) {
	session.transaction[key] = value
}

func (session *Session) ReadInt32(key string, dflt int32) int32 {
	return session.Read(key, dflt).(int32)
}

func (session *Session) WriteInt64(key string, value int64) {
	session.transaction[key] = value
}

func (session *Session) ReadInt64(key string, dflt int64) int64 {
	return session.Read(key, dflt).(int64)
}

func (session *Session) WriteBool(key string, value bool) {
	session.transaction[key] = value
}

func (session *Session) ReadBool(key string, dflt bool) bool {
	return session.Read(key, dflt).(bool)
}

func (session *Session) WriteFloat32(key string, value float32) {
	session.transaction[key] = value
}

func (session *Session) ReadFloat32(key string, dflt float32) float32 {
	return session.Read(key, dflt).(float32)
}

func (session *Session) WriteFloat64(key string, value float64) {
	session.transaction[key] = value
}

func (session *Session) ReadFloat64(key string, dflt float64) float64 {
	return session.Read(key, dflt).(float64)
}

func (session *Session) Commit() {
	session.WriteInt64("_compassLastUpdate", time.Now().UnixMilli())

	// Write
	toWrite, err := json.Marshal(session.transaction)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(
		fmt.Sprintf("%s%c%s.json",
			session.Server.SessionDirectory,
			filepath.Separator,
			UUIDToString(session.ID)), toWrite, 0644)
	if err != nil {
		panic(err)
	}

	// Prepare reading
	file, _ := os.Open(
		fmt.Sprintf("%s%c%s.json",
			session.Server.SessionDirectory,
			filepath.Separator,
			UUIDToString(session.ID)))

	defer file.Close()

	read, _ := io.ReadAll(file)
	vars := make(map[string]interface{})
	err = json.Unmarshal(read, &vars)
	if err != nil {
		panic(err)
	}

	session.transaction = make(map[string]interface{})
}

func (session *Session) Encrypt() string {
	hashedKey := sha256.Sum256([]byte(*session.Server.sessionSecret))
	block, err := aes.NewCipher(hashedKey[:])
	if err != nil {
		panic(err)
	}

	plainBytes := []byte(UUIDToString(session.ID))
	cipherText := make([]byte, aes.BlockSize+len(plainBytes))

	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainBytes)

	return base64.StdEncoding.EncodeToString(cipherText)
}

func (session *Session) ResetTransaction() {
	session.transaction = make(map[string]interface{})
}

func GetSessionById(server *Server, id string) *Session {
	newId := DecryptSessionID(server, id)
	path := fmt.Sprintf("%s%c%s.json", server.SessionDirectory, filepath.Separator, UUIDToString(newId))

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var sessionData map[string]interface{}

	if err := json.Unmarshal(data, &sessionData); err != nil {
		panic(err)
	}

	newSession := &Session{
		Server:      server,
		ID:          newId,
		transaction: make(map[string]interface{}),
		vars:        sessionData,
	}

	return newSession
}

func DecryptSessionID(server *Server, id string) UUID {
	hashedKey := sha256.Sum256([]byte(*server.sessionSecret))
	block, err := aes.NewCipher(hashedKey[:])
	if err != nil {
		panic(err)
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		panic(err)
	}

	if len(cipherBytes) < aes.BlockSize {
		panic(errors.New("cipherText too short"))
	}

	iv := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherBytes, cipherBytes)

	uuidStr := string(cipherBytes)

	uuid, err := StringToUUID(uuidStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert decrypted string to UUID: %w", err))
	}

	return uuid
}
