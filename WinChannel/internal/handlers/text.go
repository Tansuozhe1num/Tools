package handlers

import (
    "bufio"
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
    "winchannel/internal/paths"
    "winchannel/internal/service"
    "winchannel/internal/util"
)

func readTextState() (string, int64) {
    curPath := filepath.Join(paths.TextDir, "current.txt")
    verPath := filepath.Join(paths.TextDir, "version.txt")
    content := ""
    version := int64(0)
    if b, err := os.ReadFile(curPath); err == nil { content = string(b) }
    if b, err := os.ReadFile(verPath); err == nil {
        if v, err2 := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64); err2 == nil { version = v }
    }
    return content, version
}

func writeTextState(content string, clientID string) int64 {
    curPath := filepath.Join(paths.TextDir, "current.txt")
    verPath := filepath.Join(paths.TextDir, "version.txt")
    histPath := filepath.Join(paths.TextDir, "history.ndjson")
    os.MkdirAll(paths.TextDir, 0755)
    os.WriteFile(curPath, []byte(content), 0644)
    _, old := readTextState()
    version := old + 1
    os.WriteFile(verPath, []byte(strconv.FormatInt(version, 10)), 0644)
    entry := map[string]interface{}{"version": version, "client_id": clientID, "timestamp": float64(time.Now().Unix())}
    f, err := os.OpenFile(histPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err == nil {
        enc := json.NewEncoder(f)
        enc.Encode(entry)
        f.Close()
    }
    return version
}

func ApiTextState(w http.ResponseWriter, r *http.Request) {
    if !service.RequireAuth(w, r) { return }
    c, v := readTextState()
    util.WriteJSON(w, map[string]interface{}{"content": c, "version": v, "updated_at": time.Now().Format(time.RFC3339)})
}

func ApiTextUpdate(w http.ResponseWriter, r *http.Request) {
    if !service.RequireAuth(w, r) { return }
    var body struct{
        Content string `json:"content"`
        ClientID string `json:"client_id"`
    }
    dec := json.NewDecoder(r.Body)
    if err := dec.Decode(&body); err != nil { http.Error(w, err.Error(), 400); return }
    v := writeTextState(body.Content, body.ClientID)
    util.WriteJSON(w, map[string]interface{}{"ok": true, "version": v})
}

func ApiTextHistory(w http.ResponseWriter, r *http.Request) {
    if !service.RequireAuth(w, r) { return }
    qs := r.URL.Query()
    afterStr := qs.Get("after_version")
    after := int64(-1)
    if afterStr != "" { if v, err := strconv.ParseInt(afterStr, 10, 64); err == nil { after = v } }
    histPath := filepath.Join(paths.TextDir, "history.ndjson")
    var items []map[string]interface{}
    f, err := os.Open(histPath)
    if err == nil {
        defer f.Close()
        s := bufio.NewScanner(f)
        for s.Scan() {
            var m map[string]interface{}
            if err := json.Unmarshal(s.Bytes(), &m); err == nil {
                if v, ok := m["version"].(float64); ok { if int64(v) > after { items = append(items, m) } }
            }
        }
    }
    util.WriteJSON(w, map[string]interface{}{"items": items})
}