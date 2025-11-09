package handlers

import (
    "archive/zip"
    "encoding/json"
    "fmt"
    "io"
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

func ListUploads(w http.ResponseWriter, r *http.Request) {
    if !service.RequireAuth(w, r) { return }
    type item struct {
        ID         string `json:"id"`
        FileCount  int    `json:"file_count"`
        SizeBytes  int64  `json:"size_bytes"`
        CreatedAt  string `json:"created_at"`
    }
    var items []item
    entries, _ := os.ReadDir(paths.UploadsDir)
    for _, e := range entries {
        if !e.IsDir() { continue }
        dirPath := filepath.Join(paths.UploadsDir, e.Name())
        var count int
        var size int64
        filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
            if err != nil { return nil }
            if !info.IsDir() { count++; size += info.Size() }
            return nil
        })
        fi, _ := os.Stat(dirPath)
        created := fi.ModTime().Format(time.RFC3339)
        items = append(items, item{ID: e.Name(), FileCount: count, SizeBytes: size, CreatedAt: created})
    }
    util.WriteJSON(w, map[string]interface{}{"uploads": items})
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
    if !service.RequireAuth(w, r) { return }
    maxMB, _ := strconv.Atoi(util.GetenvDefault("MAX_UPLOAD_SIZE_MB", "512"))
    r.Body = http.MaxBytesReader(w, r.Body, int64(maxMB)*1024*1024)
    if err := r.ParseMultipartForm(int64(maxMB) * 1024 * 1024); err != nil {
        http.Error(w, fmt.Sprintf("parse form: %v", err), http.StatusBadRequest)
        return
    }
    uploadID := r.FormValue("upload_id")
    if uploadID == "" { uploadID = fmt.Sprintf("upload-%d", util.NowTs()) }
    destRoot := filepath.Join(paths.UploadsDir, uploadID)
    if err := os.MkdirAll(destRoot, 0755); err != nil { http.Error(w, err.Error(), 500); return }
    files := r.MultipartForm.File["files"]
    var saved int
    var bytesSaved int64
    for _, fh := range files {
        rel := fh.Filename
        rel = filepath.FromSlash(rel)
        target := filepath.Join(destRoot, rel)
        if !util.IsSafePath(destRoot, target) { continue }
        if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil { continue }
        src, err := fh.Open(); if err != nil { continue }
        defer src.Close()
        out, err := os.Create(target); if err != nil { continue }
        n, _ := io.Copy(out, src)
        out.Close()
        bytesSaved += n; saved++
    }
    util.WriteJSON(w, map[string]interface{}{"ok": true, "upload_id": uploadID, "saved_files": saved, "size_bytes": bytesSaved})
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
    if !service.RequireAuth(w, r) { return }
    prefix := "/api/download/"
    uploadID := strings.TrimPrefix(r.URL.Path, prefix)
    if uploadID == "" { http.NotFound(w, r); return }
    dirPath := filepath.Join(paths.UploadsDir, uploadID)
    if fi, err := os.Stat(dirPath); err != nil || !fi.IsDir() {
        util.WriteJSON(w, map[string]string{"error": "not_found"}); return
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
        writer, err := zw.CreateHeader(hdr); if err != nil { return nil }
        f, err := os.Open(path); if err != nil { return nil }
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
        if !util.IsSafePath(destRoot, target) { continue }
        if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil { continue }
        rc, err := f.Open(); if err != nil { continue }
        out, err := os.Create(target); if err != nil { rc.Close(); continue }
        n, _ := io.Copy(out, rc)
        out.Close(); rc.Close()
        total += n; count++
        if count > maxFiles { break }
    }
    return count, total
}

func HandleUploadZip(w http.ResponseWriter, r *http.Request) {
    if !service.RequireAuth(w, r) { return }
    maxMB, _ := strconv.Atoi(util.GetenvDefault("MAX_UPLOAD_SIZE_MB", "512"))
    r.Body = http.MaxBytesReader(w, r.Body, int64(maxMB)*1024*1024)
    if err := r.ParseMultipartForm(int64(maxMB) * 1024 * 1024); err != nil { http.Error(w, fmt.Sprintf("parse form: %v", err), 400); return }
    uploadID := r.FormValue("upload_id")
    if uploadID == "" { uploadID = fmt.Sprintf("zip-%d", util.NowTs()) }
    destRoot := filepath.Join(paths.UploadsDir, uploadID)
    if err := os.MkdirAll(destRoot, 0755); err != nil { http.Error(w, err.Error(), 500); return }
    fh := r.MultipartForm.File["zip_file"]
    if len(fh) == 0 { util.WriteJSON(w, map[string]interface{}{"ok": false, "error": "zip_file_missing"}); return }
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
    util.WriteJSON(w, map[string]interface{}{"ok": true, "upload_id": uploadID, "saved_files": extracted, "size_bytes": total, "zip_path": strings.ReplaceAll(zipPath, "\\", "/")})
}

func AdminUploadDelete(w http.ResponseWriter, r *http.Request) {
    if !service.IsAdmin(r) { http.Error(w, "admin required", 401); return }
    if r.Method != http.MethodDelete { http.Error(w, "method not allowed", 405); return }
    uploadID := strings.TrimPrefix(r.URL.Path, "/api/admin/upload/")
    if uploadID == "" { http.Error(w, "missing upload id", 400); return }
    target := filepath.Join(paths.UploadsDir, uploadID)
    if !util.IsSafePath(paths.UploadsDir, target) { http.Error(w, "invalid path", 400); return }
    if err := os.RemoveAll(target); err != nil { http.Error(w, "delete error", 500); return }
    util.WriteJSON(w, map[string]interface{}{"ok": true})
}

func AdminFolderCreate(w http.ResponseWriter, r *http.Request) {
    if !service.IsAdmin(r) { http.Error(w, "admin required", 401); return }
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", 405); return }
    var in struct{ Name string `json:"name"` }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil { http.Error(w, "bad json", 400); return }
    in.Name = strings.TrimSpace(in.Name)
    if in.Name == "" { http.Error(w, "empty name", 400); return }
    if strings.Contains(in.Name, "..") || strings.ContainsAny(in.Name, "/\\") { http.Error(w, "invalid name", 400); return }
    target := filepath.Join(paths.UploadsDir, in.Name)
    if !util.IsSafePath(paths.UploadsDir, target) { http.Error(w, "invalid path", 400); return }
    if err := os.MkdirAll(target, 0755); err != nil { http.Error(w, "mkdir error", 500); return }
    util.WriteJSON(w, map[string]interface{}{"ok": true})
}