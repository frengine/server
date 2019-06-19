package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/frengine/server/auth"
	"github.com/frengine/server/config"
	"github.com/frengine/server/handler"
	"github.com/frengine/server/project"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.ParseFromFile("config.json")
	if err != nil {
		if err == config.ErrFileNotExists {
			err := config.WriteDefault("config.json")
			if err != nil {
				log.Println("Config file not found. Could not generate it either.")
				log.Fatal(err)
				return
			}
			fmt.Println("Config file created as config.json")
			fmt.Println("Please configure it and restart the program")
			return
		}
		log.Println("Cannot parse config file. Is it valid JSON?")
		log.Fatal(err)
		return
	}

	db, err := sql.Open("postgres", cfg.MakeDBString())
	if err != nil {
		log.Println("Cannot connect to the database:")
		log.Fatal(err)
		return
	}
	err = db.Ping()
	if err != nil {
		log.Println("Cannot communicate with the database:")
		log.Fatal(err)
		return
	}

	deps := handler.Deps{
		auth.PostgresStore{db},
		project.PostgresStore{db},
		log.New(os.Stdout, "", log.Ldate|log.Ltime),
		log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Llongfile),
		cfg,
	}

	r := mux.NewRouter()
	r.Use((handler.LoggerWare{deps}).Middleware)

	{
		s := r.PathPrefix("/auth").Subrouter()

		s.Handle("/login", handler.LoginHandler{deps}).Methods("POST")
		s.Handle("/register", handler.RegisterHandler{deps}).Methods("POST")
	}

	{
		s := r.PathPrefix("/textvcs").Subrouter()
		s.Use((handler.AuthWare{deps}).Middleware)

		s.Handle("", handler.TestHandler{deps}).Methods("POST")
	}

	{
		s := r.PathPrefix("/projects").Subrouter()
		s.Use((handler.AuthWare{deps}).Middleware)

		s.Handle("", handler.ProjectListHandler{deps}).Methods("GET")
		s.Handle("", handler.ProjectCreateHandler{deps}).Methods("POST")
		s.Handle("/{id}", handler.ProjectUpdateHandler{deps}).Methods("PUT")
	}

	srv := http.Server{
		Addr:    ":8083",
		Handler: r,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	deps.LogInfo.Println("Started")

	deps.LogErr.Fatal(srv.ListenAndServe())
}
