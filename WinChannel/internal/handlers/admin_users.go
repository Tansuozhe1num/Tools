package handlers

import (
    "encoding/json"
    "net/http"
    "strings"
    "winchannel/internal/dao"
    "winchannel/internal/util"
    "winchannel/internal/service"
    "golang.org/x/crypto/bcrypt"
)

func AdminUsersList(w http.ResponseWriter, r *http.Request) {
    if !service.IsAdmin(r) { http.Error(w, "admin required", 401); return }
    if r.Method != http.MethodGet { http.Error(w, "method not allowed", 405); return }
    type u struct { Username string `json:"username"`; Role string `json:"role"` }
    var list []u
    list = append(list, u{Username: "dreamstartooo", Role: "admin"})
    dao.Users.Mu.Lock()
    for name := range dao.Users.Users {
        list = append(list, u{Username: name, Role: "user"})
    }
    dao.Users.Mu.Unlock()
    util.WriteJSON(w, map[string]interface{}{"users": list})
}

func AdminUsersCreate(w http.ResponseWriter, r *http.Request) {
    if !service.IsAdmin(r) { http.Error(w, "admin required", 401); return }
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", 405); return }
    var in struct{ Username, Password string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w, "bad json", 400); return }
    in.Username = strings.TrimSpace(in.Username)
    if in.Username == "" || in.Password == "" { http.Error(w, "empty username or password", 400); return }
    if in.Username == "dreamstartooo" { http.Error(w, "reserved admin username", 400); return }
    dao.Users.Mu.Lock()
    if _, exists := dao.Users.Users[in.Username]; exists { dao.Users.Mu.Unlock(); http.Error(w, "user exists", 409); return }
    dao.Users.Mu.Unlock()
    hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
    if err != nil { http.Error(w, "hash error", 500); return }
    dao.Users.Mu.Lock(); dao.Users.Users[in.Username] = string(hash); dao.Users.Mu.Unlock()
    if err := dao.SaveUsers(); err != nil { http.Error(w, "save error", 500); return }
    util.WriteJSON(w, map[string]interface{}{"ok": true})
}

func AdminUsersUpdatePassword(w http.ResponseWriter, r *http.Request) {
    if !service.IsAdmin(r) { http.Error(w, "admin required", 401); return }
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", 405); return }
    var in struct{ Username, NewPassword string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w, "bad json", 400); return }
    in.Username = strings.TrimSpace(in.Username)
    if in.Username == "dreamstartooo" { http.Error(w, "cannot modify admin", 400); return }
    dao.Users.Mu.Lock()
    if _, exists := dao.Users.Users[in.Username]; !exists { dao.Users.Mu.Unlock(); http.Error(w, "user not found", 404); return }
    dao.Users.Mu.Unlock()
    hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
    if err != nil { http.Error(w, "hash error", 500); return }
    dao.Users.Mu.Lock(); dao.Users.Users[in.Username] = string(hash); dao.Users.Mu.Unlock()
    if err := dao.SaveUsers(); err != nil { http.Error(w, "save error", 500); return }
    util.WriteJSON(w, map[string]interface{}{"ok": true})
}

func AdminUsersDelete(w http.ResponseWriter, r *http.Request) {
    if !service.IsAdmin(r) { http.Error(w, "admin required", 401); return }
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", 405); return }
    var in struct{ Username string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w, "bad json", 400); return }
    in.Username = strings.TrimSpace(in.Username)
    if in.Username == "dreamstartooo" { http.Error(w, "cannot delete admin", 400); return }
    dao.Users.Mu.Lock()
    if _, exists := dao.Users.Users[in.Username]; !exists { dao.Users.Mu.Unlock(); http.Error(w, "user not found", 404); return }
    delete(dao.Users.Users, in.Username)
    dao.Users.Mu.Unlock()
    if err := dao.SaveUsers(); err != nil { http.Error(w, "save error", 500); return }
    util.WriteJSON(w, map[string]interface{}{"ok": true})
}