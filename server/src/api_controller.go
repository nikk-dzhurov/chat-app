package main

import (
	"net/http"
)

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Hello, World!</h1>"))
}
