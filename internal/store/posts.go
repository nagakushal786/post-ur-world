package store

import (
	"context"
	"database/sql"
	"errors"

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
	Version int `json:"version"`
	Comments []Comment `json:"comments"`
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

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

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

func (s *PostStore) GetByID(ctx context.Context, postID int64) (*Post, error){
	query:=`
	  select id, user_id, title, content, created_at, updated_at, tags, version
	  from posts where id=$1
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err:=s.db.QueryRowContext(
		ctx,
		query,
		postID,
	).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err!=nil{
		switch{
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrNotFound
			default:
				return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) DeleteByID(ctx context.Context, postID int64) error{
	query:=`delete from posts where id=$1;`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err:=s.db.ExecContext(ctx, query, postID)
	if err!=nil{
		return err
	}

	rows, err:=result.RowsAffected()
	if err!=nil{
		return err
	}

	if rows==0{
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) UpdatePost(ctx context.Context, post *Post) error{
	query:=`
	  update posts
	  set title=$1, content=$2, version=version+1
	  where id=$3 and version=$4
	  returning version;
	`

	ctx, cancel:=context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err:=s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID, 
		post.Version,
	).Scan(&post.Version)

	if err!=nil{
		switch{
			case errors.Is(err, sql.ErrNoRows):
				return ErrNotFound
			default:
				return err
		}
	}

	return nil
}