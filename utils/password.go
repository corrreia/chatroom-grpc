package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func CheckPassword(password string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(password), []byte(hash)) == nil
}

func GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}