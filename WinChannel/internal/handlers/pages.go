package handlers

import (
    "net/http"
    "path/filepath"
    "winchannel/internal/paths"
    "winchannel/internal/service"
)

func ServeIndex(w http.ResponseWriter, r *http.Request) {
    if _, ok := service.GetSession(r); ok {
        http.Redirect(w, r, "/app", http.StatusFound)
        return
    }
    http.Redirect(w, r, "/login", http.StatusFound)
}

func ServeLogin(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, filepath.Join(paths.BaseDir, "templates", "login.html"))
}

func ServeApp(w http.ResponseWriter, r *http.Request) {
    if _, ok := service.GetSession(r); !ok {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }
    http.ServeFile(w, r, filepath.Join(paths.BaseDir, "templates", "app.html"))
}

func ServeUsersPage(w http.ResponseWriter, r *http.Request) {
    if !service.IsAdmin(r) {
        if _, ok := service.GetSession(r); ok {
            http.Redirect(w, r, "/app", http.StatusFound)
        } else {
            http.Redirect(w, r, "/login", http.StatusFound)
        }
        return
    }
    http.ServeFile(w, r, filepath.Join(paths.BaseDir, "templates", "users.html"))
}