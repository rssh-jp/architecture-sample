package repository

import (
	"context"
	"database/sql"
	"sample/domain"
)

type sqliteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository はSQLite用のユーザーリポジトリを生成します。
func NewSQLiteUserRepository(db *sql.DB) *sqliteUserRepository {
	return &sqliteUserRepository{db: db}
}

// FindByID は指定されたIDのユーザーを取得します。
func (r *sqliteUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, name, email FROM users WHERE id = $1`

	user := &domain.User{}
	err := getDB(ctx, r.db).QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Save はユーザーを保存し、採番されたIDをユーザーへ設定します。
func (r *sqliteUserRepository) Save(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email) VALUES ($1, $2)`
	result, err := getDB(ctx, r.db).ExecContext(ctx, query, user.Name, user.Email)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id

	return nil
}
