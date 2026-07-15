//go:build !embedweb

package web

import (
	"embed"
	"io/fs"
)

//go:embed placeholder
var placeholder embed.FS

// DistFS в dev-сборке (без тега embedweb) отдаёт страницу-заглушку;
// фронтенд в dev-режиме обслуживает vite dev server.
func DistFS() (fs.FS, bool) {
	sub, err := fs.Sub(placeholder, "placeholder")
	if err != nil {
		panic(err)
	}
	return sub, false
}
