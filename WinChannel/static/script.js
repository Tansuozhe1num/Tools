(function(){
  const $ = (sel) => document.querySelector(sel);
  const localUrlEl = $('#local-url');
  const networkUrlEl = $('#network-url');
  const folderInput = $('#folder-input');
  const uploadBtn = $('#upload-btn');
  const zipInput = $('#zip-input');
  const uploadZipBtn = $('#upload-zip-btn');
  const folderSummary = $('#folder-summary');
  const uploadsList = $('#uploads-list');
  const editor = $('#text-editor');
  const versionEl = $('#text-version');
  const statusEl = $('#sync-status');
  const historyList = $('#history-list');
  const passInput = $('#passphrase');
  const encStatus = $('#enc-status');

  let filesToUpload = [];
  let uploadId = null;
  let clientId = localStorage.getItem('winchannel_client_id');
  if (!clientId) {
    clientId = Math.random().toString(36).slice(2) + Date.now().toString(36);
    localStorage.setItem('winchannel_client_id', clientId);
  }

  function bytes(n){
    if (n < 1024) return n + ' B';
    if (n < 1024*1024) return (n/1024).toFixed(1) + ' KB';
    if (n < 1024*1024*1024) return (n/1024/1024).toFixed(1) + ' MB';
    return (n/1024/1024/1024).toFixed(2) + ' GB';
  }

  async function loadInfo(){
    const r = await fetch('/api/info');
    const data = await r.json();
    const [local, network] = data.urls || [];
    localUrlEl.textContent = local ? ('Local: ' + local) : '';
    networkUrlEl.textContent = network ? ('Network: ' + network) : '';
  }

  function detectRoot(files){
    if (!files.length) return 'folder-' + Date.now();
    const p = files[0].webkitRelativePath || files[0].name;
    const root = p.split('/')[0] || ('folder-' + Date.now());
    return root + '-' + Date.now();
  }

  folderInput.addEventListener('change', () => {
    filesToUpload = Array.from(folderInput.files || []);
    uploadId = detectRoot(filesToUpload);
    const count = filesToUpload.length;
    let size = 0;
    size = filesToUpload.reduce((s, f) => s + (f.size || 0), 0);
    folderSummary.textContent = count ? `待上传文件：${count} 个，共 ${bytes(size)}（上传ID：${uploadId}）` : '未选择文件夹';
  });

  uploadBtn.addEventListener('click', async () => {
    if (!filesToUpload.length) return alert('请先选择一个文件夹');
    statusEl.textContent = '上传中…';
    const fd = new FormData();
    fd.append('upload_id', uploadId);
    filesToUpload.forEach(f => fd.append('files', f, f.webkitRelativePath || f.name));
    const r = await fetch('/api/upload', { method: 'POST', body: fd });
    const data = await r.json();
    statusEl.textContent = data.ok ? '上传完成' : '上传失败';
    await loadUploads();
  });

  // ZIP 上传支持（移动端友好）
  let zipFile = null;
  zipInput.addEventListener('change', () => {
    zipFile = zipInput.files && zipInput.files[0] ? zipInput.files[0] : null;
    if (zipFile) {
      uploadId = 'zip-' + Date.now();
      folderSummary.textContent = `待上传 ZIP：${zipFile.name}（${bytes(zipFile.size)}），上传ID：${uploadId}`;
    } else {
      folderSummary.textContent = '未选择 ZIP';
    }
  });

  uploadZipBtn.addEventListener('click', async () => {
    if (!zipFile) return alert('请先选择一个 ZIP 文件');
    statusEl.textContent = '上传ZIP中…';
    const fd = new FormData();
    fd.append('upload_id', uploadId);
    fd.append('zip_file', zipFile, zipFile.name);
    const r = await fetch('/api/upload_zip', { method: 'POST', body: fd });
    const data = await r.json();
    statusEl.textContent = data.ok ? 'ZIP上传并解压完成' : ('上传失败：' + (data.error || ''));
    await loadUploads();
  });

  // --- 文本端到端加密（AES-GCM） ---
  let encryptionEnabled = false;
  let cryptoKey = null;
  const te = new TextEncoder();
  const td = new TextDecoder();
  const b64 = {
    encode: (arr) => btoa(String.fromCharCode(...arr)),
    decode: (str) => Uint8Array.from(atob(str), c => c.charCodeAt(0))
  };
  async function deriveKey(passphrase){
    const saltKey = 'winchannel_salt';
    let salt = localStorage.getItem(saltKey);
    if (!salt) {
      const s = crypto.getRandomValues(new Uint8Array(16));
      salt = b64.encode(s);
      localStorage.setItem(saltKey, salt);
    }
    const saltBytes = b64.decode(salt);
    const baseKey = await crypto.subtle.importKey('raw', te.encode(passphrase), 'PBKDF2', false, ['deriveKey']);
    cryptoKey = await crypto.subtle.deriveKey(
      { name: 'PBKDF2', salt: saltBytes, iterations: 100000, hash: 'SHA-256' },
      baseKey,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt','decrypt']
    );
    encryptionEnabled = true;
    encStatus.textContent = '已开启';
  }
  async function encryptText(text){
    if (!encryptionEnabled || !cryptoKey) return text;
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const ctBuf = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, cryptoKey, te.encode(text));
    const ct = new Uint8Array(ctBuf);
    return JSON.stringify({ iv: b64.encode(iv), ct: b64.encode(ct) });
  }
  async function decryptText(payload){
    try {
      const obj = JSON.parse(payload);
      if (!obj || !obj.iv || !obj.ct) return null;
      const iv = b64.decode(obj.iv);
      const ct = b64.decode(obj.ct);
      const ptBuf = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, cryptoKey, ct);
      return td.decode(ptBuf);
    } catch(e) { return null; }
  }
  passInput.addEventListener('change', async () => {
    const val = passInput.value.trim();
    if (!val) {
      encryptionEnabled = false;
      cryptoKey = null;
      encStatus.textContent = '未开启';
      return;
    }
    statusEl.textContent = '加密口令设置中…';
    await deriveKey(val);
    statusEl.textContent = '加密口令已设置';
  });

  async function loadUploads(){
    const r = await fetch('/api/uploads');
    const data = await r.json();
    uploadsList.innerHTML = '';
    (data.uploads || []).forEach(u => {
      const li = document.createElement('li');
      const left = document.createElement('div');
      left.textContent = `${u.id} · ${u.file_count} 文件 · ${bytes(u.size_bytes)} · ${u.created_at}`;
      const btn = document.createElement('a');
      btn.textContent = '下载ZIP';
      btn.href = `/api/download/${encodeURIComponent(u.id)}`;
      btn.setAttribute('download', `${u.id}.zip`);
      li.appendChild(left);
      li.appendChild(btn);
      uploadsList.appendChild(li);
    });
  }

  async function fetchTextState(){
    const r = await fetch('/api/text/state');
    const data = await r.json();
    versionEl.textContent = data.version || 0;
    if (!document.activeElement || document.activeElement !== editor) {
      const raw = data.content || '';
      if (encryptionEnabled && cryptoKey) {
        const dec = await decryptText(raw);
        editor.value = dec !== null ? dec : raw;
      } else {
        editor.value = raw;
      }
    }
    return data;
  }

  let lastVersion = 0;
  let typingTimer = null;
  editor.addEventListener('input', () => {
    statusEl.textContent = '编辑中…';
    if (typingTimer) clearTimeout(typingTimer);
    typingTimer = setTimeout(async () => {
      const encrypted = await encryptText(editor.value);
      const body = { content: encrypted, client_id: clientId };
      const r = await fetch('/api/text/update', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body)
      });
      const data = await r.json();
      versionEl.textContent = data.version || 0;
      lastVersion = data.version || 0;
      statusEl.textContent = '已同步';
      await fetchHistory();
    }, 400);
  });

  async function fetchHistory(){
    const r = await fetch(`/api/text/history?after_version=${encodeURIComponent(lastVersion || 0)}`);
    const data = await r.json();
    const items = data.items || [];
    if (items.length) {
      lastVersion = Math.max(...items.map(i => i.version));
      items.forEach(i => {
        const li = document.createElement('li');
        const dt = new Date(i.timestamp * 1000);
        li.textContent = `版本 ${i.version} · 客户端 ${i.client_id} · ${dt.toLocaleString()}`;
        historyList.prepend(li);
      });
    }
  }

  setInterval(async () => {
    const data = await fetchTextState();
    if ((data.version || 0) > lastVersion) {
      lastVersion = data.version || 0;
      await fetchHistory();
    }
  }, 1000);

  (async function init(){
    await loadInfo();
    await loadUploads();
    const state = await fetchTextState();
    lastVersion = state.version || 0;
    await fetchHistory();
  })();
})();