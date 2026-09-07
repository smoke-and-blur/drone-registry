package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	cookieName = "dcauth"
	sessionTTL = 30 * 24 * time.Hour
)

var (
	appPassword string
	signKey     []byte
)

// initAuth читає APP_PASSWORD; без нього сервер не стартує.
func initAuth() {
	appPassword = strings.TrimSpace(env("APP_PASSWORD", ""))
	if appPassword == "" {
		log.Fatal("APP_PASSWORD is not set: add it in the environment " +
			"(locally: copy .env.example to .env)")
	}
	// Секрет для підпису сесій. Якщо не заданий — похідний від пароля:
	// тоді зміна пароля автоматично розлогінює всіх.
	if s := env("SESSION_SECRET", ""); s != "" {
		signKey = []byte(s)
	} else {
		sum := sha256.Sum256([]byte("dc-session|" + appPassword))
		signKey = sum[:]
	}
}

// token — "expiry.signature", підписаний HMAC-SHA256.
func token(exp int64) string {
	e := strconv.FormatInt(exp, 10)
	m := hmac.New(sha256.New, signKey)
	m.Write([]byte(e))
	return e + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func validToken(v string) bool {
	e, sig, ok := strings.Cut(v, ".")
	if !ok {
		return false
	}
	exp, err := strconv.ParseInt(e, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	m := hmac.New(sha256.New, signKey)
	m.Write([]byte(e))
	want := base64.RawURLEncoding.EncodeToString(m.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(want))
}

func authed(r *http.Request) bool {
	c, err := r.Cookie(cookieName)
	return err == nil && validToken(c.Value)
}

// guard пропускає далі лише автентифікованих; сторінки редиректить
// на /login, API віддає 401 (щоб fetch не отримав HTML замість JSON).
func guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authed(r) {
			next(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if authed(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.ServeFile(w, r, "login.html")

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		// Постійний час порівняння — щоб не зливати пароль по таймінгу.
		got, want := r.FormValue("password"), appPassword
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			time.Sleep(700 * time.Millisecond) // гальмує перебір
			http.Redirect(w, r, "/login?e=1", http.StatusSeeOther)
			return
		}
		exp := time.Now().Add(sessionTTL)
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token(exp.Unix()),
			Path:     "/",
			Expires:  exp,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   isTLS(r),
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: isTLS(r),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// За проксі (Render) TLS видно лише з X-Forwarded-Proto.
func isTLS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}
