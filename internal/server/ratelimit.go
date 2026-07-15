package server

import (
	"sync"
	"time"
)

// rateLimiter — скользящее окно неудачных попыток на ключ (ip|username).
type rateLimiter struct {
	mu     sync.Mutex
	fails  map[string][]time.Time
	limit  int
	window time.Duration
	now    func() time.Time
}

func newRateLimiter(limit int, window time.Duration, now func() time.Time) *rateLimiter {
	return &rateLimiter{fails: map[string][]time.Time{}, limit: limit, window: window, now: now}
}

// recent возвращает попытки внутри окна, попутно вычищая устаревшие.
// Вызывать под mu.
func (l *rateLimiter) recent(key string) []time.Time {
	cutoff := l.now().Add(-l.window)
	kept := l.fails[key][:0]
	for _, ts := range l.fails[key] {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if len(kept) == 0 {
		delete(l.fails, key)
		return nil
	}
	l.fails[key] = kept
	return kept
}

// tooMany — достигнут ли лимит неудач в окне для ключа.
func (l *rateLimiter) tooMany(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.recent(key)) >= l.limit
}

// fail фиксирует неудачную попытку.
func (l *rateLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fails[key] = append(l.recent(key), l.now())
}

// reset сбрасывает счётчик (после успешного входа).
func (l *rateLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}
