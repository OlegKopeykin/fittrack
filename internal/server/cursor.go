package server

import (
	"encoding/base64"
	"strconv"
	"strings"
)

// encodeCursor/decodeCursor — keyset-курсор пагинации истории тренировок:
// пара (date, id) в base64.
func encodeCursor(date string, id int64) string {
	return base64.RawURLEncoding.EncodeToString([]byte(date + "|" + strconv.FormatInt(id, 10)))
}

func decodeCursor(c string) (string, int64, bool) {
	b, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return "", 0, false
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, false
	}
	return parts[0], id, true
}
