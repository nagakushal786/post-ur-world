package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nagakushal786/post-ur-world/internal/store"
)

type application struct{
	config config
	store store.Store
}

type config struct{
	addr string
	db dbConfig
}

type dbConfig struct{
	addr string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime string
}

func (app *application) mount() http.Handler{
	router:=chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.ClientIPFromRemoteAddr)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middleware.Timeout(60*time.Second))
	
	router.Route("/v1", func(r chi.Router){
		r.Get("/health", app.healthCheckHandler)

		// /v1/posts
		r.Route("/posts", func(r chi.Router){
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func (r chi.Router){
				r.Use(app.postsContextMiddleware)
				
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
			})
		})
	})

	return router
} 

func (app *application) run(router http.Handler) error{
	srv:=&http.Server{
		Addr: app.config.addr,
		Handler: router,
		WriteTimeout: time.Second*30,
		ReadTimeout: time.Second*10,
		IdleTimeout: time.Minute,
	}

	log.Printf("Server has started at port %s", app.config.addr)

	return srv.ListenAndServe()
}