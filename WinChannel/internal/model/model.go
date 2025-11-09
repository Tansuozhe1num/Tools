package model

import (
    "sync"
    "time"
)

type UserStore struct {
    Mu    sync.Mutex
    Users map[string]string // username -> bcrypt hash
}

type Session struct {
    Username string
    Role     string
    Expires  time.Time
}

type SessionStore struct {
    Mu sync.Mutex
    M  map[string]Session
}