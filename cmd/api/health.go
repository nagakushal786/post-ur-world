package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, req *http.Request){
	data:=map[string]string{
		"status": "ok",
		"env": "Encrypted",
		"version": "0.0.1",
	}

	if err:=app.jsonResponse(w, http.StatusOK, data); err!=nil{
		app.internalServerError(w, req, err)
	}
}