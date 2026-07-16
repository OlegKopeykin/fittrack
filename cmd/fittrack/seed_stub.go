//go:build !e2eseed

package main

import "database/sql"

// Релизная сборка: e2e-сид отсутствует. Ни owner-аккаунта с фиксированным
// паролем, ни соответствующего кода в бинаре нет. Полная версия — в seed.go
// под тегом сборки e2eseed.
func maybeSeedE2E(*sql.DB) {}
