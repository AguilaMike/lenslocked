package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/AguilaMike/lenslocked/pkg/internal/rand"
	"github.com/google/uuid"
)

const (
	// The minimum number of bytes to be used for each session token.
	MinBytesPerToken = 32
)

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
	// Token is only set when creating a new session. When looking up a session
	// this will be left empty, as we only store the hash of a session token
	// in our database and we cannot reverse it into a raw token.
	Token string `json:"-"`
}

type SessionService struct {
	DB *sql.DB
	// BytesPerToken is used to determine how many bytes to use when generating
	// each session token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be
	// used.
	BytesPerToken int
}

// Create will create a new session for the user provided. The session token
// will be returned as the Token field on the Session type, but only the hashed
// session token is stored in the database.
func (ss *SessionService) Create(userID uuid.UUID) (*Session, error) {
	tokenService := TokenManager{}
	token, tokenHash, err := tokenService.New()
	if err != nil {
		return nil, fmt.Errorf("%s %w", "error creating token", err)
	}
	ID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("%s %w", "error creating uuid", err)
	}
	session := Session{
		ID:        ID,
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now().Unix(),
		TokenHash: tokenHash,
	}

	row := ss.DB.QueryRow(`
		INSERT INTO sessions (id, user_id, token_hash, created_at)
		VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO
		UPDATE SET token_hash = $3, updated_at = $4
		RETURNING id;`, ID, userID, tokenHash, session.CreatedAt)
	err = row.Scan(&session.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	return &session, nil
}

func (ss *SessionService) User(token string) (*User, error) {
	tokenHash := TokenManager{}.Hash(token)
	var user User
	row := ss.DB.QueryRow(`
		SELECT users.id, users.email, users.password_hash
		  FROM users
	INNER JOIN sessions ON sessions.user_id = users.id
		 WHERE sessions.token_hash = $1;`, tokenHash)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	return &user, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := TokenManager{}.Hash(token)
	_, err := ss.DB.Exec(`DELETE FROM sessions WHERE token_hash = $1;`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

type TokenManager struct {
	// BytesPerToken is used to determine how many bytes to use when generating
	// each session token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be
	// used.
	BytesPerToken int
}

func (tm TokenManager) New() (string, string, error) {
	bytesPerToken := tm.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return "", "", fmt.Errorf("create: %w", err)
	}
	tokenHash := tm.getHash(token)
	return token, tokenHash, nil
}

func (tm TokenManager) Hash(token string) string {
	return tm.getHash(token)
}

func (tm TokenManager) getHash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	// base64 encode the data into a string
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
