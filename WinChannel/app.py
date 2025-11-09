import os
import io
import json
import time
import zipfile
import socket
import shutil
from datetime import datetime
from pathlib import Path
from flask import Flask, request, jsonify, send_file, render_template

BASE_DIR = Path(__file__).parent.resolve()
STORAGE_DIR = BASE_DIR / "storage"
UPLOAD_DIR = STORAGE_DIR / "uploads"
TEXT_DIR = STORAGE_DIR / "text"
for d in (UPLOAD_DIR, TEXT_DIR):
    d.mkdir(parents=True, exist_ok=True)

def now_ts():
    return int(time.time())

def get_local_ip():
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except:
        return "127.0.0.1"

app = Flask(__name__, static_folder=str(BASE_DIR / "static"), template_folder=str(BASE_DIR / "templates"))
app.config['MAX_CONTENT_LENGTH'] = int(os.environ.get('MAX_UPLOAD_SIZE_MB', '512')) * 1024 * 1024

@app.route("/")
def index():
    return render_template("index.html")

@app.route("/api/info")
def api_info():
    port = int(os.environ.get("PORT", "8000"))
    ip = get_local_ip()
    return jsonify({"host_ip": ip, "port": port, "urls": [f"http://localhost:{port}/", f"http://{ip}:{port}/"]})

@app.route("/api/uploads", methods=["GET"])
def api_uploads():
    items = []
    if UPLOAD_DIR.exists():
        for p in sorted(UPLOAD_DIR.iterdir()):
            if p.is_dir():
                count = 0
                size = 0
                for root, dirs, files in os.walk(p):
                    for f in files:
                        count += 1
                        fp = Path(root) / f
                        try:
                            size += fp.stat().st_size
                        except:
                            pass
                created = datetime.fromtimestamp(p.stat().st_mtime).isoformat()
                items.append({"id": p.name, "file_count": count, "size_bytes": size, "created_at": created})
    return jsonify({"uploads": items})

@app.route("/api/upload", methods=["POST"])
def api_upload():
    upload_id = request.form.get("upload_id") or f"upload-{now_ts()}"
    dest_root = UPLOAD_DIR / upload_id
    dest_root.mkdir(parents=True, exist_ok=True)
    files = request.files.getlist("files")
    saved = 0
    bytes_saved = 0
    for f in files:
        rel = f.filename.replace("\\", "/")
        folder = os.path.dirname(rel)
        base = os.path.basename(rel)
        dest_dir = dest_root / folder
        dest_dir.mkdir(parents=True, exist_ok=True)
        dest_path = dest_dir / base
        f.save(dest_path)
        try:
            bytes_saved += dest_path.stat().st_size
        except:
            pass
        saved += 1
    return jsonify({"ok": True, "upload_id": upload_id, "saved_files": saved, "size_bytes": bytes_saved})

@app.route("/api/download/<upload_id>", methods=["GET"])
def api_download(upload_id):
    dir_path = UPLOAD_DIR / upload_id
    if not dir_path.exists() or not dir_path.is_dir():
        return jsonify({"error": "not_found"}), 404
    mem = io.BytesIO()
    with zipfile.ZipFile(mem, "w", zipfile.ZIP_DEFLATED) as z:
        for root, dirs, files in os.walk(dir_path):
            for f in files:
                fp = Path(root) / f
                arcname = str(fp.relative_to(dir_path)).replace("\\", "/")
                z.write(fp, arcname)
    mem.seek(0)
    return send_file(mem, mimetype="application/zip", as_attachment=True, download_name=f"{upload_id}.zip")

def read_text_state():
    current = TEXT_DIR / "current.txt"
    version_file = TEXT_DIR / "version.txt"
    content = ""
    version = 0
    if current.exists():
        content = current.read_text(encoding="utf-8", errors="ignore")
    if version_file.exists():
        try:
            version = int(version_file.read_text().strip())
        except:
            version = 0
    return content, version

def write_text_state(content, client_id=None):
    _, old_version = read_text_state()
    current = TEXT_DIR / "current.txt"
    version_file = TEXT_DIR / "version.txt"
    history = TEXT_DIR / "history.ndjson"
    current.write_text(content, encoding="utf-8")
    version = old_version + 1
    version_file.write_text(str(version))
    entry = {"version": version, "client_id": client_id or "", "timestamp": time.time()}
    with history.open("a", encoding="utf-8") as h:
        h.write(json.dumps(entry, ensure_ascii=False) + "\n")
    return version

