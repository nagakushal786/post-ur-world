package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nagakushal786/post-ur-world/docs"
	"github.com/nagakushal786/post-ur-world/internal/auth"
	"github.com/nagakushal786/post-ur-world/internal/mailer"
	"github.com/nagakushal786/post-ur-world/internal/store"
	"github.com/nagakushal786/post-ur-world/internal/store/cache"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct{
	config config
	store store.Store
	logger *zap.SugaredLogger
	mailer mailer.Client
	authenticator auth.Authenticator
	cacheStorage cache.Storage
}

type config struct{
	addr string
	db dbConfig
	env string
	apiURL string
	mail mailConfig
	frontendURL string
	auth authConfig
	redisCfg redisConfig
}

type redisConfig struct{
	addr string
	pw string
	db int
	enabled bool
}

type authConfig struct{
	basic basicConfig
	token tokenConfig
}

type basicConfig struct{
	username string
	password string
}

type tokenConfig struct{
	secret string
	exp time.Duration
	issuer string
	audience string
}

type mailConfig struct{
	exp time.Duration
	sendGrid sendGridConfig
	fromEmail string
}

type sendGridConfig struct{
	apiKey string
}

type dbConfig struct{
	addr string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime string
}

func (app *application) mount() http.Handler{
	router:=chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Use(middleware.RequestID)
	router.Use(middleware.ClientIPFromRemoteAddr)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middleware.Timeout(60*time.Second))
	
	router.Route("/v1", func(r chi.Router){
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)

		docsURL:=fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		// /v1/posts
		r.Route("/posts", func(r chi.Router){
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func (r chi.Router){
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
			})
		})

		// /v1/users
		r.Route("/users", func(r chi.Router){
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router){
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router){
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		// Public routes
		r.Route("/authentication", func(r chi.Router){
			r.Post("/register", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return router
} 

func (app *application) run(router http.Handler) error{
	// docs
	docs.SwaggerInfo.Version=version
	docs.SwaggerInfo.Host=app.config.apiURL
	docs.SwaggerInfo.BasePath="/v1"

	srv:=&http.Server{
		Addr: app.config.addr,
		Handler: router,
		WriteTimeout: time.Second*30,
		ReadTimeout: time.Second*10,
		IdleTimeout: time.Minute,
	}

	app.logger.Infow("Server has started", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}