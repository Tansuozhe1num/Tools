@echo off
set PORT=8000
set USE_WAITRESS=1
set THREADS=8
rem 可选：最大上传大小（MB）。默认 512。
set MAX_UPLOAD_SIZE_MB=512
python "%~dp0app.py"