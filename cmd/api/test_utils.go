package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nagakushal786/post-ur-world/internal/auth"
	"github.com/nagakushal786/post-ur-world/internal/store"
	"github.com/nagakushal786/post-ur-world/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application{
	t.Helper()

	// logger:=zap.NewNop().Sugar()
	logger:=zap.Must(zap.NewProduction()).Sugar()
	mockStore:=store.NewMockStore()
	mockCacheStore:=cache.NewMockStore()
	testAuth:=&auth.TestAuthenticator{}

	return &application{
		logger: logger,
		store: mockStore,
		cacheStorage: mockCacheStore,
		authenticator: testAuth,
		config: cfg,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder{
	rr:=httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int){
	if actual!=expected{
		t.Errorf("Expected the response code to be %d and we got %d", expected, actual)
	}
}