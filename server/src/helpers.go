package main

import (
	"math/rand"
)

const idLen = 16
const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const charsetLen = len(charset)

func generateID() string {
	id := ""
	var randChar byte
	for i := 0; i < idLen; i++ {
		randChar = charset[rand.Intn(charsetLen)]
		id += string(randChar)
	}

	return id
}
