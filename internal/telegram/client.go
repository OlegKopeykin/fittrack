// Package telegram — тонкий клиент Bot API: проверка токена, поиск чата по
// последним апдейтам и отправка файла-документа. За интерфейсом Client, чтобы
// в тестах подменять фейком без сетевых вызовов.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

type Client interface {
	// GetMe проверяет токен и возвращает username бота.
	GetMe(ctx context.Context, token string) (string, error)
	// ResolveChatID берёт chat_id из последнего сообщения боту (после /start).
	ResolveChatID(ctx context.Context, token string) (string, error)
	// SendDocument шлёт файл в чат с необязательной подписью.
	SendDocument(ctx context.Context, token, chatID, filename string, data []byte, caption string) error
}

type HTTPClient struct {
	http *http.Client
	base string
}

func New() *HTTPClient {
	return &HTTPClient{http: &http.Client{Timeout: 30 * time.Second}, base: "https://api.telegram.org"}
}

func (c *HTTPClient) url(token, method string) string {
	return fmt.Sprintf("%s/bot%s/%s", c.base, token, method)
}

func (c *HTTPClient) GetMe(ctx context.Context, token string) (string, error) {
	var out struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := c.getJSON(ctx, c.url(token, "getMe"), &out); err != nil {
		return "", err
	}
	if !out.OK {
		return "", fmt.Errorf("telegram: %s", out.Description)
	}
	return out.Result.Username, nil
}

func (c *HTTPClient) ResolveChatID(ctx context.Context, token string) (string, error) {
	var out struct {
		OK     bool `json:"ok"`
		Result []struct {
			Message struct {
				Chat struct {
					ID int64 `json:"id"`
				} `json:"chat"`
			} `json:"message"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := c.getJSON(ctx, c.url(token, "getUpdates"), &out); err != nil {
		return "", err
	}
	if !out.OK {
		return "", fmt.Errorf("telegram: %s", out.Description)
	}
	// Берём последнее сообщение с непустым чатом.
	for i := len(out.Result) - 1; i >= 0; i-- {
		if id := out.Result[i].Message.Chat.ID; id != 0 {
			return strconv.FormatInt(id, 10), nil
		}
	}
	return "", fmt.Errorf("не найдено сообщений боту — откройте бота и нажмите Start")
}

func (c *HTTPClient) SendDocument(ctx context.Context, token, chatID, filename string, data []byte, caption string) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("chat_id", chatID)
	if caption != "" {
		_ = mw.WriteField("caption", caption)
	}
	fw, err := mw.CreateFormFile("document", filename)
	if err != nil {
		return err
	}
	if _, err := fw.Write(data); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(token, "sendDocument"), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &out)
	if !out.OK {
		return fmt.Errorf("telegram sendDocument: %s", out.Description)
	}
	return nil
}

func (c *HTTPClient) getJSON(ctx context.Context, url string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}
