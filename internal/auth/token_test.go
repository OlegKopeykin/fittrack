package auth_test

import (
	"strings"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/auth"
)

func TestNewAPIToken(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		plain, hash, prefix, err := auth.NewAPIToken()
		if err != nil {
			t.Fatalf("NewAPIToken: %v", err)
		}
		if !strings.HasPrefix(plain, "fit_") {
			t.Errorf("токен %q без префикса fit_", plain)
		}
		if !strings.HasPrefix(plain, prefix) || len(prefix) != len("fit_")+6 {
			t.Errorf("префикс %q не согласуется с токеном %q", prefix, plain)
		}
		if hash != auth.HashToken(plain) {
			t.Errorf("хэш не совпадает с HashToken(plain)")
		}
		if hash == plain {
			t.Error("хэш равен открытому токену")
		}
		if seen[plain] {
			t.Fatalf("токен %q повторился", plain)
		}
		seen[plain] = true
	}
}

func TestHashTokenStable(t *testing.T) {
	if auth.HashToken("fit_abc") != auth.HashToken("fit_abc") {
		t.Error("HashToken недетерминирован")
	}
	if auth.HashToken("fit_abc") == auth.HashToken("fit_abd") {
		t.Error("разные токены дали одинаковый хэш")
	}
}
