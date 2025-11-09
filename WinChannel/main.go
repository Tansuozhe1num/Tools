package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    "winchannel/internal/dao"
    "winchannel/internal/paths"
    "winchannel/internal/router"
    "winchannel/internal/util"
)

func main() {
    if err := paths.EnsureDirs(); err != nil {
        log.Fatalf("init dirs: %v", err)
    }
    dao.LoadUsers()

    mux := http.NewServeMux()
    router.Register(mux)

    portStr := util.GetenvDefault("PORT", "8000")
    ip := util.GetLocalIP()
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

    enableTLS := util.GetenvDefault("ENABLE_TLS", "0") == "1"
    cert := os.Getenv("TLS_CERT")
    key := os.Getenv("TLS_KEY")
    if enableTLS && cert != "" && key != "" {
        fmt.Printf("HTTPS Enabled. Use https://localhost:%s/ and https://%s:%s/\n", portStr, ip, portStr)
        log.Fatal(srv.ListenAndServeTLS(cert, key))
    } else {
        log.Fatal(srv.ListenAndServe())
    }
}