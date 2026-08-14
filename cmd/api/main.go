package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main(){
	godotenv.Load()

	portString:=os.Getenv("PORT")
	if portString==""{
		log.Fatal("PORT is not found in environment")
	}
	
	cfg:=config{
		addr: ":"+portString,
	}

	app:=&application{
		config: cfg,
	}

	router:=app.mount()
	log.Fatal(app.run(router))
}