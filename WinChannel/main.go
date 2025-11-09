package main

import (
    "archive/zip"
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
)

type InfoResponse struct {
    HostIP string   `json:"host_ip"`
    Port   int      `json:"port"`
    Urls   []string `json:"urls"`
}

var (
    baseDir    = filepath.Join(".", "WinChannel")
    storageDir = filepath.Join(baseDir, "storage")
    uploadsDir = filepath.Join(storageDir, "uploads")
    textDir    = filepath.Join(storageDir, "text")
)

func ensureDirs() {
    for _, d := range []string{uploadsDir, textDir} {
        if err := os.MkdirAll(d, 0755); err != nil {
            log.Fatalf("mkdir %s: %v", d, err)
        }
    }
}

func nowTs() int64 { return time.Now().Unix() }

func getLocalIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "127.0.0.1"
    }
    defer conn.Close()
    localAddr := conn.LocalAddr().(*net.UDPAddr)
    return localAddr.IP.String()
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, filepath.Join(baseDir, "templates", "index.html"))
}

func apiInfo(w http.ResponseWriter, r *http.Request) {
    portStr := getenvDefault("PORT", "8000")
    port, _ := strconv.Atoi(portStr)
    ip := getLocalIP()
    writeJSON(w, InfoResponse{HostIP: ip, Port: port, Urls: []string{fmt.Sprintf("http://localhost:%d/", port), fmt.Sprintf("http://%s:%d/", ip, port)}})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    enc := json.NewEncoder(w)
    if err := enc.Encode(v); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func listUploads(w http.ResponseWriter, r *http.Request) {
    type item struct {
        ID         string `json:"id"`
        FileCount  int    `json:"file_count"`
        SizeBytes  int64  `json:"size_bytes"`
        CreatedAt  string `json:"created_at"`
    }
    var items []item
    entries, _ := os.ReadDir(uploadsDir)
    for _, e := range entries {
        if !e.IsDir() { continue }
        dirPath := filepath.Join(uploadsDir, e.Name())
        var count int
        var size int64
        filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
            if err != nil { return nil }
            if !info.IsDir() {
                count++
                size += info.Size()
            }
            return nil
        })
        fi, _ := os.Stat(dirPath)
        created := fi.ModTime().Format(time.RFC3339)
        items = append(items, item{ID: e.Name(), FileCount: count, SizeBytes: size, CreatedAt: created})
    }
    writeJSON(w, map[string]interface{}{"uploads": items})
}

func getenvDefault(k, def string) string {
    v := os.Getenv(k)
    if v == "" { return def }
    return v
}

