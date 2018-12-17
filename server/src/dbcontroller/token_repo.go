package dbcontroller

import (
	"time"

	"../model"
	"github.com/jinzhu/gorm"
)

type TokenRepo struct {
	db          *gorm.DB
	idGenerator *IDGenerator
}

func (r *TokenRepo) Get(token string) (*model.AccessToken, error) {
	accessToken := model.AccessToken{}
	err := r.db.Where("token = ?", token).First(&accessToken).Error
	if err != nil {
		return nil, err
	}

	return &accessToken, nil
}

func (r *TokenRepo) GetByUserID(userID string) (*model.AccessToken, error) {
	accessToken := model.AccessToken{}
	err := r.db.Where("user_id = ?", userID).First(&accessToken).Error
	if err != nil {
		return nil, err
	}

	return &accessToken, nil
}

func (r *TokenRepo) Create(accessToken *model.AccessToken) error {
	token := r.idGenerator.generateN(64)

	expiresAt := time.Now().Add(time.Minute * 30)

	accessToken.Token = token
	accessToken.ExpiresAt = &expiresAt

	err := r.db.Create(accessToken).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *TokenRepo) Exists(userID string) (bool, error) {
	var count int64

	err := r.db.Model(&model.AccessToken{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return true, err
	}

	exists := count > 0

	return exists, nil
}

func (r *TokenRepo) Delete(t *model.AccessToken) error {
	return r.db.Delete(t).Error
}
