package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AguilaMike/lenslocked/pkg/app/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              uuid.UUID       `json:"id"`
	Email           string          `json:"email"`
	EmailNormalized string          `json:"email_normalized"`
	PasswordHash    string          `json:"password"`
	IsAdmin         bool            `json:"is_admin"`
	Details         json.RawMessage `json:"details"`
	CreatedAt       int64           `json:"created_at"`
	UpdatedAt       int64           `json:"updated_at"`
	DeletedAt       int64           `json:"deleted_at"`
}

type UserService struct {
	DB *sql.DB
}

func (us *UserService) Create(email, password string) (*User, error) {
	ID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("%s %w", "error creating uuid", err)
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	passwordHash := string(hashedBytes)

	user := User{
		ID:              ID,
		Email:           email,
		EmailNormalized: strings.ToLower(email),
		PasswordHash:    passwordHash,
		CreatedAt:       time.Now().Unix(),
	}
	row := us.DB.QueryRow(`INSERT INTO users (id, email, email_normalized, password_hash, created_at)
  		VALUES ($1, $2, $3, $4, $5) RETURNING id;`, ID, user.Email, user.EmailNormalized, passwordHash, user.CreatedAt)
	err = row.Scan(&user.ID)
	if err != nil {
		// See if we can use this error as a PgError
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			// This is a PgError, so see if it matches a unique violation.
			if pgError.Code == pgerrcode.UniqueViolation {
				// If this is true, it has to be an email violation since this is the
				// only way to trigger this type of violation with our SQL.
				err = ErrEmailTaken
			}
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func (us UserService) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)
	user := User{
		Email: email,
	}
	row := us.DB.QueryRow(`
		SELECT id, password_hash
		FROM users WHERE email=$1`, email)
	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotFound
		}
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", ErrPasswordError)
	}
	user.PasswordHash = ""
	return &user, nil
}

func (us *UserService) UpdatePassword(userID uuid.UUID, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	passwordHash := string(hashedBytes)
	_, err = us.DB.Exec(`
	  UPDATE users
		SET password_hash = $2
		WHERE id = $1;`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
