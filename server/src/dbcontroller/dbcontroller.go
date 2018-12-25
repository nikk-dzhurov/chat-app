package dbcontroller

import (
	// used by gorm
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"../model"
)

type Store struct {
	db          *gorm.DB
	idGenerator *IDGenerator
	hasher      *Hasher

	UserRepo     *UserRepo
	ChatRepo     *ChatRepo
	TokenRepo    *TokenRepo
	MessageRepo  *MessageRepo
	ChatUserRepo *ChatUserRepo
}

func NewStore() (*Store, error) {
	var err error
	db, err := gorm.Open("mysql", "test:test@tcp(db:3306)/chatapp?parseTime=true")
	if err != nil {
		return nil, err
	}

	db.LogMode(true)
	db.SingularTable(true)
	db.DB().SetMaxOpenConns(10)
	db.Callback().Create().Remove("gorm:update_time_stamp")

	idGenerator := NewIDGenerator(-1)
	hasher := NewHasher(-1)

	baseRepo := BaseEntityRepo{
		db:          db,
		idGenerator: idGenerator,
	}

	return &Store{
		db: db,
		UserRepo: &UserRepo{
			BaseEntityRepo: baseRepo,
			hasher:         hasher,
		},
		ChatRepo: &ChatRepo{
			BaseEntityRepo: baseRepo,
		},
		MessageRepo: &MessageRepo{
			BaseEntityRepo: baseRepo,
		},
		ChatUserRepo: &ChatUserRepo{
			db:          db,
			idGenerator: idGenerator,
		},
		TokenRepo: &TokenRepo{
			db:          db,
			idGenerator: idGenerator,
		},
	}, nil
}

func (store *Store) AutoMigrate() {
	models := []interface{}{
		&model.Chat{},
		&model.Message{},
		&model.ChatUser{},
		&model.User{},
		&model.AccessToken{},
		&model.UserAvatar{},
	}

	store.db.AutoMigrate(models...)
}

func (store *Store) Close() {
	store.db.Close()
}
