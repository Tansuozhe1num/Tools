# 梦始 - WinChannel

一个跨设备文件与文本传输的本地 Web 工具，支持：

- 文件夹传输：选择本地目录批量上传，保留结构；支持历史列表与一键 ZIP 下载。
- ZIP 上传（移动端友好）：可直接上传 ZIP，后端自动解压并入库。
- 文本传输：内置 TXT 编辑器，跨设备近实时同步（1s 轮询），记录版本与时间；支持可选端到端加密。
- 局域网访问：同一网络的 iPhone、macOS、Windows 设备可通过浏览器访问。

---

## 快速开始

1) 安装依赖（首次运行）：

```
pip install -r WinChannel/requirements.txt
```

2) 启动（推荐）：

- Windows：双击 `WinChannel/launch.bat`（默认使用 Waitress 高可用模式）。
- 或命令行：

```
set PORT=8000
set USE_WAITRESS=1
python WinChannel/app.py
```

启动后终端显示 Local/Network URL；同网设备使用 Network URL 访问。

3) 可选：启用 HTTPS（自签）

- 生成证书（示例，推荐使用 `mkcert` 在本机与设备上建立受信 CA）：
- 将证书路径配置到环境变量，并关闭 Waitress：

```
set USE_WAITRESS=0
set ENABLE_TLS=1
set TLS_CERT=WinChannel\certs\server.crt
set TLS_KEY=WinChannel\certs\server.key
python WinChannel\app.py
```

---

## 使用指南

### 文件夹上传（桌面 Chrome/Edge）
- 页面“文件夹传输”中点击“选择文件夹”，选择一个目录；点击“开始上传”。
- 上传完成后在“已存储的上传”中可看到记录，并可“下载ZIP”。

### ZIP 上传（移动端适配）
- iPhone 或 Android 可先在文件管理中将文件夹压缩为 `.zip`。
- 页面中选择 ZIP 文件后，点击“上传ZIP并解压”；后端会自动解压到 `storage/uploads/<upload_id>/`。

### 文本传输
- 在“文本传输（实时同步）”编辑器中输入文本，几百毫秒后自动保存并同步。
- 任何设备修改都会提升版本并写入历史（右侧显示版本与状态，底部显示历史）。

---

## 高可用与配置

- 服务器：默认使用 Waitress（多线程）作为 WSGI，提升稳定性与并发处理能力；如需 HTTPS，切换为 Flask TLS 模式或置于 Caddy/Nginx 反向代理后启用 HTTPS。
- 线程数：`THREADS` 环境变量（默认 8）。
- 上传大小限制：`MAX_UPLOAD_SIZE_MB`（默认 512MB）。
- 端口：`PORT`（默认 8000）。

示例（`launch.bat` 已内置）：

```
set PORT=8000
set USE_WAITRESS=1
set THREADS=8
set MAX_UPLOAD_SIZE_MB=512
python WinChannel/app.py
```

> 生产环境建议置于反向代理（如 Nginx）之后，并结合端口映射/NAT 或内网穿透进行公网访问。

---

## 安全说明

- ZIP 解压使用路径校验，防止 Zip Slip（路径穿越）。
- 文本支持端到端加密：在页面上输入相同口令后，以 AES-GCM 加密内容，服务器只存储密文。
- 如需严格 HTTPS：
  - 局域网场景建议使用 `mkcert` 生成本地受信证书，并在各设备导入信任；
  - 公网场景建议使用 Caddy 自动签发证书或 Nginx + Let’s Encrypt。

---

## Go 后端建议（可选）

- 若追求更高吞吐与更低资源占用，可采用 Go（如 Gin/Fiber）实现后端，TLS 原生支持更完善。
- 迁移路径：
  - 复刻 REST API：`/api/upload`, `/api/upload_zip`, `/api/uploads`, `/api/download/:id`, `/api/text/*`；
  - 静态与模板资源保持同路径；
  - 启用 HTTP/2 + TLS，首选 Caddy 作为前端（自动证书与压缩）。
  - 可在接入层做内容压缩与限速，提升稳定性。
- 文件上传保留相对路径；不允许写到工作目录之外。
- 历史记录采用追加写入，不包含用户内容，仅记录版本与时间。

---

## 目录结构

- `WinChannel/app.py` 后端（Flask + Waitress）
- `WinChannel/templates/index.html` 前端页面
- `WinChannel/static/style.css` 样式
- `WinChannel/static/script.js` 前端逻辑
- `WinChannel/storage/uploads/` 目录与 ZIP 存储（ZIP 会保留在对应上传目录下）
- `WinChannel/storage/text/` 文本内容与历史记录

---

## API 摘要

- `GET /api/info` 本机与局域网的访问地址。
- `POST /api/upload` 上传整目录文件（`webkitdirectory`）。
- `POST /api/upload_zip` 上传 ZIP 并安全解压入库。
- `GET /api/uploads` 列出所有上传集（文件数、大小、时间）。
- `GET /api/download/:upload_id` 下载指定上传为 ZIP。
- `GET /api/text/state` 获取当前文本与版本。
- `POST /api/text/update` 更新文本并记录版本。
- `GET /api/text/history?after_version=n` 拉取增量历史。

---

## 常见问题

- iOS Safari 不能直接选文件夹？请先在“文件”App 中将文件夹压缩为 ZIP 后上传。
- 上传中断或失败？请提高 `MAX_UPLOAD_SIZE_MB`，并检查网络稳定性。
- 局域网访问不到？请确认设备在同一网络，并关闭系统防火墙的阻挡或为 Python 进程允许入站。