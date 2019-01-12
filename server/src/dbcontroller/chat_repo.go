package dbcontroller

import (
	"time"

	"../model"
)

type ChatRepo struct {
	BaseEntityRepo
}

func (r *ChatRepo) ListByUserID(userID string, chats *[]model.Chat) error {
	return r.db.Joins("left join chat_user on chat_user.chat_id = chat.id").Where("chat_user.user_id = ?", userID).Find(&chats).Error
}

func (r *ChatRepo) Create(chat *model.Chat) error {

	now := time.Now()
	chat.CreatedAt = &now
	chat.UpdatedAt = &now

	var err error
	chat.ID, err = r.GetValidID(r)
	if err != nil {
		return err
	}

	err = r.db.Create(chat).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepo) Update(chat *model.Chat) error {

	oldChat := model.Chat{}
	err := r.Get(chat.ID, &oldChat)
	if err != nil {
		return err
	}

	now := time.Now()
	oldChat.UpdatedAt = &now
	oldChat.Title = chat.Title

	*chat = oldChat
	err = r.db.Save(chat).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepo) UpdateUpdatedAt(chatID string, date *time.Time) error {

	if date == nil {
		now := time.Now()
		date = &now
	}

	chat := model.Chat{}
	err := r.db.Model(&chat).Where("id = ?", chatID).Update("updated_at", date).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepo) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(model.Chat{}).Error
}

func (r *ChatRepo) DirectChatExists(firstUserID, secondUserID string) (bool, error) {
	var count int64

	err := r.db.Model(&model.Chat{}).Where("(creator_id = ? AND direct_user_id = ?) OR (creator_id = ? AND direct_user_id = ?)", firstUserID, secondUserID, secondUserID, firstUserID).Count(&count).Error
	if err != nil {
		return true, err
	}

	exists := count > 0

	return exists, nil
}

func (r *ChatRepo) Exists(id string) (bool, error) {
	var count int64

	err := r.db.Model(&model.Chat{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return true, err
	}

	exists := count > 0

	return exists, nil
}
