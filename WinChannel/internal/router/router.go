package router

import (
    "net/http"
    "path/filepath"
    "winchannel/internal/handlers"
    "winchannel/internal/paths"
)

func Register(mux *http.ServeMux) {
    // Pages
    mux.HandleFunc("/", handlers.ServeIndex)
    mux.HandleFunc("/login", handlers.ServeLogin)
    mux.HandleFunc("/app", handlers.ServeApp)
    mux.HandleFunc("/users", handlers.ServeUsersPage)

    // Static
    staticPath := filepath.Join(paths.BaseDir, "static")
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

    // Auth & info
    mux.HandleFunc("/api/auth/register", handlers.AuthRegister)
    mux.HandleFunc("/api/auth/login", handlers.AuthLogin)
    mux.HandleFunc("/api/auth/logout", handlers.AuthLogout)
    mux.HandleFunc("/api/auth/me", handlers.AuthMe)
    mux.HandleFunc("/api/info", handlers.ApiInfo)

    // Uploads
    mux.HandleFunc("/api/uploads", handlers.ListUploads)
    mux.HandleFunc("/api/upload", handlers.HandleUpload)
    mux.HandleFunc("/api/upload_zip", handlers.HandleUploadZip)
    mux.HandleFunc("/api/download/", handlers.HandleDownload)
    mux.HandleFunc("/api/admin/upload/", handlers.AdminUploadDelete) // DELETE /api/admin/upload/{id}
    mux.HandleFunc("/api/admin/folder/create", handlers.AdminFolderCreate)

    // Text state
    mux.HandleFunc("/api/text/state", handlers.ApiTextState)
    mux.HandleFunc("/api/text/update", handlers.ApiTextUpdate)
    mux.HandleFunc("/api/text/history", handlers.ApiTextHistory)
    
    // Admin - users
    mux.HandleFunc("/api/admin/users", handlers.AdminUsersList)
    mux.HandleFunc("/api/admin/users/create", handlers.AdminUsersCreate)
    mux.HandleFunc("/api/admin/users/update_password", handlers.AdminUsersUpdatePassword)
    mux.HandleFunc("/api/admin/users/delete", handlers.AdminUsersDelete)
}