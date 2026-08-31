package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nagakushal786/post-ur-world/internal/store"
)

type postKey string
const postCtx postKey="post"

type userKey string
const userCtx userKey="user"

func (app *application) postsContextMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter, req *http.Request){
		idParam:=chi.URLParam(req, "postID")
		id, err:=strconv.ParseInt(idParam, 10, 64)
		if err!=nil{
			app.internalServerError(w, req, err)
			return
		}

		ctx:=req.Context()
		post, err:=app.store.Posts.GetByID(ctx, id)
		if err!=nil{
			switch{
				case errors.Is(err, store.ErrNotFound):
					app.notFoundError(w, req, err)
				default:
					app.internalServerError(w, req, err)
			}
			return
		}

		ctx=context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (app *application) usersContextMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		userID, err:=strconv.ParseInt(chi.URLParam(req, "userID"), 10, 64)
		if err!=nil{
			app.badRequestError(w, req, err)
			return
		}

		ctx:=req.Context()
		user, err:=app.store.Users.GetByID(ctx, userID)
		if err!=nil{
			switch err{
				case store.ErrNotFound:
					app.notFoundError(w, req, err)
				default:
					app.internalServerError(w, req, err)
			}
			return
		}

		ctx=context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler{
	return func(next http.Handler) http.Handler{
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
			// read the auth header
			authHeader:=req.Header.Get("Authorization")
			if authHeader==""{
				app.unAuthorizationBasicError(w, req, fmt.Errorf("Authorization header is missing"))
				return
			}

			// Authorization: Basic <token>
			// parse it to base64 string
			parts:=strings.Split(authHeader, " ")
			if len(parts)!=2 || parts[0]!="Basic"{
				app.unAuthorizationBasicError(w, req, fmt.Errorf("Authorization header is malformed"))
				return
			}

			// decode it
			decoded, err:=base64.StdEncoding.DecodeString(parts[1])
			if err!=nil{
				app.unAuthorizationBasicError(w, req, err)
				return
			}

			username:=app.config.auth.basic.username
			password:=app.config.auth.basic.password

			// check the credentials
			creds:=strings.SplitN(string(decoded), ":", 2)
			if len(creds)!=2 || creds[0]!=username || creds[1]!=password{
				app.unAuthorizationBasicError(w, req, fmt.Errorf("Invalid credentials"))
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		// read the auth header
		authHeader:=req.Header.Get("Authorization")
		if authHeader==""{
			app.unAuthorizationError(w, req, fmt.Errorf("Authorization header is missing"))
			return
		}

		// Authorization: Bearer <token>
		// parse it to base64 string
		parts:=strings.Split(authHeader, " ")
		if len(parts)!=2 || parts[0]!="Bearer"{
			app.unAuthorizationError(w, req, fmt.Errorf("Authorization header is malformed"))
			return
		}

		token:=parts[1]
		jwtToken, err:=app.authenticator.ValidateToken(token)
		if err!=nil{
			app.unAuthorizationError(w, req, err)
			return
		}

		claims, _:=jwtToken.Claims.(jwt.MapClaims)

		userID, err:=strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err!=nil{
			app.unAuthorizationError(w, req, err)
			return
		}

		ctx:=req.Context()
		user, err:=app.getUser(ctx, userID)
		if err!=nil{
			app.unAuthorizationError(w, req, err)
			return
		}

		ctx=context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc{
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		user:=getUserFromCtx(req)
		post:=getPostFromCtx(req)

		// If it is the user's post
		if post.UserID==user.ID{
			next.ServeHTTP(w, req)
			return
		}

		// Role precedence check
		allowed, err:=app.checkRolePrecedence(req.Context(), user, requiredRole)
		if err!=nil{
			app.internalServerError(w, req, err)
			return
		}

		if !allowed{
			app.forbiddenResponse(w, req)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error){
	role, err:=app.store.Roles.GetByName(ctx, roleName)
	if err!=nil{
		return false, err
	}

	return user.Role.Level>=role.Level, nil
}

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error){
	if !app.config.redisCfg.enabled{
		return app.store.Users.GetByID(ctx, userID)
	}
	
	// Caching
	user, err:=app.cacheStorage.Users.Get(ctx, userID)
	if err!=nil{
		return nil, err
	}

	if user!=nil{
		app.logger.Infow("Cache hit", "key", "user", "id", userID)
		return user, nil
	}

	app.logger.Infow("Fetching from DB", "id", userID)
	user, err=app.store.Users.GetByID(ctx, userID)
	if err!=nil{
		return nil, err
	}

	if err:=app.cacheStorage.Users.Set(ctx, user); err!=nil{
		return nil, err
	}

	return user, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		if app.config.rateLimiterCfg.Enabled{
			if allow, retryAfter:=app.rateLimiter.Allow(req.RemoteAddr); !allow{
				app.rateLimitExceededResponse(w, req, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, req)
	})
}