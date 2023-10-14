package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
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
	}
	row := us.DB.QueryRow(`INSERT INTO users (id, email, email_normalized, password_hash)
  		VALUES ($1, $2, $3) RETURNING id`, ID, user.Email, user.EmailNormalized, passwordHash)
	err = row.Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}
