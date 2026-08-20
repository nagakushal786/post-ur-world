package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nagakushal786/post-ur-world/internal/store"
)

type CreatePostPayload struct{
	Title string `json:"title" validate:"required,max=100"`
	Content string `json:"content" validate:"required,max=1000"`
	Tags []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, req *http.Request){
	var payload CreatePostPayload

	if err:=readJSON(w, req, &payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	if err:=Validate.Struct(payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	post:=&store.Post{
		Title: payload.Title,
		Content: payload.Content,
		UserID: 1,
		Tags: payload.Tags,
	}

	if err:=app.store.Posts.Create(req.Context(), post); err!=nil{
		app.internalServerError(w, req, err)
		return
	}

	if err:=app.jsonResponse(w, http.StatusCreated, post); err!=nil{
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, req *http.Request){
	post:=getPostFromCtx(req)

	comments, err:=app.store.Comments.GetByPostID(req.Context(), post.ID)
	if err!=nil{
		app.internalServerError(w, req, err)
		return
	}

	post.Comments=comments

	if err:=app.jsonResponse(w, http.StatusOK, post); err!=nil{
		app.internalServerError(w, req, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, req *http.Request){
	idParam:=chi.URLParam(req, "postID")
	id, err:=strconv.ParseInt(idParam, 10, 64)
	if err!=nil{
		app.internalServerError(w, req, err)
		return
	}

	err=app.store.Posts.DeleteByID(req.Context(), id)
	if err!=nil{
		switch{
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, req, err)
			default:
				app.internalServerError(w, req, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getPostFromCtx(req *http.Request) *store.Post{
	post, _:=req.Context().Value(postCtx).(*store.Post)
	return post
}

type UpdatePostPayload struct{
	Title *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, req *http.Request){
	post:=getPostFromCtx(req)

	var payload UpdatePostPayload
	if err:=readJSON(w, req, &payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	if err:=Validate.Struct(payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	if payload.Content!=nil{
		post.Content=*payload.Content
	}

	if payload.Title!=nil{
		post.Title=*payload.Title
	}

	err:=app.store.Posts.UpdatePost(req.Context(), post)
	if err!=nil{
		switch{
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, req, err)
			default:
				app.internalServerError(w, req, err)
		}
		return
	}

	if err:=app.jsonResponse(w, http.StatusOK, post); err!=nil{
		app.internalServerError(w, req, err)
	}
}