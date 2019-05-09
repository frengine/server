package main

import (
	"log"
	"net/http"
	"os"

	"github.com/frengine/server/handler"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Use(handler.LoggerWare)

	deps := handler.Deps{
		nil,
		log.New(os.Stdout, "", log.Ldate|log.Ltime),
		log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile),
	}

	r.Handle("/auth/login", handler.LoginHandler{deps}).Methods("POST")
	r.Handle("/auth/register", handler.RegisterHandler{deps}).Methods("POST")

	srv := http.Server{
		Addr:    ":8083",
		Handler: r,

		// TODO: sane timeouts
	}

	deps.LogInfo.Println("Started")

	deps.LogErr.Fatal(srv.ListenAndServe())
}
