package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nagakushal786/post-ur-world/internal/db"
	"github.com/nagakushal786/post-ur-world/internal/store"
)

func main(){
	godotenv.Load()
	db_url:=os.Getenv("DB_URL")
	if db_url==""{
		log.Fatal("DB_URL is not found in environment")
	}

	conn, err:=db.New(db_url, 3, 3, "15m")
	if err!=nil{
		log.Fatal(err)
	}

	defer conn.Close()

	store:=store.NewPostgresStore(conn)
	db.Seed(store, conn)
}