package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Gallery struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt *int64    `json:"updated_at"`
}

type GalleryService struct {
	DB *sql.DB
}

func (service *GalleryService) Create(title string, userID uuid.UUID) (*Gallery, error) {
	ID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("%s %w", "error creating uuid", err)
	}
	gallery := Gallery{
		ID:        ID,
		Title:     title,
		UserID:    userID,
		CreatedAt: time.Now().Unix(),
	}
	row := service.DB.QueryRow(`
		INSERT INTO galleries (id, title, user_id, created_at)
		VALUES ($1, $2, $3, $4) RETURNING id;`, gallery.ID, gallery.Title, gallery.UserID, gallery.CreatedAt)
	err = row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByID(id uuid.UUID) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}
	row := service.DB.QueryRow(`
		SELECT title, user_id, created_at, updated_at
		FROM galleries
		WHERE id = $1;`, gallery.ID)
	err := row.Scan(&gallery.Title, &gallery.UserID, &gallery.CreatedAt, &gallery.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("query gallery by id: %w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByUserID(userID uuid.UUID) ([]Gallery, error) {
	rows, err := service.DB.Query(`
		SELECT id, title, created_at, updated_at
		FROM galleries
		WHERE user_id = $1;`, userID)
	if err != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}
	var galleries []Gallery
	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}
		err := rows.Scan(&gallery.ID, &gallery.Title, &gallery.CreatedAt, &gallery.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("query galleries by user: %w", err)
		}
		galleries = append(galleries, gallery)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}
	return galleries, nil
}

func (service *GalleryService) Update(gallery *Gallery) error {
	_, err := service.DB.Exec(`
		UPDATE galleries
		SET title = $2, updated_at = $3
		WHERE id = $1;`, gallery.ID, gallery.Title, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}
	return nil
}

func (service *GalleryService) Delete(id uuid.UUID) error {
	_, err := service.DB.Exec(`
		DELETE FROM galleries
		WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("delete gallery by id: %w", err)
	}
	return nil
}
