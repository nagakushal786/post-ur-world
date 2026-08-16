package store

import (
	"context"
	"database/sql"
)

type Store struct{
	Posts interface{
		Create(context.Context, *Post) error
	}
	Users interface{
		Create(context.Context, *User) error
	}
}

func NewPostgresStore(db *sql.DB) Store{
	return Store{
		Posts: &PostStore{db},
		Users: &UserStore{db},
	}
}