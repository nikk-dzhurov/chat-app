package main

import (
	"time"
)

type Chat struct {
	ID string
	CreatorID string
	Title string
	CreatedAt *time.Time
}

type Message struct {
	ID string
	UserID string
	ChatID string
	Message string
	CreatedAt *time.Time
}

type ChatUser struct {
	ChatID string
	UserID string
	CreatedAt *time.Time
}

type User struct {
	ID string
	Username string
	Password string
	Name string
	CreatedAt *time.Time
}
