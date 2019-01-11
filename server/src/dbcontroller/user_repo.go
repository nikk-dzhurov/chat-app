package dbcontroller

import (
	"time"

	"../model"
)

type UserRepo struct {
	BaseEntityRepo
	hasher *Hasher
}

func (r *UserRepo) List(users *[]model.User) error {
	return r.db.Find(users).Error
}

func (r *UserRepo) Create(user *model.User) error {

	now := time.Now()
	user.CreatedAt = &now
	user.FullName = ""

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

func (r *UserRepo) UpdateUpdatedAt(userID string, date *time.Time) error {

	if date == nil {
		now := time.Now()
		date = &now
	}

	user := model.User{}
	err := r.Get(userID, &user)
	if err != nil {
		return err
	}

	user.UpdatedAt = date

	err = r.db.Save(&user).Error
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

	*user = old
	err = r.db.Save(user).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(model.User{}).Error
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

func (r *UserRepo) HasAvatar(userID string) (bool, error) {
	var count int64

	err := r.db.Model(&model.UserAvatar{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return true, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (r *UserRepo) GetAvatar(userID string, avatar *model.UserAvatar) error {
	return r.db.Where("user_id = ?", userID).First(avatar).Error
}

func (r *UserRepo) CreateAvatar(avatar *model.UserAvatar) error {
	return r.db.Create(avatar).Error
}

func (r *UserRepo) UpdateAvatar(avatar *model.UserAvatar) error {
	return r.db.Save(avatar).Error
}
