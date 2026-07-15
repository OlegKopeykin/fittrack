package config

import "os"

// Config — конфигурация процесса, читается из окружения (FITTRACK_*).
type Config struct {
	// Addr — адрес HTTP-сервера (FITTRACK_ADDR).
	Addr string
	// DBPath — путь к файлу SQLite (FITTRACK_DB).
	DBPath string
	// PublicOrigin — внешний origin приложения для CSRF-проверки
	// (FITTRACK_PUBLIC_ORIGIN, например https://fit.example.com).
	PublicOrigin string
	// SecureCookies — Secure-флаг сессионной куки; включается автоматически,
	// когда PublicOrigin задан по https.
	SecureCookies bool
}

// FromEnv собирает конфигурацию из переменных окружения с дефолтами.
func FromEnv() Config {
	origin := os.Getenv("FITTRACK_PUBLIC_ORIGIN")
	return Config{
		Addr:          envOr("FITTRACK_ADDR", "127.0.0.1:8080"),
		DBPath:        envOr("FITTRACK_DB", "fittrack.db"),
		PublicOrigin:  origin,
		SecureCookies: len(origin) >= 8 && origin[:8] == "https://",
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
