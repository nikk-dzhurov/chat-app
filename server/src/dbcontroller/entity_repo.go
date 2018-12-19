package dbcontroller

import "github.com/jinzhu/gorm"

type EntityRepo interface {
	Exists(string) (bool, error)
}

type BaseEntityRepo struct {
	db          *gorm.DB
	idGenerator *IDGenerator
}

func (r *BaseEntityRepo) Get(id string, result interface{}) error {
	err := r.db.Where("id = ?", id).First(result).Error
	if err != nil {
		return err
	}

	return nil
}

func (br *BaseEntityRepo) GetValidID(r EntityRepo) (string, error) {
	id := ""
	exists := true
	var err error
	for exists {
		id = br.idGenerator.generate()

		exists, err = r.Exists(id)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}
