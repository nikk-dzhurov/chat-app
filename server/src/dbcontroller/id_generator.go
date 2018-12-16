package dbcontroller

import (
	"math/rand"
)

const defaultIDLen = 16
const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type IDGenerator struct {
	idLen      int
	charset    string
	charsetLen int
}

func NewIDGenerator(idLen int) *IDGenerator {
	if (idLen < 8 || idLen > 32) {
		idLen = defaultIDLen
	}

	return &IDGenerator{
		idLen:      defaultIDLen,
		charset:    charset,
		charsetLen: len(charset),
	}
}

func (g *IDGenerator) generate() string {
	id := ""
	var charIdx int
	var randChar byte
	for i := 0; i < g.idLen; i++ {
		charIdx = rand.Intn(g.charsetLen)
		randChar = g.charset[charIdx]
		id += string(randChar)
	}

	return id
}

func (g *IDGenerator) generateN(n int) string {
	if n < 1 {
		n = g.idLen
	}

	id := ""
	var charIdx int
	var randChar byte
	for i := 0; i < n; i++ {
		charIdx = rand.Intn(g.charsetLen)
		randChar = g.charset[charIdx]
		id += string(randChar)
	}

	return id
}
