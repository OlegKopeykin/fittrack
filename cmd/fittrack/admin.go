package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/auth"
	"github.com/OlegKopeykin/fittrack/internal/backup"
	"github.com/OlegKopeykin/fittrack/internal/config"
	"github.com/OlegKopeykin/fittrack/internal/db"
	"github.com/OlegKopeykin/fittrack/internal/db/gen"
	"github.com/OlegKopeykin/fittrack/internal/telegram"
)

// runAdmin — сервисные подкоманды на боксе:
//
//	fittrack admin create-invite [--owner] [--days N]
//	fittrack admin reset-password <username>   (новый пароль — со stdin)
func runAdmin(cfg config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("использование: fittrack admin <create-invite|reset-password>")
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	q := gen.New(conn)
	ctx := context.Background()

	switch args[0] {
	case "token":
		fs := flag.NewFlagSet("token", flag.ContinueOnError)
		user := fs.String("user", "", "имя пользователя")
		name := fs.String("name", "cli", "имя токена")
		days := fs.Int("days", 0, "срок действия в днях (0 — бессрочно)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *user == "" {
			return fmt.Errorf("укажите --user")
		}
		u, err := q.GetUserByUsername(ctx, *user)
		if err != nil {
			return fmt.Errorf("пользователь %q не найден", *user)
		}
		plaintext, hash, prefix, err := auth.NewAPIToken()
		if err != nil {
			return err
		}
		var expires sql.NullString
		if *days > 0 {
			expires = sql.NullString{
				String: time.Now().UTC().AddDate(0, 0, *days).Format(time.RFC3339),
				Valid:  true,
			}
		}
		if _, err := q.CreateApiToken(ctx, gen.CreateApiTokenParams{
			UserID: u.ID, Name: *name, TokenHash: hash, Prefix: prefix,
			CreatedAt: time.Now().UTC().Format(time.RFC3339), ExpiresAt: expires,
		}); err != nil {
			return err
		}
		fmt.Println(plaintext)
		return nil

	case "create-invite":
		fs := flag.NewFlagSet("create-invite", flag.ContinueOnError)
		owner := fs.Bool("owner", false, "инвайт с ролью owner")
		days := fs.Int("days", 14, "срок действия в днях (0 — бессрочно)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		code, err := auth.NewInviteCode()
		if err != nil {
			return err
		}
		role := "user"
		if *owner {
			role = "owner"
		}
		var expires sql.NullString
		if *days > 0 {
			expires = sql.NullString{
				String: time.Now().UTC().AddDate(0, 0, *days).Format(time.RFC3339),
				Valid:  true,
			}
		}
		inv, err := q.CreateInvite(ctx, gen.CreateInviteParams{
			Code:      code,
			Role:      role,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			ExpiresAt: expires,
		})
		if err != nil {
			return err
		}
		fmt.Printf("инвайт (%s): %s\n", inv.Role, inv.Code)
		return nil

	case "reset-password":
		if len(args) < 2 {
			return fmt.Errorf("использование: fittrack admin reset-password <username>")
		}
		username := args[1]
		fmt.Fprint(os.Stderr, "новый пароль: ")
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		password = strings.TrimRight(password, "\r\n")
		if len([]rune(password)) < 8 {
			return fmt.Errorf("пароль короче 8 символов")
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		rows, err := q.UpdateUserPassword(ctx, gen.UpdateUserPasswordParams{
			PasswordHash: hash,
			Username:     username,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("пользователь %q не найден", username)
		}
		fmt.Println("пароль обновлён")
		return nil

	case "export-telegram":
		return runExportTelegram(ctx, q)

	default:
		return fmt.Errorf("неизвестная подкоманда %q", args[0])
	}
}

// runExportTelegram обходит пользователей с включённым экспортом и шлёт бэкап
// тем, у кого подошёл срок по частоте. Запускается ночным таймером.
func runExportTelegram(ctx context.Context, q *gen.Queries) error {
	settings, err := q.ListEnabledTelegram(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	tg := telegram.New()
	sent, skipped := 0, 0
	for _, st := range settings {
		if !backup.IsDue(st.Frequency, st.LastSentAt, now) {
			skipped++
			continue
		}
		user, err := q.GetUserByID(ctx, st.UserID)
		if err != nil {
			slog.Error("export-telegram: пользователь", "user_id", st.UserID, "err", err)
			continue
		}
		nowStr := now.Format(time.RFC3339)
		exp, err := backup.Build(ctx, q, st.UserID, user.Username, appVersion, nowStr)
		if err != nil {
			slog.Error("export-telegram: сборка", "user", user.Username, "err", err)
			continue
		}
		data, err := backup.Marshal(exp)
		if err != nil {
			slog.Error("export-telegram: json", "user", user.Username, "err", err)
			continue
		}
		filename := "fittrack-" + user.Username + "-" + nowStr[:10] + ".json"
		if err := tg.SendDocument(ctx, st.BotToken, st.ChatID, filename, data, ""); err != nil {
			slog.Error("export-telegram: отправка", "user", user.Username, "err", err)
			continue
		}
		if err := q.TouchTelegramSent(ctx, gen.TouchTelegramSentParams{LastSentAt: nowStr, UserID: st.UserID}); err != nil {
			slog.Error("export-telegram: touch", "user", user.Username, "err", err)
		}
		sent++
	}
	slog.Info("export-telegram завершён", "sent", sent, "skipped", skipped)
	return nil
}
