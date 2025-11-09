@echo off
set PORT=8000
rem 启用 HTTPS：将 ENABLE_TLS=1 并设置 TLS_CERT/TLS_KEY 路径
rem set ENABLE_TLS=1
rem set TLS_CERT=WinChannel\certs\server.crt
rem set TLS_KEY=WinChannel\certs\server.key
go run "WinChannel\main.go"