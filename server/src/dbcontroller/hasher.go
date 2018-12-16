package dbcontroller

import (
	"golang.org/x/crypto/bcrypt"
)

type Hasher struct {
	hashCost int
}

func NewHasher(hashCost int) *Hasher {
	if (hashCost > bcrypt.MaxCost || hashCost < bcrypt.MinCost) {
		hashCost = bcrypt.DefaultCost
	}

	return &Hasher{
		hashCost: hashCost,
	}
}

func (h *Hasher) HashString(s string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(s), h.hashCost)
}

func (h *Hasher) CompareHashAndPassword(hash, pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}
