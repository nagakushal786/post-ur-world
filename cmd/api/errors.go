package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, req *http.Request, err error){
	app.logger.Errorw("Internal server error", "method", req.Method, "path", req.URL.Path, "error", err.Error())

	writeError(w, http.StatusInternalServerError, "The server encountered an issue")
}

func (app *application) badRequestError(w http.ResponseWriter, req *http.Request, err error){
	// log.Printf("Bad request error: %s, path: %s, error: %s", req.Method, req.URL.Path, err.Error())

	app.logger.Warnf("Bad request error", "method", req.Method, "path", req.URL.Path, "error", err.Error())

	writeError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, req *http.Request, err error){
	app.logger.Warnf("Not found error", "method", req.Method, "path", req.URL.Path, "error", err.Error())

	writeError(w, http.StatusNotFound, "Not found")
}

func (app *application) conflictError(w http.ResponseWriter, req *http.Request, err error){
	app.logger.Errorf("Conflict error", "method", req.Method, "path", req.URL.Path, "error", err.Error())

	writeError(w, http.StatusConflict, err.Error())
}