func isSafePath(base, target string) bool {
    baseAbs, err1 := filepath.Abs(base)
    targetAbs, err2 := filepath.Abs(target)
    if err1 != nil || err2 != nil { return false }
    baseAbs = filepath.Clean(baseAbs)
    targetAbs = filepath.Clean(targetAbs)
    if baseAbs == targetAbs { return true }
    // ensure target under base
    if !strings.HasPrefix(targetAbs, baseAbs+string(os.PathSeparator)) { return false }
    return true
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
    maxMB, _ := strconv.Atoi(getenvDefault("MAX_UPLOAD_SIZE_MB", "512"))
    r.Body = http.MaxBytesReader(w, r.Body, int64(maxMB)*1024*1024)
    if err := r.ParseMultipartForm(int64(maxMB) * 1024 * 1024); err != nil {
        http.Error(w, fmt.Sprintf("parse form: %v", err), http.StatusBadRequest)
        return
    }
    uploadID := r.FormValue("upload_id")
    if uploadID == "" { uploadID = fmt.Sprintf("upload-%d", nowTs()) }
    destRoot := filepath.Join(uploadsDir, uploadID)
    if err := os.MkdirAll(destRoot, 0755); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    files := r.MultipartForm.File["files"]
    var saved int
    var bytesSaved int64
    for _, fh := range files {
        rel := fh.Filename // may contain webkitRelativePath
        rel = filepath.FromSlash(rel)
        target := filepath.Join(destRoot, rel)
        if !isSafePath(destRoot, target) { continue }
        if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil { continue }
        src, err := fh.Open()
        if err != nil { continue }
        defer src.Close()
        out, err := os.Create(target)
        if err != nil { continue }
        n, _ := io.Copy(out, src)
        out.Close()
        bytesSaved += n
        saved++
    }
    writeJSON(w, map[string]interface{}{"ok": true, "upload_id": uploadID, "saved_files": saved, "size_bytes": bytesSaved})
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
    prefix := "/api/download/"
    uploadID := strings.TrimPrefix(r.URL.Path, prefix)
    if uploadID == "" {
        http.NotFound(w, r); return
    }
    dirPath := filepath.Join(uploadsDir, uploadID)
    if fi, err := os.Stat(dirPath); err != nil || !fi.IsDir() {
        writeJSON(w, map[string]string{"error": "not_found"}); return
    }
    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", uploadID))
    zw := zip.NewWriter(w)
    defer zw.Close()
    filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil { return nil }
        if info.IsDir() { return nil }
        rel, _ := filepath.Rel(dirPath, path)
        rel = filepath.ToSlash(rel)
        hdr := &zip.FileHeader{Name: rel, Method: zip.Deflate}
        hdr.SetModTime(info.ModTime())
        writer, err := zw.CreateHeader(hdr)
        if err != nil { return nil }
        f, err := os.Open(path)
        if err != nil { return nil }
        defer f.Close()
        io.Copy(writer, f)
        return nil
    })
}

func safeExtractZip(zr *zip.Reader, destRoot string, maxFiles int) (int, int64) {
    var count int
    var total int64
    for _, f := range zr.File {
        name := f.Name
        if name == "" || strings.HasSuffix(name, "/") { continue }
        target := filepath.Join(destRoot, filepath.FromSlash(name))
        if !isSafePath(destRoot, target) { continue }
        if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil { continue }
        rc, err := f.Open()
        if err != nil { continue }
        out, err := os.Create(target)
        if err != nil { rc.Close(); continue }
        n, _ := io.Copy(out, rc)
        out.Close(); rc.Close()
        total += n
        count++
        if count > maxFiles { break }
    }
    return count, total
}

func handleUploadZip(w http.ResponseWriter, r *http.Request) {
    maxMB, _ := strconv.Atoi(getenvDefault("MAX_UPLOAD_SIZE_MB", "512"))
    r.Body = http.MaxBytesReader(w, r.Body, int64(maxMB)*1024*1024)
    if err := r.ParseMultipartForm(int64(maxMB) * 1024 * 1024); err != nil {
        http.Error(w, fmt.Sprintf("parse form: %v", err), http.StatusBadRequest)
        return
    }
    uploadID := r.FormValue("upload_id")
    if uploadID == "" { uploadID = fmt.Sprintf("zip-%d", nowTs()) }
    destRoot := filepath.Join(uploadsDir, uploadID)
    if err := os.MkdirAll(destRoot, 0755); err != nil { http.Error(w, err.Error(), 500); return }
    fh := r.MultipartForm.File["zip_file"]
    if len(fh) == 0 { writeJSON(w, map[string]interface{}{"ok": false, "error": "zip_file_missing"}); return }
    file := fh[0]
    zipPath := filepath.Join(destRoot, fmt.Sprintf("%s.zip", uploadID))
    src, err := file.Open(); if err != nil { http.Error(w, err.Error(), 500); return }
    out, err := os.Create(zipPath); if err != nil { src.Close(); http.Error(w, err.Error(), 500); return }
    n, _ := io.Copy(out, src)
    out.Close(); src.Close()
    fi, err := os.Stat(zipPath); if err != nil { http.Error(w, err.Error(), 500); return }
    zr, err := zip.OpenReader(zipPath); if err != nil { http.Error(w, "bad_zip", 400); return }
    defer zr.Close()
    extracted, total := safeExtractZip(&zr.Reader, destRoot, 20000)
    _ = n; _ = fi
    writeJSON(w, map[string]interface{}{"ok": true, "upload_id": uploadID, "saved_files": extracted, "size_bytes": total, "zip_path": strings.ReplaceAll(zipPath, "\\", "/")})
}

