package store

import (
	"context"
	"database/sql"
)

type User struct{
	ID int64 `json:"id"`
	Username string `json:"username"`
	Email string `json:"email"`
	Password string `json:"-"`
	CreatedAt string `json:"created_at"`
}

type UserStore struct{
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, user *User) error{
	query:=`
	  insert into users (username, password, email)
	  values ($1, $2, $3)
	  returning id, created_at
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err:=s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err!=nil{
		return err
	}

	return nil
}