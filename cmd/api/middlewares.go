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