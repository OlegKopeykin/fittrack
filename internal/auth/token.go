package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"strings"
)

// TokenPrefix — префикс персональных API-токенов (машинный API).
const TokenPrefix = "fit_"

// NewAPIToken генерирует персональный токен: возвращает открытое значение
// (показывается пользователю один раз), его sha256-хэш (хранится в БД) и
// короткий префикс для отображения в списке.
func NewAPIToken() (plaintext, hash, prefix string, err error) {
	b := make([]byte, 20)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	body := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
	plaintext = TokenPrefix + body
	hash = HashToken(plaintext)
	prefix = plaintext[:len(TokenPrefix)+6]
	return plaintext, hash, prefix, nil
}

// HashToken возвращает hex-sha256 токена (для хранения и сравнения).
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
