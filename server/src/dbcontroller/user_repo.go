package dbcontroller

import (
	"time"

	"../model"
)

type UserRepo struct {
	BaseEntityRepo
	hasher *Hasher
}

func (r *UserRepo) Create(user *model.User) error {

	now := time.Now()
	user.CreatedAt = &now

	var err error
	user.ID, err = r.GetValidID(r)
	if err != nil {
		return err
	}

	hash, err := r.hasher.HashString(user.Password)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)

	err = r.db.Create(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) Update(user *model.User) error {

	old := model.User{}
	err := r.Get(user.ID, &old)
	if err != nil {
		return err
	}

	old.FullName = user.FullName

	user = &old
	err = r.db.Save(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	user := model.User{}
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) Exists(id string) (bool, error) {
	var count int64

	err := r.db.Model(&model.User{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return true, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (r *UserRepo) ExistsUsername(username string) (bool, error) {
	var count int64

	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return true, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (r *UserRepo) VerifyPassword(hash, password string) bool {
	if err := r.hasher.CompareHashAndPassword(hash, password); err != nil {
		return false
	}

	return true
}
