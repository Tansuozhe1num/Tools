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
  // Auth & admin controls
  const authUsername = document.querySelector('#auth-username');
  const authPassword = document.querySelector('#auth-password');
  const btnRegister = document.querySelector('#btn-register');
  const btnLogin = document.querySelector('#btn-login');
  const btnLogout = document.querySelector('#btn-logout');
  const authStatus = document.querySelector('#auth-status');
  const newFolderName = document.querySelector('#new-folder-name');
  const createFolderBtn = document.querySelector('#create-folder-btn');
  const adminControls = document.querySelectorAll('.admin-only');

  // Page detection
  const isLoginPage = !!document.querySelector('#login-root');
  const isAppPage = !!document.querySelector('#file-card');
  const isUsersPage = !!document.querySelector('#users-page');

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
    const r = await apiFetch('/api/info');
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

  if (folderInput) {
    folderInput.addEventListener('change', () => {
      filesToUpload = Array.from(folderInput.files || []);
      uploadId = detectRoot(filesToUpload);
      const count = filesToUpload.length;
      let size = 0;
      size = filesToUpload.reduce((s, f) => s + (f.size || 0), 0);
      if (folderSummary) {
        folderSummary.textContent = count ? `待上传文件：${count} 个，共 ${bytes(size)}（上传ID：${uploadId}）` : '未选择文件夹';
      }
    });
  }

  if (uploadBtn) {
    uploadBtn.addEventListener('click', async () => {
      if (!filesToUpload.length) return alert('请先选择一个文件夹');
      if (statusEl) statusEl.textContent = '上传中…';
      const fd = new FormData();
      fd.append('upload_id', uploadId);
      filesToUpload.forEach(f => fd.append('files', f, f.webkitRelativePath || f.name));
      const r = await apiFetch('/api/upload', { method: 'POST', body: fd });
      const data = await r.json();
      if (statusEl) statusEl.textContent = data.ok ? '上传完成' : '上传失败';
      await loadUploads();
    });
  }

  // ZIP 上传支持（移动端友好）
  let zipFile = null;
  if (zipInput) {
    zipInput.addEventListener('change', () => {
      zipFile = zipInput.files && zipInput.files[0] ? zipInput.files[0] : null;
      if (zipFile) {
        uploadId = 'zip-' + Date.now();
        if (folderSummary) folderSummary.textContent = `待上传 ZIP：${zipFile.name}（${bytes(zipFile.size)}），上传ID：${uploadId}`;
      } else {
        if (folderSummary) folderSummary.textContent = '未选择 ZIP';
      }
    });
  }

  if (uploadZipBtn) {
    uploadZipBtn.addEventListener('click', async () => {
      if (!zipFile) return alert('请先选择一个 ZIP 文件');
      if (statusEl) statusEl.textContent = '上传ZIP中…';
      const fd = new FormData();
      fd.append('upload_id', uploadId);
      fd.append('zip_file', zipFile, zipFile.name);
      const r = await apiFetch('/api/upload_zip', { method: 'POST', body: fd });
      const data = await r.json();
      if (statusEl) statusEl.textContent = data.ok ? 'ZIP上传并解压完成' : ('上传失败：' + (data.error || ''));
      await loadUploads();
    });
  }

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
  if (passInput) {
    passInput.addEventListener('change', async () => {
      const val = passInput.value.trim();
      if (!val) {
        encryptionEnabled = false;
        cryptoKey = null;
        if (encStatus) encStatus.textContent = '未开启';
        return;
      }
      if (statusEl) statusEl.textContent = '加密口令设置中…';
      await deriveKey(val);
      if (statusEl) statusEl.textContent = '加密口令已设置';
    });
  }

  async function loadUploads(){
    if (!uploadsList) return;
    const r = await apiFetch('/api/uploads');
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
      // admin delete button
      if (auth.role === 'admin') {
        const del = document.createElement('button');
        del.textContent = '删除上传';
        del.style.marginLeft = '10px';
        del.addEventListener('click', async () => {
          if (!confirm(`确定删除上传目录 ${u.id} ?`)) return;
          const r2 = await apiFetch(`/api/admin/upload/${encodeURIComponent(u.id)}`, { method: 'DELETE' });
          const d2 = await r2.json();
          if (d2.ok) { await loadUploads(); }
          else { alert('删除失败'); }
        });
        li.appendChild(del);
      }
      li.appendChild(left);
      li.appendChild(btn);
      uploadsList.appendChild(li);
    });
  }

  async function fetchTextState(){
    if (!editor || !versionEl) return { version: 0, content: '' };
    const r = await apiFetch('/api/text/state');
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
  if (editor) {
    editor.addEventListener('input', () => {
      if (statusEl) statusEl.textContent = '编辑中…';
      if (typingTimer) clearTimeout(typingTimer);
      typingTimer = setTimeout(async () => {
        const encrypted = await encryptText(editor.value);
        const body = { content: encrypted, client_id: clientId };
        const r = await apiFetch('/api/text/update', {
          method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body)
        });
        const data = await r.json();
        if (versionEl) versionEl.textContent = data.version || 0;
        lastVersion = data.version || 0;
        if (statusEl) statusEl.textContent = '已同步';
        await fetchHistory();
      }, 400);
    });
  }

  async function fetchHistory(){
    if (!historyList) return;
    const r = await apiFetch(`/api/text/history?after_version=${encodeURIComponent(lastVersion || 0)}`);
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

  if (editor) {
    setInterval(async () => {
      const data = await fetchTextState();
      if ((data.version || 0) > lastVersion) {
        lastVersion = data.version || 0;
        await fetchHistory();
      }
    }, 1000);
  }

  // --- Auth helpers ---
  let auth = { authenticated: false, username: null, role: null };
  function updateAdminUI(){
    const isAdmin = auth.role === 'admin';
    adminControls && adminControls.forEach(el => {
      el.style.display = isAdmin ? 'flex' : 'none';
    });
  }
  async function refreshAuth(){
    try {
      const r = await apiFetch('/api/auth/me');
      const d = await r.json();
      auth = { authenticated: !!d.authenticated, username: d.username || null, role: d.role || null };
    } catch(e) { auth = { authenticated: false, username: null, role: null }; }
    if (authStatus) authStatus.textContent = auth.authenticated ? `${auth.username}（${auth.role || 'user'}）` : '未登录';
    const userDisplay = document.querySelector('#user-display');
    if (userDisplay) userDisplay.textContent = auth.authenticated ? `用户：${auth.username}` : '未登录';
    updateAdminUI();
    if (!auth.authenticated && isAppPage) {
      // ensure unauthenticated users stay on login
      window.location.href = '/login';
    }
  }
  btnRegister && btnRegister.addEventListener('click', async () => {
    const u = (authUsername.value || '').trim();
    const p = authPassword.value || '';
    if (!u || !p) return alert('请输入用户名与密码');
    const r = await apiFetch('/api/auth/register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username: u, password: p }) });
    if (r.ok) {
      try { await r.json(); } catch(e) {}
      setTimeout(() => { window.location.replace('/app'); }, 50);
      return;
    } else { alert('注册失败'); }
  });
  btnLogin && btnLogin.addEventListener('click', async () => {
    const u = (authUsername.value || '').trim();
    const p = authPassword.value || '';
    if (!u || !p) return alert('请输入用户名与密码');
    const r = await apiFetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username: u, password: p }) });
    if (r.ok) {
      try { await r.json(); } catch(e) {}
      setTimeout(() => { window.location.replace('/app'); }, 50);
      return;
    } else { alert('登录失败'); }
  });
  btnLogout && btnLogout.addEventListener('click', async () => {
    const r = await apiFetch('/api/auth/logout', { method: 'POST' });
    if (r.ok) { await refreshAuth(); await loadUploads(); }
  });

  // 管理用户页入口（应用页右上角）
  const btnUsers = document.querySelector('#btn-users');
  if (btnUsers) {
    btnUsers.addEventListener('click', () => {
      window.location.href = '/users';
    });
  }

  // Admin create folder
  createFolderBtn && createFolderBtn.addEventListener('click', async () => {
    const name = (newFolderName.value || '').trim();
    if (!name) return alert('请输入文件夹名称');
    const r = await apiFetch('/api/admin/folder/create', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name }) });
    const d = await r.json();
    if (d && d.ok) { newFolderName.value = ''; await loadUploads(); }
    else { alert('新建失败'); }
  });

  // Fetch wrapper to include credentials
  function apiFetch(url, opts){
    return fetch(url, Object.assign({ credentials: 'include' }, opts || {}));
  }

  function setupUsersPage(){
    const btnApp = document.querySelector('#btn-app');
    if (btnApp) btnApp.addEventListener('click', () => window.location.href = '/app');
    // 绑定创建用户
    const createBtn = document.querySelector('#user-create-btn');
    const nameInput = document.querySelector('#user-create-name');
    const passInput = document.querySelector('#user-create-pass');
    if (createBtn) {
      createBtn.addEventListener('click', async () => {
        const username = (nameInput.value || '').trim();
        const password = (passInput.value || '').trim();
        if (!username || !password) return alert('用户名与密码不能为空');
        const r = await apiFetch('/api/admin/users/create', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username, password }) });
        if (r.ok) { nameInput.value=''; passInput.value=''; await loadUsersList(); }
        else { alert('创建失败'); }
      });
    }
    // 初次加载列表
    loadUsersList();
  }

  async function loadUsersList(){
    const tbody = document.querySelector('#users-list');
    if (!tbody) return;
    const r = await apiFetch('/api/admin/users');
    const d = await r.json();
    tbody.innerHTML = '';
    (d.users || []).forEach(u => {
      const tr = document.createElement('tr');
      const td1 = document.createElement('td'); td1.style.padding='8px'; td1.textContent = u.username;
      const td2 = document.createElement('td'); td2.style.padding='8px'; td2.textContent = u.role || 'user';
      const td3 = document.createElement('td'); td3.style.padding='8px';
      if (u.role === 'admin') {
        td3.textContent = '—';
      } else {
        const del = document.createElement('button'); del.textContent = '删除'; del.className='danger';
        const np = document.createElement('input'); np.type='password'; np.placeholder='新密码'; np.style.marginLeft='8px';
        const upd = document.createElement('button'); upd.textContent='重置密码'; upd.className='ghost'; upd.style.marginLeft='8px';
        del.addEventListener('click', async ()=>{
          if (!confirm(`确认删除用户 ${u.username} ?`)) return;
          const r2 = await apiFetch('/api/admin/users/delete', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ username: u.username }) });
          if (r2.ok) { await loadUsersList(); } else { alert('删除失败'); }
        });
        upd.addEventListener('click', async ()=>{
          const newPassword = (np.value || '').trim();
          if (!newPassword) return alert('请输入新密码');
          const r3 = await apiFetch('/api/admin/users/update_password', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ username: u.username, new_password: newPassword }) });
          if (r3.ok) { np.value=''; alert('已重置'); } else { alert('重置失败'); }
        });
        td3.appendChild(del); td3.appendChild(np); td3.appendChild(upd);
      }
      tr.appendChild(td1); tr.appendChild(td2); tr.appendChild(td3);
      tbody.appendChild(tr);
    });
  }

  (async function init(){
    await loadInfo();
    await refreshAuth();
    if (isLoginPage) {
      if (auth.authenticated) { window.location.href = '/app'; return; }
      return;
    }
    if (!auth.authenticated) { return; }
    if (isUsersPage) {
      if (auth.role !== 'admin') { window.location.href = '/app'; return; }
      setupUsersPage();
      return;
    }
    if (isAppPage) {
      await loadUploads();
      const state = await fetchTextState();
      lastVersion = state.version || 0;
      await fetchHistory();
    }
  })();
})();