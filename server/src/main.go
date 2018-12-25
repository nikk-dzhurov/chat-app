package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
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

	origins := []string{"http://localhost:3001"}
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	r := mux.NewRouter()
	r.HandleFunc("/login", api.login).Methods(http.MethodPost)
	r.HandleFunc("/register", api.register).Methods(http.MethodPost)
	r.HandleFunc("/logout", api.logout).Methods(http.MethodPost)
	r.HandleFunc("/users", api.listUsers).Methods(http.MethodGet)
	r.HandleFunc("/user/{userID}/avatar", api.getAvatar).Methods(http.MethodGet)
	r.HandleFunc("/user/{userID}/avatar", api.uploadAvatar).Methods(http.MethodPost)

	r.HandleFunc("/chat", api.createChat).Methods(http.MethodPost)
	r.HandleFunc("/chats", api.listChats).Methods(http.MethodGet)
	r.HandleFunc("/chat/{chatID}", api.getChat).Methods(http.MethodGet)
	r.HandleFunc("/chat/{chatID}", api.updateChat).Methods(http.MethodPut)
	r.HandleFunc("/chat/{chatID}", api.deleteChat).Methods(http.MethodDelete)

	r.HandleFunc("/chat/{chatID}/message", api.createMessage).Methods(http.MethodPost)
	r.HandleFunc("/chat/{chatID}/messages", api.listMessages).Methods(http.MethodGet)
	r.HandleFunc("/chat/{chatID}/message/{messageID}", api.getMessage).Methods(http.MethodGet)
	r.HandleFunc("/chat/{chatID}/message/{messageID}", api.updateMessage).Methods(http.MethodPut)
	r.HandleFunc("/chat/{chatID}/message/{messageID}", api.deleteMessage).Methods(http.MethodDelete)

	corsRouter := handlers.CORS(handlers.AllowedOrigins(origins), handlers.AllowedMethods(methods), handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
		handlers.ExposedHeaders([]string{"Authorization", "Content-Type"}))(r)

	loggedRouter := handlers.LoggingHandler(os.Stdout, corsRouter)
	srv := &http.Server{
		Addr:         "0.0.0.0:" + PORT,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      loggedRouter,
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
