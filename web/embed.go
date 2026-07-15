//go:build embedweb

package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// DistFS возвращает собранный фронтенд. Второе значение — true, если
// SPA действительно встроена (сборка с тегом embedweb).
func DistFS() (fs.FS, bool) {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub, true
}
