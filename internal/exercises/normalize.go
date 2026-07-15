// Package exercises — доменная логика каталога упражнений.
package exercises

import (
	"strings"
	"unicode"
)

// Normalize приводит имя/алиас к канонической форме для поиска и дедупа:
// нижний регистр, ё→е, схлопнутые пробелы, обрезка по краям.
func Normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true // чтобы срезать ведущие пробелы
	for _, r := range strings.ToLower(s) {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		if r == 'ё' {
			r = 'е'
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return strings.TrimRight(b.String(), " ")
}
