package main

import (
	"net/http"
)

// HealthChecker godoc
//
// @Summary Health check
// @Description Health check endpoint
// @Tags health
// @Produce json
// @Success 200 {object} string "ok"
// @Router /health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, req *http.Request){
	data:=map[string]string{
		"status": "ok",
		"env": app.config.env,
		"version": "0.0.1",
	}

	if err:=app.jsonResponse(w, http.StatusOK, data); err!=nil{
		app.internalServerError(w, req, err)
	}
}