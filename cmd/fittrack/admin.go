package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/auth"
	"github.com/OlegKopeykin/fittrack/internal/config"
	"github.com/OlegKopeykin/fittrack/internal/db"
	"github.com/OlegKopeykin/fittrack/internal/db/gen"
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

	default:
		return fmt.Errorf("неизвестная подкоманда %q", args[0])
	}
}
