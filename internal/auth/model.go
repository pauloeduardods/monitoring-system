package auth

import (
	"database/sql"
	"monitoring-system/pkg/app_error"
	"monitoring-system/pkg/logger"
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
	res, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	logger.Info("Table users created %v", res)
	return &AuthModel{sqlDB: db, logger: logger}, nil
}

func (a *AuthModel) GetByUsername(username string) (*AuthEntity, error) {
	var entity AuthEntity
	err := a.sqlDB.QueryRow("SELECT * FROM users WHERE username = ?", username).Scan(&entity.ID, &entity.Username, &entity.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app_error.NewApiError(404, "User not found")
		}
		return nil, err
	}

	return &entity, nil
}

func (a *AuthModel) Save(username, password string) error {
	_, err := a.sqlDB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
	return err
}
