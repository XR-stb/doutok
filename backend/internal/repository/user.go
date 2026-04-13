package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/xiaoran/doutok/internal/model"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, username, phone, password, nickname, avatar, bio, gender, status, role, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	_, err := DB.ExecContext(ctx, query,
		user.ID, user.Username, user.Phone, user.Password, user.Nickname,
		user.Avatar, user.Bio, user.Gender, user.Status, user.Role, now, now)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, username, COALESCE(phone,''), COALESCE(email,''), nickname, avatar, bio, gender, status, role, 
              follow_count, fan_count, like_count, video_count, created_at, updated_at 
              FROM users WHERE id = ? AND status != 3`
	user := &model.User{}
	err := DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Phone, &user.Email, &user.Nickname,
		&user.Avatar, &user.Bio, &user.Gender, &user.Status, &user.Role,
		&user.FollowCount, &user.FanCount, &user.LikeCount, &user.VideoCount,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, username, COALESCE(phone,''), COALESCE(email,''), password, nickname, avatar, bio, gender, status, role,
              follow_count, fan_count, like_count, video_count, created_at, updated_at
              FROM users WHERE username = ? AND status != 3`
	user := &model.User{}
	err := DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Phone, &user.Email, &user.Password, &user.Nickname,
		&user.Avatar, &user.Bio, &user.Gender, &user.Status, &user.Role,
		&user.FollowCount, &user.FanCount, &user.LikeCount, &user.VideoCount,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	query := `SELECT id, username, COALESCE(phone,''), COALESCE(email,''), password, nickname, avatar, bio, gender, status, role,
              follow_count, fan_count, like_count, video_count, created_at, updated_at
              FROM users WHERE phone = ? AND status != 3`
	user := &model.User{}
	err := DB.QueryRowContext(ctx, query, phone).Scan(
		&user.ID, &user.Username, &user.Phone, &user.Email, &user.Password, &user.Nickname,
		&user.Avatar, &user.Bio, &user.Gender, &user.Status, &user.Role,
		&user.FollowCount, &user.FanCount, &user.LikeCount, &user.VideoCount,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepo) Update(ctx context.Context, id int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	query := "UPDATE users SET "
	args := make([]interface{}, 0, len(fields)+1)
	first := true
	for k, v := range fields {
		if !first {
			query += ", "
		}
		query += fmt.Sprintf("%s = ?", k)
		args = append(args, v)
		first = false
	}
	query += " WHERE id = ?"
	args = append(args, id)
	_, err := DB.ExecContext(ctx, query, args...)
	return err
}

func (r *UserRepo) IncrCounter(ctx context.Context, id int64, field string, delta int64) error {
	query := fmt.Sprintf("UPDATE users SET %s = %s + ? WHERE id = ?", field, field)
	_, err := DB.ExecContext(ctx, query, delta, id)
	return err
}
