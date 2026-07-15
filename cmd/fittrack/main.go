package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/config"
	"github.com/OlegKopeykin/fittrack/internal/db"
	"github.com/OlegKopeykin/fittrack/internal/seed"
	"github.com/OlegKopeykin/fittrack/internal/server"
	"github.com/OlegKopeykin/fittrack/web"
)

func main() {
	cfg := config.FromEnv()

	if len(os.Args) > 1 && os.Args[1] == "admin" {
		if err := runAdmin(cfg, os.Args[2:]); err != nil {
			slog.Error("admin", "err", err)
			os.Exit(1)
		}
		return
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("база данных", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := seed.LoadCatalog(context.Background(), conn); err != nil {
		slog.Error("сид каталога", "err", err)
		os.Exit(1)
	}

	if os.Getenv("FITTRACK_E2E_SEED") == "1" {
		if err := seedE2E(conn); err != nil {
			slog.Error("e2e seed", "err", err)
			os.Exit(1)
		}
	}

	static, embedded := web.DistFS()
	if !embedded {
		slog.Warn("сборка без тега embedweb — вместо SPA отдаётся заглушка")
	}

	srv := &http.Server{
		Addr: cfg.Addr,
		Handler: server.New(server.Options{
			DB:            conn,
			Static:        static,
			PublicOrigin:  cfg.PublicOrigin,
			SecureCookies: cfg.SecureCookies,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("fittrack запускается", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("http-сервер завершился с ошибкой", "err", err)
		os.Exit(1)
	}
}
