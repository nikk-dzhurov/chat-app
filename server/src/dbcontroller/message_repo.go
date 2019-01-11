package dbcontroller

import (
	"time"

	"../model"
)

type MessageRepo struct {
	BaseEntityRepo
}

func (r *MessageRepo) ListByChatID(chatID string, messages *[]model.Message) error {
	return r.db.Where("chat_id = ?", chatID).Find(&messages).Error
}

func (r *MessageRepo) Create(message *model.Message) error {

	now := time.Now()
	message.CreatedAt = &now
	message.UpdatedAt = &now

	var err error
	message.ID, err = r.GetValidID(r)
	if err != nil {
		return err
	}

	err = r.db.Create(message).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MessageRepo) Update(message *model.Message) error {

	oldMsg := model.Message{}
	err := r.Get(message.ID, &oldMsg)
	if err != nil {
		return err
	}

	now := time.Now()
	oldMsg.UpdatedAt = &now
	oldMsg.Message = message.Message

	*message = oldMsg
	err = r.db.Save(message).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *MessageRepo) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(model.Message{}).Error
}

func (r *MessageRepo) DeleteByChatID(chatID string) error {
	return r.db.Where("chat_id = ?", chatID).Delete(model.Message{}).Error
}

func (r *MessageRepo) Exists(id string) (bool, error) {
	var count int64

	err := r.db.Model(&model.Message{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return true, err
	}

	exists := count > 0

	return exists, nil
}
