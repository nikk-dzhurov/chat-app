package dbcontroller

import (
	"time"

	"../model"
)

type MessageRepo struct {
	BaseEntityRepo
}

func (r *MessageRepo) Create(message *model.Message) error {

	now := time.Now()
	message.CreatedAt = &now

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

	oldMsg.Message = message.Message

	message = &oldMsg
	err = r.db.Save(message).Error
	if err != nil {
		return err
	}

	return nil
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
