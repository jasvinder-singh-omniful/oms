package hash

import (
	"fmt"

	"github.com/omniful/go_commons/log"
	"golang.org/x/crypto/bcrypt"
)


func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error when hashing password %s", err.Error())
	}

	return string(hash), nil
}

func VerifyPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Errorf("error when verifying hashed password %s", err.Error())
		return false
	}

	return true
}

