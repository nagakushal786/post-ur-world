package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Post struct{
	ID int64 `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Content string `json:"content"`
	Title string `json:"title"`
	UserID int64 `json:"user_id"`
	Tags []string `json:"tags"`
}

type PostStore struct{
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error{
	query:=`
	  insert into posts (content, title, user_id, tags)
	  values ($1, $2, $3, $4)
	  returning id, created_at, updated_at
	`

	err:=s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err!=nil{
		return err
	}

	return nil
}