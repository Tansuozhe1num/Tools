package service

import (
    "crypto/rand"
    "encoding/hex"
    "net/http"
    "time"
    "winchannel/internal/model"
)

const SessionCookie = "SESSION"

var Sessions = &model.SessionStore{M: map[string]model.Session{}}

func RandToken(n int) (string, error) {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

func SetSession(w http.ResponseWriter, username, role string) error {
    tok, err := RandToken(32)
    if err != nil { return err }
    s := model.Session{Username: username, Role: role, Expires: time.Now().Add(30 * 24 * time.Hour)}
    Sessions.Mu.Lock(); Sessions.M[tok] = s; Sessions.Mu.Unlock()
    http.SetCookie(w, &http.Cookie{
        Name:     SessionCookie,
        Value:    tok,
        Path:     "/",
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        Expires:  s.Expires,
    })
    return nil
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
    if c, err := r.Cookie(SessionCookie); err == nil {
        Sessions.Mu.Lock(); delete(Sessions.M, c.Value); Sessions.Mu.Unlock()
    }
    http.SetCookie(w, &http.Cookie{
        Name:     SessionCookie,
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
    })
}

func GetSession(r *http.Request) (model.Session, bool) {
    c, err := r.Cookie(SessionCookie)
    if err != nil { return model.Session{}, false }
    Sessions.Mu.Lock(); s, ok := Sessions.M[c.Value]
    if ok && time.Now().After(s.Expires) { delete(Sessions.M, c.Value); ok = false }
    Sessions.Mu.Unlock()
    return s, ok
}

func IsAdmin(r *http.Request) bool {
    s, ok := GetSession(r)
    return ok && s.Role == "admin"
}

func RequireAuth(w http.ResponseWriter, r *http.Request) bool {
    if _, ok := GetSession(r); !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return false
    }
    return true
}