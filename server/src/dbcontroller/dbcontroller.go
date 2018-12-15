package dbcontroller

import (
	// used by gorm
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Store struct {
	db          *gorm.DB
	idGenerator *IDGenerator
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

	return &Store{
		db:          db,
		idGenerator: NewIDGenerator(),
	}, nil
}

func (store *Store) AutoMigrate() {
	models := []interface{}{
		&Chat{},
		&Message{},
		&ChatUser{},
		&User{},
	}

	store.db.AutoMigrate(models...)
}

func (store *Store) Close() {
	store.db.Close()
}
