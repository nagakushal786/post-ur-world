package main

import (
	"net/http"

	"github.com/nagakushal786/post-ur-world/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, req *http.Request){
	// pagination, filters, sorting
	// /feed?limit=10&offset=10&sort=asc
	fq:=store.PaginatedFeedQuery{
		Limit: 20,
		Offset: 0,
		Sort: "desc",
		Tags: []string{},
		Search: "",
	}

	fq, err:=fq.Parse(req)
	if err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	if err:=Validate.Struct(fq); err!=nil{
		app.badRequestError(w, req, err)
		return
	}

	ctx:=req.Context()

	feed, err:=app.store.Posts.GetUserFeed(ctx, int64(1), fq)
	if err!=nil{
		app.internalServerError(w, req, err)
		return
	}

	if err:=app.jsonResponse(w, http.StatusOK, feed); err!=nil{
		app.internalServerError(w, req, err)
	}
}