@app.route("/api/text/state", methods=["GET"])
def api_text_state():
    content, version = read_text_state()
    return jsonify({"content": content, "version": version, "updated_at": datetime.fromtimestamp(int(time.time())).isoformat()})

@app.route("/api/text/update", methods=["POST"])
def api_text_update():
    data = request.get_json(force=True, silent=True) or {}
    content = data.get("content", "")
    client_id = data.get("client_id")
    version = write_text_state(content, client_id=client_id)
    return jsonify({"ok": True, "version": version})

@app.route("/api/text/history", methods=["GET"])
def api_text_history():
    since = request.args.get("after_version")
    try:
        since = int(since) if since is not None else -1
    except:
        since = -1
    history = TEXT_DIR / "history.ndjson"
    items = []
    if history.exists():
        with history.open("r", encoding="utf-8", errors="ignore") as h:
            for line in h:
                try:
                    entry = json.loads(line.strip())
                    if entry.get("version", -1) > since:
                        items.append(entry)
                except:
                    pass
    return jsonify({"items": items})

# --- ZIP upload & safe extraction ---
def _is_safe_path(base: Path, target: Path) -> bool:
    try:
        base = base.resolve()
        target = target.resolve()
        return str(target).startswith(str(base))
    except Exception:
        return False

def safe_extract_zip(zf: zipfile.ZipFile, dest: Path, max_files: int = 20000):
    count = 0
    size_total = 0
    for info in zf.infolist():
        name = info.filename
        if not name or name.endswith('/'):
            continue
        # Normalize path to prevent Zip Slip
        target_path = (dest / name).resolve()
        if not _is_safe_path(dest, target_path):
            # Skip unsafe path entries
            continue
        target_path.parent.mkdir(parents=True, exist_ok=True)
        with zf.open(info) as src, open(target_path, 'wb') as out:
            shutil.copyfileobj(src, out)
        try:
            size_total += target_path.stat().st_size
        except Exception:
            pass
        count += 1
        if count > max_files:
            break
    return count, size_total

@app.route("/api/upload_zip", methods=["POST"])
def api_upload_zip():
    # Accepts a single zip file and optional upload_id, extracts into uploads/<upload_id>
    upload_id = request.form.get("upload_id") or f"zip-{now_ts()}"
    file = request.files.get("zip_file")
    if not file:
        return jsonify({"ok": False, "error": "zip_file_missing"}), 400
    dest_root = UPLOAD_DIR / upload_id
    dest_root.mkdir(parents=True, exist_ok=True)
    # Persist zip for reference
    zip_path = dest_root / f"{upload_id}.zip"
    try:
        file.save(zip_path)
    except Exception as e:
        return jsonify({"ok": False, "error": "save_failed", "detail": str(e)}), 500
    try:
        with zipfile.ZipFile(zip_path, 'r') as zf:
            extracted_count, total_size = safe_extract_zip(zf, dest_root)
    except zipfile.BadZipFile:
        return jsonify({"ok": False, "error": "bad_zip"}), 400
    except Exception as e:
        return jsonify({"ok": False, "error": "extract_failed", "detail": str(e)}), 500
    return jsonify({
        "ok": True,
        "upload_id": upload_id,
        "saved_files": extracted_count,
        "size_bytes": total_size,
        "zip_path": str(zip_path.relative_to(BASE_DIR)).replace('\\', '/')
    })

def main():
    port = int(os.environ.get("PORT", "8000"))
    ip = get_local_ip()
    print(f"Local URL: http://localhost:{port}/")
    print(f"Network URL: http://{ip}:{port}/")
    enable_tls = os.environ.get("ENABLE_TLS", "0") == "1"
    tls_cert = os.environ.get("TLS_CERT")
    tls_key = os.environ.get("TLS_KEY")

    use_waitress = os.environ.get("USE_WAITRESS", "1") == "1" and not enable_tls
    if use_waitress:
        try:
            from waitress import serve
            # threads for simple HA; production reverse-proxy recommended for internet exposure
            serve(app, host="0.0.0.0", port=port, threads=int(os.environ.get("THREADS", "8")))
            return
        except Exception as e:
            print(f"[WARN] Waitress unavailable ({e}), falling back to Flask dev server.")
    if enable_tls and tls_cert and tls_key:
        print(f"HTTPS Enabled. Use https://localhost:{port}/ and https://{ip}:{port}/")
        app.run(host="0.0.0.0", port=port, debug=False, threaded=True, ssl_context=(tls_cert, tls_key))
    else:
        app.run(host="0.0.0.0", port=port, debug=False, threaded=True)

if __name__ == "__main__":
    main()