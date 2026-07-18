//go:build !e2eseed

package main

import (
	"database/sql"

	"github.com/OlegKopeykin/fittrack/internal/telegram"
)

// Релизная сборка: e2e-сид отсутствует. Ни owner-аккаунта с фиксированным
// паролем, ни соответствующего кода в бинаре нет. Полная версия — в seed.go
// под тегом сборки e2eseed.
func maybeSeedE2E(*sql.DB) {}

// Релизная сборка использует реальный Telegram-клиент (nil → server.New).
func e2eTelegram() telegram.Client { return nil }
