package main

import (
	"log"

	"github.com/nagakushal786/post-ur-world/internal/db"
	"github.com/nagakushal786/post-ur-world/internal/store"
)

func main(){
	db_url:="postgres://postgres:kushal786@localhost/post_ur_world?sslmode=disable"

	conn, err:=db.New(db_url, 3, 3, "15m")
	if err!=nil{
		log.Fatal(err)
	}

	defer conn.Close()

	store:=store.NewPostgresStore(conn)
	db.Seed(store)
}