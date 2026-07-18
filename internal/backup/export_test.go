package backup

import (
	"testing"
	"time"
)

func TestIsDue(t *testing.T) {
	now := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	ago := func(d time.Duration) string { return now.Add(-d).Format(time.RFC3339) }

	cases := []struct {
		name string
		freq string
		last string
		want bool
	}{
		{"никогда не слали", "daily", "", true},
		{"ежедневно — прошли сутки", "daily", ago(25 * time.Hour), true},
		{"ежедневно — час назад", "daily", ago(1 * time.Hour), false},
		{"еженедельно — 8 дней", "weekly", ago(8 * 24 * time.Hour), true},
		{"еженедельно — 3 дня", "weekly", ago(3 * 24 * time.Hour), false},
		{"ежемесячно — 31 день", "monthly", ago(31 * 24 * time.Hour), true},
		{"ежемесячно — 10 дней", "monthly", ago(10 * 24 * time.Hour), false},
		{"битая метка — слать", "daily", "не-дата", true},
	}
	for _, c := range cases {
		if got := IsDue(c.freq, c.last, now); got != c.want {
			t.Errorf("%s: IsDue = %v, want %v", c.name, got, c.want)
		}
	}
}
