package utils

import (
	"crypto/rand"
	"fmt"
	"os"
)

func GenerateToken() string {
	token := make([]byte, 32)
    if _, err := rand.Read(token); err!= nil {
        fmt.Println(err)
        os.Exit(1)
    }
	return fmt.Sprintf("%x", token)
}