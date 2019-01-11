package dbcontroller

import (
	"time"

	"../model"
	"github.com/jinzhu/gorm"
)

type ChatUserRepo struct {
	db          *gorm.DB
	idGenerator *IDGenerator
}

func (r *ChatUserRepo) Get(chatID, userID string, chatUser *model.ChatUser) error {
	return r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(chatUser).Error
}

func (r *ChatUserRepo) ListByChatID(chatID string, chatUsers *[]model.ChatUser) error {
	return r.db.Where("chat_id = ?", chatID).Find(&chatUsers).Error
}
//
// func (r *ChatUserRepo) ListByUserID(userID string, chatUsers *[]model.ChatUser) error {
// 	return r.db.Where("user_id = ?", userID).Find(&chatUsers).Error
// }

func (r *ChatUserRepo) Create(chatUser *model.ChatUser) error {

	now := time.Now()
	chatUser.CreatedAt = &now
	chatUser.UpdatedAt = &now

	return r.db.Create(chatUser).Error
}

func (r *ChatUserRepo) Delete(chatID, userID string) error {
	return r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(model.ChatUser{}).Error
}

func (r *ChatUserRepo) DeleteByChatID(chatID string) error {
	return r.db.Where("chat_id = ?", chatID).Delete(model.ChatUser{}).Error
}

func (r *ChatUserRepo) DeleteByUserID(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(model.ChatUser{}).Error
}

func (r *ChatUserRepo) Exists(chatID, userID string) (bool, error) {
	var count int64

	err := r.db.Model(&model.ChatUser{}).Where("chat_id = ? AND user_id = ?", chatID, userID).Count(&count).Error
	if err != nil {
		return true, err
	}

	exists := count > 0

	return exists, nil
}
