package auth

import (
	"context"
	"database/sql"
	"errors"
	"monitoring-system/src/internal/modules/user-manager/domain/auth"
	"monitoring-system/src/pkg/app_error"
	"monitoring-system/src/pkg/logger"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

type authRepository struct {
	sqlDB  *sql.DB
	logger logger.Logger
}

func NewAuthRepository(ctx context.Context, db *sql.DB, logger logger.Logger) (auth.AuthRepository, error) {
	_, err := db.ExecContext(ctx, `
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

func (a *authRepository) GetByUsername(ctx context.Context, username string) (*auth.AuthEntity, error) {
	var entity auth.AuthEntity
	var id string

	err := a.sqlDB.QueryRowContext(ctx, "SELECT id, username, password FROM users WHERE username = ?", username).Scan(&id, &entity.Username, &entity.Password)
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

func (a *authRepository) Save(ctx context.Context, username, password string) error {
	id := uuid.New()
	_, err := a.sqlDB.ExecContext(ctx, "INSERT INTO users (id, username, password) VALUES (?, ?, ?)", id, username, password)
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
