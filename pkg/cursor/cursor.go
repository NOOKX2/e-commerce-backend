package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

func Encode(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func Decode(s string, v any) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("empty cursor")
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}
