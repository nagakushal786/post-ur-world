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

func (app *application) getUserHandler(w http.ResponseWriter, req *http.Request){
	user:=getUserFromCtx(req)

	if err:=app.jsonResponse(w, http.StatusOK, user); err!=nil{
		app.internalServerError(w, req, err)
	}
}

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