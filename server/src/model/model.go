package model

import (
	"time"
)

type Chat struct {
	ID        string     `json:"id" db:"id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin; primary_key; not null;"`
	CreatorID string     `json:"creatorId" db:"creator_id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin; index; not null;"`
	Title     string     `json:"title" db:"title" sql:"type:varchar(256)"`
	CreatedAt *time.Time `json:"createdAt" db:"created_at" sql:"type:datetime(3)"`
}

func (c Chat) TableName() string {
	return "chat"
}

type Message struct {
	ID        string     `json:"id" db:"id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin; primary_key; not null;"`
	UserID    string     `json:"userId" db:"user_id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin; index; not null;"`
	ChatID    string     `json:"chatId" db:"chat_id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin;index; not null;"`
	Message   string     `json:"message" db:"message" sql:"type:longtext CHARSET utf8mb4 COLLATE utf8mb4_general_ci"`
	CreatedAt *time.Time `json:"createdAt" db:"created_at" sql:"type:datetime(3)"`
}

func (m Message) TableName() string {
	return "message"
}

type ChatUser struct {
	ChatID    string     `json:"chatId" db:"chat_id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin; primary_key; not null;"`
	UserID    string     `json:"userId" db:"user_id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin; primary_key; not null;"`
	CreatedAt *time.Time `json:"createdAt" db:"created_at" sql:"type:datetime(3)"`
}

func (cu ChatUser) TableName() string {
	return "chat_user"
}

type User struct {
	ID           string     `json:"id" db:"id" sql:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin; primary_key; not null;"`
	Username     string     `json:"username" db:"username" sql:"type:varchar(256) CHARACTER SET ascii COLLATE ascii_bin; index; not null;"`
	Password     string     `json:"password" sql:"-"`
	PasswordHash string     `json:"-" db:"password_hash" sql:"type:varchar(256) CHARACTER SET ascii COLLATE ascii_bin; not null;"`
	FullName     string     `json:"fullName" db:"full_name" sql:"type:varchar(256) CHARSET utf8mb4 COLLATE utf8mb4_general_ci"`
	CreatedAt    *time.Time `json:"createdAt" db:"created_at" sql:"type:datetime(3)"`
}

func (u User) TableName() string {
	return "user"
}
