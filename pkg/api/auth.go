// Package api: простая аутентификация по паролю из TODO_PASSWORD.
// Реализован мини-JWT (HS256): подпись HMAC от header.payload с ключом = пароль.
// Токен кладётся в cookie "token", срок — 8 часов. Middleware auth(...) проверяет токен.
package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

// b64 — кодек base64url без паддинга, как в классических JWT.
var b64 = base64.RawURLEncoding

// jwtHeader — заголовок токена (тип и алгоритм).
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// jwtPayload — полезная нагрузка токена.
// Sum — hex(sha256(password)) для привязки токена к текущему паролю.
// Exp — unix-время истечения (через 8 часов).
type jwtPayload struct {
	Sum string `json:"sum"`
	Exp int64  `json:"exp"`
}

// sha256Hex возвращает hex-строку от sha256(input).
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(h)*2)
	j := 0
	for _, b := range h {
		out[j] = hexdigits[b>>4]
		out[j+1] = hexdigits[b&0x0f]
		j += 2
	}
	return string(out)
}

// makeJWT формирует токен: base64(header).base64(payload).base64(HMACSHA256(signing, password)).
func makeJWT(pass string) (string, error) {
	h := jwtHeader{Alg: "HS256", Typ: "JWT"}
	p := jwtPayload{
		Sum: sha256Hex(pass),
		Exp: time.Now().Add(8 * time.Hour).Unix(),
	}
	hb, _ := json.Marshal(h)
	pb, _ := json.Marshal(p)
	hs := b64.EncodeToString(hb)
	ps := b64.EncodeToString(pb)
	signing := hs + "." + ps

	mac := hmac.New(sha256.New, []byte(pass))
	_, _ = mac.Write([]byte(signing))
	sig := mac.Sum(nil)
	ss := b64.EncodeToString(sig)

	return signing + "." + ss, nil
}

// validateJWT проверяет подпись, срок действия и соответствие паролю.
func validateJWT(token, pass string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	signing := parts[0] + "." + parts[1]

	// проверка подписи
	mac := hmac.New(sha256.New, []byte(pass))
	_, _ = mac.Write([]byte(signing))
	expect := mac.Sum(nil)

	got, err := b64.DecodeString(parts[2])
	if err != nil || !hmac.Equal(got, expect) {
		return false
	}

	// проверка payload
	pb, err := b64.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var p jwtPayload
	if err := json.Unmarshal(pb, &p); err != nil {
		return false
	}
	if time.Now().Unix() >= p.Exp {
		return false
	}
	if p.Sum != sha256Hex(pass) {
		return false
	}
	return true
}

// auth — middleware для защиты маршрутов.
// Если переменная окружения TODO_PASSWORD пуста, защита отключена.
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if pass == "" {
			next(w, r)
			return
		}
		c, err := r.Cookie("token")
		if err != nil || !validateJWT(c.Value, pass) {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next(w, r)
	})
}

// signinHandler — обработчик POST /api/signin.
// Принимает JSON {"password": "..."} и возвращает {"token": "..."} при успехе.
func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}
	var in struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, map[string]string{"error": "json parse error"})
		return
	}
	pass := os.Getenv("TODO_PASSWORD")
	if pass == "" {
		writeJSON(w, map[string]string{"error": "auth disabled"})
		return
	}
	if in.Password != pass {
		writeJSON(w, map[string]string{"error": "invalid password"})
		return
	}
	tok, _ := makeJWT(pass)
	writeJSON(w, map[string]string{"token": tok})
}
