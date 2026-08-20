package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound=errors.New("Resource not found")
	QueryTimeoutDuration=time.Second*5
)

type Store struct{
	Posts interface{
		GetByID(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		DeleteByID(context.Context, int64) error
		UpdatePost(context.Context, *Post) error
	}
	Users interface{
		Create(context.Context, *User) error
	}
	Comments interface{
		GetByPostID(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}
}

func NewPostgresStore(db *sql.DB) Store{
	return Store{
		Posts: &PostStore{db},
		Users: &UserStore{db},
		Comments: &CommentStore{db},
	}
}