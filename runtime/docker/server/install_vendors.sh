#!/bin/bash

if [ ! -d "/go/src" ]; then
	echo "Install vendors"
	go get -u github.com/go-sql-driver/mysql
	go get -u github.com/gorilla/handlers
	go get -u github.com/gorilla/mux
	go get -u github.com/gorilla/websocket
	go get -u github.com/jinzhu/gorm
	go get -u golang.org/x/crypto/...
fi
