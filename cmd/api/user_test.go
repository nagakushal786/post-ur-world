package main

import (
	"net/http"
	"testing"

	"github.com/nagakushal786/post-ur-world/internal/store/cache"
	"github.com/stretchr/testify/mock"
)

func TestGetUser(t *testing.T){
	t.Run("should not allow unauthenticated requests", func(t *testing.T){
		withRedis:=config{
			redisCfg: redisConfig{
				enabled: true,
			},
		}

		app:=newTestApplication(t, withRedis)
		mux:=app.mount()

		req, err:=http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err!=nil{
			t.Fatal(err)
		}

		rr:=executeRequest(req, mux)

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow authenticated requests", func(t *testing.T){
		withRedis:=config{
			redisCfg: redisConfig{
				enabled: true,
			},
		}

		app:=newTestApplication(t, withRedis)
		mux:=app.mount()
		testToken, err:=app.authenticator.GenerateToken(nil)
		if err!=nil{
			t.Fatal(err)
		}

		mockCacheStore:=app.cacheStorage.Users.(*cache.MockUserStorage)

		// AuthMiddleware is defaulting the ID to 2 (found from the error)
		mockCacheStore.On("Get", int64(2)).Return(nil, nil)
		mockCacheStore.On("Get", int64(1)).Return(nil, nil)
		mockCacheStore.On("Set", mock.Anything).Return(nil)

		req, err:=http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err!=nil{
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr:=executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.Calls=nil
	})

	t.Run("should hit the cache first and if not exists it sets the user on the cache", func(t *testing.T){
		withRedis:=config{
			redisCfg: redisConfig{
				enabled: true,
			},
		}

		app:=newTestApplication(t, withRedis)
		mux:=app.mount()
		testToken, err:=app.authenticator.GenerateToken(nil)
		if err!=nil{
			t.Fatal(err)
		}

		mockCacheStore:=app.cacheStorage.Users.(*cache.MockUserStorage)

		mockCacheStore.On("Get", int64(2)).Return(nil, nil)
		mockCacheStore.On("Get", int64(1)).Return(nil, nil)
		mockCacheStore.On("Set", mock.Anything, mock.Anything).Return(nil)

		req, err:=http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err!=nil{
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr:=executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.AssertNumberOfCalls(t, "Get", 2)

		mockCacheStore.Calls=nil
	})

	t.Run("should not hit the cache if it is not enabled", func(t *testing.T){
		withRedis:=config{
			redisCfg: redisConfig{
				enabled: false,
			},
		}

		app:=newTestApplication(t, withRedis)
		mux:=app.mount()
		testToken, err:=app.authenticator.GenerateToken(nil)
		if err!=nil{
			t.Fatal(err)
		}

		mockCacheStore:=app.cacheStorage.Users.(*cache.MockUserStorage)

		req, err:=http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err!=nil{
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr:=executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.AssertNotCalled(t, "Get")

		mockCacheStore.Calls=nil
	})
}