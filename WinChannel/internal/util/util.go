package util

import (
    "encoding/json"
    "net"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
)

func WriteJSON(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    enc := json.NewEncoder(w)
    if err := enc.Encode(v); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func GetenvDefault(k, def string) string {
    v := os.Getenv(k)
    if v == "" { return def }
    return v
}

func NowTs() int64 { return time.Now().Unix() }

func IsSafePath(base, target string) bool {
    baseAbs, err1 := filepath.Abs(base)
    targetAbs, err2 := filepath.Abs(target)
    if err1 != nil || err2 != nil { return false }
    baseAbs = filepath.Clean(baseAbs)
    targetAbs = filepath.Clean(targetAbs)
    if baseAbs == targetAbs { return true }
    if !strings.HasPrefix(targetAbs, baseAbs+string(os.PathSeparator)) { return false }
    return true
}

func GetLocalIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "127.0.0.1"
    }
    defer conn.Close()
    localAddr := conn.LocalAddr().(*net.UDPAddr)
    return localAddr.IP.String()
}