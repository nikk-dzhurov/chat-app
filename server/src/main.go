package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

const gracefullShutdownTimeout = time.Second*5
const idleTimeout = time.Second*60
const readTimeout = time.Second*5
const writeTimeout = time.Second*5

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", helloWorldHandler)

	srv := &http.Server{
		Addr: "0.0.0.0:80",
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	<-c

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), gracefullShutdownTimeout)
	defer cancel()

	srv.Shutdown(ctx)

	log.Println("bye")

	os.Exit(0)
}
