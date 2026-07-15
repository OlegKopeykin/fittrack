package exercises_test

import (
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/exercises"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Жим гантелей лёжа", "жим гантелей лежа"},
		{"  Присед   в   Смите  ", "присед в смите"},
		{"Face Pull", "face pull"},
		{"РДЛ", "рдл"},
		{"Тяга\tнижнего\nблока", "тяга нижнего блока"},
		{"ЁЖ и ёж", "еж и еж"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := exercises.Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
