package dbcontroller

import "github.com/jinzhu/gorm"

type Repo interface {
	Exists(string) (bool, error)
}

type BaseRepo struct {
	db          *gorm.DB
	idGenerator *IDGenerator
}

func (r *BaseRepo) Get(id string, result interface{}) error {
	err := r.db.Where("id = ?", id).First(result).Error
	if err != nil {
		return err
	}

	return nil
}

func (br *BaseRepo) GetValidID(r Repo) (string, error) {
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
