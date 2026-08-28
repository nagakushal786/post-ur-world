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

// CreatePost godoc
//
// @Summary Creates a post
// @Description Creates a post
// @Tags posts
// @Accept json
// @Produce json
// @Param payload body CreatePostPayload true "Post Payload"
// @Success 201 {object} store.Post
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 500 {object} error
// @Security ApiKeyAuth
// @Router /posts [post]
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

	user:=getUserFromCtx(req)

	post:=&store.Post{
		Title: payload.Title,
		Content: payload.Content,
		UserID: user.ID,
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

// GetPost godoc
//
// @Summary Fetches a post
// @Description Fetches a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} store.Post
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 500 {object} error
// @Security ApiKeyAuth
// @Router /posts/{id} [get]
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

// DeletePost godoc
//
// @Summary Deletes a post
// @Description Deletes a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 204 {object} string
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Security ApiKeyAuth
// @Router /posts/{id} [delete]
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

// UpdatePost godoc
//
// @Summary Updates a post
// @Description Updates a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param payload body UpdatePostPayload true "Post Payload"
// @Success 200 {object} store.Post
// @Failure 400 {object} error
// @Failure 401 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Security ApiKeyAuth
// @Router /posts/{id} [patch]
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