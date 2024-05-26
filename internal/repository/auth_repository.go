package repository

import (
	"database/sql"
	"errors"
	"monitoring-system/domain/auth"
	"monitoring-system/pkg/app_error"
	"monitoring-system/pkg/logger"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

type authRepository struct {
	sqlDB  *sql.DB
	logger logger.Logger
}

func NewAuthRepository(db *sql.DB, logger logger.Logger) (auth.AuthRepository, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id       VARCHAR(36) PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL
		)
	`)
	if err != nil {
		logger.Error("Error creating users table: %v", err)
		return nil, err
	}

	return &authRepository{sqlDB: db, logger: logger}, nil
}

func (a *authRepository) GetByUsername(username string) (*auth.AuthEntity, error) {
	var entity auth.AuthEntity
	var id string

	err := a.sqlDB.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&id, &entity.Username, &entity.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app_error.NewApiError(404, "User not found")
		}
		a.logger.Error("Error", err)
		a.logger.Error("Error querying user by username: %v, error: %v", username, err)
		return nil, err
	}

	entity.ID, err = uuid.Parse(id)
	if err != nil {
		a.logger.Error("Error parsing user id: %v", err)
		return nil, err
	}

	return &entity, nil
}

func (a *authRepository) Save(username, password string) error {
	id := uuid.New()
	_, err := a.sqlDB.Exec("INSERT INTO users (id, username, password) VALUES (?, ?, ?)", id, username, password)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return app_error.NewApiError(409, "Username already exists")
			}
		}
		a.logger.Error("Error saving user: %v", err)
		return err
	}
	return nil
}
