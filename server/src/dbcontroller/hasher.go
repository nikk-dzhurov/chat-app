package dbcontroller

import (
	"golang.org/x/crypto/bcrypt"
)

const defaultHashCost = 16

type Hasher struct {
	hashCost int
}

func NewHasher(hashCost int) *Hasher {
	if (hashCost > bcrypt.MaxCost || hashCost < bcrypt.MinCost) {
		hashCost = defaultHashCost
	}

	return &Hasher{
		hashCost: hashCost,
	}
}

func (h Hasher) HashString(s string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(s), h.hashCost)
}
