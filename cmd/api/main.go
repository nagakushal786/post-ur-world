package main

import (
	"expvar"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/nagakushal786/post-ur-world/internal/auth"
	"github.com/nagakushal786/post-ur-world/internal/db"
	"github.com/nagakushal786/post-ur-world/internal/mailer"
	"github.com/nagakushal786/post-ur-world/internal/ratelimiter"
	"github.com/nagakushal786/post-ur-world/internal/store"
	"github.com/nagakushal786/post-ur-world/internal/store/cache"
	"go.uber.org/zap"
)

const version = "0.0.1"

// @title Post Ur World API
// @description API for post ur world, a platform to interact with fellow engineers
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description JWT Bearer token authentication

// @securityDefinitions.basic BasicAuth
// @description Basic authentication

func main(){
	godotenv.Load()

	portString:=os.Getenv("PORT")
	if portString==""{
		log.Fatal("PORT is not found in environment")
	}

	db_url:=os.Getenv("DB_URL")
	if db_url==""{
		log.Fatal("DB_URL is not found in environment")
	}

	max_open_coons, err:=strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if err!=nil{
		log.Fatal(err)
	}

	max_idle_conns, err:=strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if err!=nil{
		log.Fatal(err)
	}

	max_idle_time:=os.Getenv("DB_MAX_IDLE_TIME")
	if max_idle_time==""{
		log.Fatal("DB_MAX_IDLE_TIME is not found in environment")
	}

	env:=os.Getenv("ENV")
	if env==""{
		log.Fatal("ENV is not found in environment")
	}

	api_url:=os.Getenv("API_URL")
	if api_url==""{
		log.Fatal("API_URL is not found in environment")
	}

	sendGrid_api_key:=os.Getenv("SENDGRID_API_KEY")
	if sendGrid_api_key==""{
		log.Fatal("API Key is not found in environment")
	}

	from_email:=os.Getenv("FROM_EMAIL")
	if from_email==""{
		log.Fatal("FROM_EMAIL is not found in environment")
	}

	frontend_url:=os.Getenv("FRONTEND_URL")
	if frontend_url==""{
		log.Fatal("FRONTEND_URL is not found in environment")
	}

	auth_basic_user:=os.Getenv("AUTH_BASIC_USER")
	if auth_basic_user==""{
		log.Fatal("AUTH_BASIC_USER is not found in environment")
	}

	auth_basic_pass:=os.Getenv("AUTH_BASIC_PASS")
	if auth_basic_pass==""{
		log.Fatal("AUTH_BASIC_PASS is not found in environment")
	}

	auth_token_secret:=os.Getenv("AUTH_TOKEN_SECRET")
	if auth_token_secret==""{
		log.Fatal("AUTH_TOKEN_SECRET is not found in environment")
	}

	redis_addr:=os.Getenv("REDIS_ADDR")
	if redis_addr==""{
		log.Fatal("REDIS_ADDR is not found in environment")
	}

	redis_pw:=os.Getenv("REDIS_PW")

	redis_db, err:=strconv.Atoi(os.Getenv("REDIS_DB"))
	if err!=nil{
		log.Fatal("REDIS_DB is not found in environment")
	}

	redis_enabled, err:=strconv.ParseBool(os.Getenv("REDIS_ENABLED"))
	if err!=nil{
		log.Fatal("REDIS_ENABLED is not found in environment")
	}

	rl_requests_count, err:=strconv.Atoi(os.Getenv("RATELIMITER_REQUESTS_COUNT"))
	if err!=nil{
		log.Fatal("RATELIMITER_REQUESTS_COUNT is not found in environment")
	}

	rl_enabled, err:=strconv.ParseBool(os.Getenv("RATELIMITER_ENABLED"))
	if err!=nil{
		log.Fatal("RATELIMITER_ENABLED is not found in environment")
	}
		
	cfg:=config{
		addr: ":"+portString,
		db: dbConfig{
			addr: db_url,
			maxOpenConns: max_open_coons,
			maxIdleConns: max_idle_conns,
			maxIdleTime: max_idle_time,
		},
		apiURL: api_url,
		env: env,
		mail: mailConfig{
			exp: time.Hour*24*3, // 3 days
			sendGrid: sendGridConfig{
				apiKey: sendGrid_api_key,
			},
			fromEmail: from_email,
		},
		frontendURL: frontend_url,
		auth: authConfig{
			basic: basicConfig{
				username: auth_basic_user,
				password: auth_basic_pass,
			},
			token: tokenConfig{
				secret: auth_token_secret,
				exp: time.Hour*24*3,
				issuer: "posturworld",
				audience: "posturworld",
			},
		},
		redisCfg: redisConfig{
			addr: redis_addr,
			pw: redis_pw,
			db: redis_db,
			enabled: redis_enabled,
		},
		rateLimiterCfg: ratelimiter.Config{
			RequestsPerTimeFrame: rl_requests_count,
			TimeFrame: time.Second*5,
			Enabled: rl_enabled,
		},
	}

	// Logger
	logger:=zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err:=db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err!=nil{
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("Database connection pool established")

	// Cache
	var rdb *redis.Client
	if cfg.redisCfg.enabled{
		rdb=cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("Redis connection pool established")
	}

	// Rate Limiter
	rateLimiter:=ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiterCfg.RequestsPerTimeFrame,
		cfg.rateLimiterCfg.TimeFrame,
	)

	store:=store.NewPostgresStore(db)
	cacheStorage:=cache.NewRedisStorage(rdb)

	mailer:=mailer.NewSendGrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)

	jwtAuthenticator:=auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.issuer, cfg.auth.token.audience)

	app:=&application{
		config: cfg,
		store: store,
		logger: logger,
		mailer: mailer,
		authenticator: jwtAuthenticator,
		cacheStorage: cacheStorage,
		rateLimiter: rateLimiter,
	}

	// Metrics collected

	// Version
	expvar.NewString("version").Set(version)
	// Database
	expvar.Publish("database", expvar.Func(func() any{
		return db.Stats()
	}))
	// Number of go-routines
	expvar.Publish("goroutines", expvar.Func(func() any{
		return runtime.NumGoroutine()
	}))

	router:=app.mount()
	logger.Fatal(app.run(router))
}