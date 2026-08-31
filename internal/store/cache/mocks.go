package cache

import (
	"context"

	"github.com/nagakushal786/post-ur-world/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockStore() Storage{
	return Storage{
		Users: &MockUserStorage{},
	}
}

type MockUserStorage struct{
	mock.Mock
}

func (m *MockUserStorage) Get(ctx context.Context, userID int64) (*store.User, error){
	args:=m.Called(userID)
	return nil, args.Error(1)
}

func (m *MockUserStorage) Set(ctx context.Context, user *store.User) error{
	args:=m.Called(user)
	return args.Error(0)
}