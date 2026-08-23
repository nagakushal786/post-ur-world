package main

import (
	"net/http"

	"github.com/nagakushal786/post-ur-world/internal/store"
)

type FollowUser struct{
	UserID int64 `json:"user_id"`
}

func getUserFromCtx(req *http.Request) *store.User{
	user, _:=req.Context().Value(userCtx).(*store.User)
	return user
}

// GetUser godoc
//
// @Summary Fetches a user profile
// @Description Fetches a user profile by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} store.User
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Security ApiKeyAuth
// @Router /users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, req *http.Request){
	user:=getUserFromCtx(req)

	if err:=app.jsonResponse(w, http.StatusOK, user); err!=nil{
		app.internalServerError(w, req, err)
	}
}

// FollowUser godoc
//
// @Summary Follows a user
// @Description Follows a user profile by ID
// @Tags users
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Success 204 {string} string "User followed"
// @Failure 400 {object} error "User payload missing"
// @Failure 404 {object} error "User not found"
// @Security ApiKeyAuth
// @Router /users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, req *http.Request){
	followerUser:=getUserFromCtx(req)

	// Revert back to auth userID from ctx
	var payload FollowUser
	if err:=readJSON(w, req, &payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	ctx:=req.Context()
	err:=app.store.Followers.Follow(ctx, followerUser.ID, payload.UserID)
	if err!=nil{
		switch err{
			case store.ErrConflict:
				app.conflictError(w, req, err)
			default:
				app.internalServerError(w, req, err)
		}
		return
	}

	if err:=app.jsonResponse(w, http.StatusNoContent, nil); err!=nil{
		app.internalServerError(w, req, err)
	}
}

// UnfollowUser godoc
//
// @Summary Unfollows a user
// @Description Unfollows a user profile by ID
// @Tags users
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Success 204 {string} string "User unfollowed"
// @Failure 400 {object} error "User payload missing"
// @Failure 404 {object} error "User not found"
// @Security ApiKeyAuth
// @Router /users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, req *http.Request){
	unfollowedUser:=getUserFromCtx(req)

	// Revert back to auth userID from ctx
	var payload FollowUser
	if err:=readJSON(w, req, &payload); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	ctx:=req.Context()
	err:=app.store.Followers.Unfollow(ctx, unfollowedUser.ID, payload.UserID)
	if err!=nil{
		app.internalServerError(w, req, err)
		return
	}

	if err:=app.jsonResponse(w, http.StatusNoContent, nil); err!=nil{
		app.internalServerError(w, req, err)
	}
}