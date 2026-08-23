package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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