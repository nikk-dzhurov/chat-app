package dbcontroller

import (
	"time"

	"../model"
)

type ChatRepo struct {
	BaseEntityRepo
}

func (r *ChatRepo) Create(chat *model.Chat) error {

	now := time.Now()
	chat.CreatedAt = &now

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

	oldChat.Title = chat.Title

	chat = &oldChat
	err = r.db.Save(chat).Error
	if err != nil {
		return err
	}

	return nil
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
