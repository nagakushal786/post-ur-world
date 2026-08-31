package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Store{
	return Store{
		Users: &MockUserStorage{},
	}
}

type MockUserStorage struct{}

func (m *MockUserStorage) Create(ctx context.Context, tx *sql.Tx, user *User) error{
	return nil
}

func (m *MockUserStorage) GetByID(ctx context.Context, userID int64) (*User, error){
	return &User{}, nil
}

func (m *MockUserStorage) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error{
	return nil
}

func (m *MockUserStorage) Activate(ctx context.Context, token string) error{
	return nil
}
func (m *MockUserStorage) Delete(ctx context.Context, userID int64) error{
	return nil
}

func (m *MockUserStorage) GetByEmail(ctx context.Context, email string) (*User, error){
	return &User{}, nil
}