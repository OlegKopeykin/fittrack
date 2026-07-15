package config_test

import (
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/config"
)

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		wantAddr string
	}{
		{
			name:     "дефолты без окружения",
			env:      map[string]string{"FITTRACK_ADDR": "", "FITTRACK_DB": "", "FITTRACK_PUBLIC_ORIGIN": ""},
			wantAddr: "127.0.0.1:8080",
		},
		{
			name:     "адрес из FITTRACK_ADDR",
			env:      map[string]string{"FITTRACK_ADDR": "127.0.0.1:9090"},
			wantAddr: "127.0.0.1:9090",
		},
		{
			// Пин контракта envOr: пустая переменная = «не задано» → дефолт.
			// Иначе рефакторинг на os.LookupEnv молча превратит Addr в ""
			// (ListenAndServe(":http") на всех интерфейсах).
			name:     "пустая FITTRACK_ADDR — дефолт",
			env:      map[string]string{"FITTRACK_ADDR": ""},
			wantAddr: "127.0.0.1:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			cfg := config.FromEnv()
			if cfg.Addr != tt.wantAddr {
				t.Errorf("Addr = %q, want %q", cfg.Addr, tt.wantAddr)
			}
		})
	}
}

func TestFromEnvDBAndOrigin(t *testing.T) {
	t.Run("дефолтный путь БД", func(t *testing.T) {
		t.Setenv("FITTRACK_DB", "")
		if got := config.FromEnv().DBPath; got != "fittrack.db" {
			t.Errorf("DBPath = %q, want fittrack.db", got)
		}
	})
	t.Run("путь БД из окружения", func(t *testing.T) {
		t.Setenv("FITTRACK_DB", "/var/lib/fittrack/fittrack.db")
		if got := config.FromEnv().DBPath; got != "/var/lib/fittrack/fittrack.db" {
			t.Errorf("DBPath = %q", got)
		}
	})
	t.Run("https-origin включает Secure-куки", func(t *testing.T) {
		t.Setenv("FITTRACK_PUBLIC_ORIGIN", "https://fit.example.com")
		cfg := config.FromEnv()
		if cfg.PublicOrigin != "https://fit.example.com" || !cfg.SecureCookies {
			t.Errorf("cfg = %+v, want PublicOrigin=https://… и SecureCookies=true", cfg)
		}
	})
	t.Run("без origin куки не Secure", func(t *testing.T) {
		t.Setenv("FITTRACK_PUBLIC_ORIGIN", "")
		if config.FromEnv().SecureCookies {
			t.Error("SecureCookies = true без PublicOrigin, want false")
		}
	})
}
