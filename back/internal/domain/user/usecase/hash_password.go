package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword gera o hash bcrypt da senha com pepper.
func HashPassword(password string, secret []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pepper(password, secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// pepper aplica HMAC-SHA256 com PASSWORD_SECRET antes do bcrypt. Isso adiciona
// o "pepper" e mantém a entrada abaixo do limite de 72 bytes do bcrypt.
func pepper(password string, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(password))
	return []byte(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}
