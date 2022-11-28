package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func GenerateToken() string {
	//generate a random 32 bytes token
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		log.Fatal(err)
	}

	//convert to base64
	return base64.StdEncoding.EncodeToString(token)
}