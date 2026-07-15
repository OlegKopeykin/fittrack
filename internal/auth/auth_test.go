package auth_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/auth"
)

func TestHashPasswordFormat(t *testing.T) {
	hash, err := auth.HashPassword("секретный пароль")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Errorf("hash = %q, want PHC-строку argon2id с параметрами OWASP", hash)
	}
}

func TestHashPasswordSaltsDiffer(t *testing.T) {
	h1, err1 := auth.HashPassword("пароль")
	h2, err2 := auth.HashPassword("пароль")
	if err1 != nil || err2 != nil {
		t.Fatalf("HashPassword: %v / %v", err1, err2)
	}
	if h1 == h2 {
		t.Error("одинаковые хэши для двух вызовов — соль не случайная")
	}
}

func TestVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("правильный")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	tests := []struct {
		name     string
		password string
		encoded  string
		want     bool
	}{
		{"верный пароль", "правильный", hash, true},
		{"неверный пароль", "неправильный", hash, false},
		{"пустой пароль", "", hash, false},
		{"мусор вместо хэша", "правильный", "не-hash", false},
		{"пустой хэш", "правильный", "", false},
		{"подделанные параметры", "правильный", strings.Replace(hash, "t=2", "t=1", 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := auth.VerifyPassword(tt.password, tt.encoded); got != tt.want {
				t.Errorf("VerifyPassword = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewInviteCode(t *testing.T) {
	seen := map[string]bool{}
	valid := regexp.MustCompile(`^[A-Z2-7]{12}$`)
	for i := 0; i < 50; i++ {
		code, err := auth.NewInviteCode()
		if err != nil {
			t.Fatalf("NewInviteCode: %v", err)
		}
		if !valid.MatchString(code) {
			t.Fatalf("code = %q, want 12 символов base32 (A-Z, 2-7)", code)
		}
		if seen[code] {
			t.Fatalf("код %q повторился — генерация не случайная", code)
		}
		seen[code] = true
	}
}
