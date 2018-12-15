package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"./dbcontroller"
)

const gracefullShutdownTimeout = time.Second * 5
const idleTimeout = time.Second * 60
const readTimeout = time.Second * 5
const writeTimeout = time.Second * 5

const PORT = "80"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	store, err := dbcontroller.NewStore()
	if err != nil {
		log.Printf("Failed to initilize ChatApp store: %+v\n", err)
		os.Exit(1)
	}
	defer store.Close()
	log.Println("Store initialization completed")

	store.AutoMigrate()
	log.Println("Auto migration completed")

	api := apiController{
		store: store,
	}

	r := mux.NewRouter()
	r.HandleFunc("/register", api.register).Methods(http.MethodPost)
	r.HandleFunc("/", api.helloWorldHandler)

	srv := &http.Server{
		Addr:         "0.0.0.0:" + PORT,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      r,
	}

	go func() {
		log.Println("Listen on port:" + PORT)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	<-c

	log.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), gracefullShutdownTimeout)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Bye")

	os.Exit(0)
}
