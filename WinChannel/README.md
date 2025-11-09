# 梦始 - WinChannel

一个跨设备文件与文本传输的本地 Web 工具（Go 后端），支持：

- 文件夹传输：选择本地目录批量上传，保留结构；支持历史列表与一键 ZIP 下载。
- ZIP 上传（移动端友好）：可直接上传 ZIP，后端自动解压并入库。
- 文本传输：内置 TXT 编辑器，跨设备近实时同步（1s 轮询），记录版本与时间；支持可选端到端加密。
- 局域网访问：同一网络的 iPhone、macOS、Windows 设备可通过浏览器访问。
- 用户系统：普通用户可注册与登录；管理员可进行上传目录管理。

---

## 快速开始（Go）

1) 启动服务：

- Windows：双击 `WinChannel/launch_go.bat`。
- 或命令行：

```
set PORT=8000
go run WinChannel\main.go
```

终端会显示 Local 与 Network URL；同网设备使用 Network URL 访问。

2) 可选：启用 HTTPS（自签或受信证书）

```
set ENABLE_TLS=1
set TLS_CERT=WinChannel\certs\server.crt
set TLS_KEY=WinChannel\certs\server.key
go run WinChannel\main.go
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

## 配置

- 端口：`PORT`（默认 8000）。
- 上传大小限制：`MAX_UPLOAD_SIZE_MB`（默认 512MB）。
- HTTPS：`ENABLE_TLS=1` 并设置 `TLS_CERT` 与 `TLS_KEY`。

> 生产建议置于 Caddy/Nginx 反向代理，并启用 HTTP/2 与压缩。

---

## 安全说明

- ZIP 解压使用路径校验，防止 Zip Slip（路径穿越）。
- 文本支持端到端加密：在页面上输入相同口令后，以 AES-GCM 加密内容，服务器只存储密文。
  - 局域网场景建议使用 `mkcert` 生成本地受信证书，并在各设备导入信任；
  - 公网场景建议使用 Caddy 自动签发证书或 Nginx + Let’s Encrypt。

## 用户与管理员功能

- 管理员账户：用户名 `dreamstartooo`，密码 `123456`。
- 普通用户：可在页面通过“注册”创建账户并登录。
- 登录后：
  - 管理员可以删除任意上传目录（列表项右侧“删除上传”）以及在上传根目录新建文件夹。
  - 普通用户仅进行上传与下载，不具备删除/新建权限。

---

## 目录结构

- `WinChannel/main.go` 后端（Go）
- `WinChannel/templates/index.html` 前端页面
- `WinChannel/static/style.css` 样式
- `WinChannel/static/script.js` 前端逻辑
- `WinChannel/storage/uploads/` 目录与 ZIP 存储（ZIP 保留在对应上传目录下）
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
- `POST /api/auth/register` 注册普通用户。
- `POST /api/auth/login` 登录（管理员 `dreamstartooo/123456` 或普通用户）。
- `POST /api/auth/logout` 退出登录。
- `GET /api/auth/me` 获取当前登录状态。
- `DELETE /api/admin/upload/:id` 管理员删除上传目录。
- `POST /api/admin/folder/create` 管理员在上传根目录新建文件夹。

---

## 常见问题

- iOS Safari 不能直接选文件夹？请先在“文件”App 中将文件夹压缩为 ZIP 后上传。
- 上传中断或失败？请提高 `MAX_UPLOAD_SIZE_MB`，并检查网络稳定性。
- 局域网访问不到？请确认设备在同一网络，并关闭系统防火墙阻挡或允许进程入站。