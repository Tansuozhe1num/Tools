package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "winchannel/internal/dao"
    "winchannel/internal/model"
    "winchannel/internal/service"
    "winchannel/internal/util"
    "golang.org/x/crypto/bcrypt"
)

func AuthRegister(w http.ResponseWriter, r *http.Request) {
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
    _ = service.SetSession(w, in.Username, "user")
    util.WriteJSON(w, map[string]interface{}{"ok": true, "username": in.Username, "role": "user"})
}

func AuthLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", 405); return }
    var in struct{ Username, Password string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w, "bad json", 400); return }
    in.Username = strings.TrimSpace(in.Username)
    if in.Username == "dreamstartooo" {
        if in.Password != "123456" { http.Error(w, "invalid admin password", 401); return }
        _ = service.SetSession(w, in.Username, "admin")
        util.WriteJSON(w, map[string]interface{}{"ok": true, "username": in.Username, "role": "admin"})
        return
    }
    dao.Users.Mu.Lock(); hash, ok := dao.Users.Users[in.Username]; dao.Users.Mu.Unlock()
    if !ok { http.Error(w, "user not found", 404); return }
    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)); err != nil { http.Error(w, "invalid password", 401); return }
    _ = service.SetSession(w, in.Username, "user")
    util.WriteJSON(w, map[string]interface{}{"ok": true, "username": in.Username, "role": "user"})
}

func AuthLogout(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", 405); return }
    service.ClearSession(w, r)
    util.WriteJSON(w, map[string]interface{}{"ok": true})
}

func AuthMe(w http.ResponseWriter, r *http.Request) {
    if s, ok := service.GetSession(r); ok {
        util.WriteJSON(w, map[string]interface{}{"authenticated": true, "username": s.Username, "role": s.Role})
        return
    }
    util.WriteJSON(w, map[string]interface{}{"authenticated": false})
}

func ApiInfo(w http.ResponseWriter, r *http.Request) {
    portStr := util.GetenvDefault("PORT", "8000")
    port, _ := strconv.Atoi(portStr)
    ip := util.GetLocalIP()
    util.WriteJSON(w, model.InfoResponse{HostIP: ip, Port: port, Urls: []string{"http://localhost:" + portStr + "/", "http://" + ip + ":" + portStr + "/"}})
}