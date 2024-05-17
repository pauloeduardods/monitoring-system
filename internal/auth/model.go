package auth

import (
	"database/sql"
	"errors"
	"monitoring-system/pkg/app_error"
	"monitoring-system/pkg/logger"

	"github.com/mattn/go-sqlite3"
)

type AuthEntity struct {
	ID       int
	Username string
	Password string
}

type AuthRepository interface {
	GetByUsername(username string) (*AuthEntity, error)
	Save(username, password string) error
}

type AuthModel struct {
	sqlDB  *sql.DB
	logger logger.Logger
}

func NewAuthRepository(db *sql.DB, logger logger.Logger) (AuthRepository, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL
		)
	`)
	if err != nil {
		logger.Error("Error creating users table: %v", err)
		return nil, err
	}

	return &AuthModel{sqlDB: db, logger: logger}, nil
}

func (a *AuthModel) GetByUsername(username string) (*AuthEntity, error) {
	var entity AuthEntity
	err := a.sqlDB.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&entity.ID, &entity.Username, &entity.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app_error.NewApiError(404, "User not found")
		}
		a.logger.Error("Error querying user by username: %v, error: %v", username, err)
		return nil, err
	}

	return &entity, nil
}

func (a *AuthModel) Save(username, password string) error {
	_, err := a.sqlDB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
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
