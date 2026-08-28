package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound=errors.New("Resource not found")
	ErrConflict=errors.New("Resource already exists")
	QueryTimeoutDuration=time.Second*5
)

type Store struct{
	Posts interface{
		GetByID(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		DeleteByID(context.Context, int64) error
		UpdatePost(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface{
		Create(context.Context, *sql.Tx, *User) error
		GetByID(context.Context, int64) (*User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
		GetByEmail(context.Context, string) (*User, error)
	}
	Comments interface{
		GetByPostID(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}
	Followers interface{
		Follow(context.Context, int64, int64) error
		Unfollow(context.Context, int64, int64) error
	}
	Roles interface{
		GetByName(context.Context, string) (*Role, error)
	}
}

func NewPostgresStore(db *sql.DB) Store{
	return Store{
		Posts: &PostStore{db},
		Users: &UserStore{db},
		Comments: &CommentStore{db},
		Followers: &FollowerStore{db},
		Roles: &RoleStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error{
	tx, err:=db.BeginTx(ctx, nil)
	if err!=nil{
		return err
	}

	if err:=fn(tx); err!=nil{
		_=tx.Rollback()
		return err
	}

	return tx.Commit()
}