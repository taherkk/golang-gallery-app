package controllers

import (
	"fmt"
	"net/http"
)

const (
	CookieSession = "session"
)

func newCookie(name string, value string) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
	}
	return &cookie
}

func setCookie(w http.ResponseWriter, name string, value string) {
	cookie := newCookie(name, value)
	http.SetCookie(w, cookie)
}

func deleteCookie(w http.ResponseWriter, name string) {
	cookie := newCookie(name, "")
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

// not completely necessary
func readCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("cookie not found: %w", err)
	}

	return cookie.Value, nil
}
