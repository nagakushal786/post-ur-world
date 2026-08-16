package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/nagakushal786/post-ur-world/internal/db"
	"github.com/nagakushal786/post-ur-world/internal/store"
)

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
	
	cfg:=config{
		addr: ":"+portString,
		db: dbConfig{
			addr: db_url,
			maxOpenConns: max_open_coons,
			maxIdleConns: max_idle_conns,
			maxIdleTime: max_idle_time,
		},
	}

	db, err:=db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err!=nil{
		log.Fatal(err)
	}

	defer db.Close()
	log.Println("Database connection pool established")

	store:=store.NewPostgresStore(db)

	app:=&application{
		config: cfg,
		store: store,
	}

	router:=app.mount()
	log.Fatal(app.run(router))
}