func readTextState() (string, int64) {
    curPath := filepath.Join(textDir, "current.txt")
    verPath := filepath.Join(textDir, "version.txt")
    content := ""
    version := int64(0)
    if b, err := os.ReadFile(curPath); err == nil { content = string(b) }
    if b, err := os.ReadFile(verPath); err == nil {
        if v, err2 := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64); err2 == nil { version = v }
    }
    return content, version
}

func writeTextState(content string, clientID string) int64 {
    curPath := filepath.Join(textDir, "current.txt")
    verPath := filepath.Join(textDir, "version.txt")
    histPath := filepath.Join(textDir, "history.ndjson")
    os.MkdirAll(textDir, 0755)
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

func apiTextState(w http.ResponseWriter, r *http.Request) {
    c, v := readTextState()
    writeJSON(w, map[string]interface{}{"content": c, "version": v, "updated_at": time.Now().Format(time.RFC3339)})
}

func apiTextUpdate(w http.ResponseWriter, r *http.Request) {
    var body struct{
        Content string `json:"content"`
        ClientID string `json:"client_id"`
    }
    dec := json.NewDecoder(r.Body)
    if err := dec.Decode(&body); err != nil { http.Error(w, err.Error(), 400); return }
    v := writeTextState(body.Content, body.ClientID)
    writeJSON(w, map[string]interface{}{"ok": true, "version": v})
}

func apiTextHistory(w http.ResponseWriter, r *http.Request) {
    qs := r.URL.Query()
    afterStr := qs.Get("after_version")
    after := int64(-1)
    if afterStr != "" { if v, err := strconv.ParseInt(afterStr, 10, 64); err == nil { after = v } }
    histPath := filepath.Join(textDir, "history.ndjson")
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
    writeJSON(w, map[string]interface{}{"items": items})
}

func main() {
    ensureDirs()
    mux := http.NewServeMux()
    mux.HandleFunc("/", serveIndex)
    mux.HandleFunc("/api/info", apiInfo)
    mux.HandleFunc("/api/uploads", listUploads)
    mux.HandleFunc("/api/upload", handleUpload)
    mux.HandleFunc("/api/upload_zip", handleUploadZip)
    mux.HandleFunc("/api/text/state", apiTextState)
    mux.HandleFunc("/api/text/update", apiTextUpdate)
    mux.HandleFunc("/api/text/history", apiTextHistory)
    mux.HandleFunc("/api/download/", handleDownload)

    // static files
    staticPath := filepath.Join(baseDir, "static")
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

    portStr := getenvDefault("PORT", "8000")
    ip := getLocalIP()
    fmt.Printf("Local URL: http://localhost:%s/\n", portStr)
    fmt.Printf("Network URL: http://%s:%s/\n", ip, portStr)

    srv := &http.Server{
        Addr:              ":" + portStr,
        Handler:           mux,
        ReadTimeout:       10 * time.Minute,
        WriteTimeout:      10 * time.Minute,
        ReadHeaderTimeout: 15 * time.Second,
        IdleTimeout:       2 * time.Minute,
    }

    enableTLS := getenvDefault("ENABLE_TLS", "0") == "1"
    cert := os.Getenv("TLS_CERT")
    key := os.Getenv("TLS_KEY")
    if enableTLS && cert != "" && key != "" {
        fmt.Printf("HTTPS Enabled. Use https://localhost:%s/ and https://%s:%s/\n", portStr, ip, portStr)
        log.Fatal(srv.ListenAndServeTLS(cert, key))
    } else {
        log.Fatal(srv.ListenAndServe())
    }
}