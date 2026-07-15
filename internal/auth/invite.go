package auth

import "crypto/rand"

const inviteAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

// NewInviteCode генерирует криптослучайный инвайт-код:
// 12 символов base32 без паддинга (A-Z, 2-7), удобен для ручного ввода.
func NewInviteCode() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i, v := range b {
		// len(alphabet) == 32, 256%32 == 0 — модуль без смещения.
		b[i] = inviteAlphabet[int(v)%len(inviteAlphabet)]
	}
	return string(b), nil
}
