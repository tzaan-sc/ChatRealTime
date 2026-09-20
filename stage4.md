# Hướng Dẫn Chi Tiết Giai Đoạn 4: Khởi Tạo Frontend Svelte & Giao Diện Nhắn Tin
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Stack: Vite + Svelte, Vanilla CSS (Glassmorphism, Dark Mode), WebSocket Client

Tài liệu này sẽ hướng dẫn bạn từng bước xây dựng giao diện ứng dụng chat thời gian thực bằng **Svelte**, kết nối với Backend Go đã tạo ở Giai đoạn 2 và Giai đoạn 3, quản lý trạng thái đăng nhập và hiển thị tin nhắn tức thì.

---

## Mục lục
1. [Bước 1: Khởi tạo Project Svelte với Vite](#buoc-1-khoi-tao-project-svelte)
2. [Bước 2: Xây dựng Hệ thống Giao diện & Design System (`app.css`)](#buoc-2-he-thong-giao-dien)
3. [Bước 3: Tầng Services & Svelte Stores](#buoc-3-services--stores)
4. [Bước 4: Xây dựng các Components Giao diện](#buoc-4-components-giao-dien)
5. [Bước 5: Lắp ráp Ứng dụng chính `App.svelte`](#buoc-5-lap-rap-app-svelte)
6. [Bước 6: Chạy thử nghiệm và kết nối với Backend](#buoc-6-chay-thu-nghiem)
7. [Các lỗi thường gặp và cách xử lý](#cac-loi-thuong-gap-va-cach-xu-ly)

---

## 📋 Checklist Tiến Độ Giai Đoạn 4

- [x] **1. Khởi tạo dự án Frontend bằng Vite + Svelte:**
  ```powershell
  npm create vite@latest frontend -- --template svelte
  cd frontend
  npm install
  ```
- [x] **2. Cài đặt các thư viện icon & tiện ích:**
  ```powershell
  npm install lucide-svelte
  ```
- [x] **3. Thiết lập Design System & CSS Tokens:**
  - `src/app.css` (Màu sắc HSL, Dark Mode, Glassmorphism, Animation, Custom Scrollbar)
- [x] **4. Xây dựng Stores & WebSocket Service:**
  - `src/lib/stores/auth.js` (Lưu User & JWT Token vào LocalStorage)
  - `src/lib/stores/chat.js` (Lưu danh sách hội thoại & tin nhắn đang chat)
  - `src/lib/services/websocket.js` (Tự động kết nối, tự động reconnect, bắt sự kiện)
  - `src/lib/services/api.js` (Gọi REST API: Auth, Lịch sử chat, Danh bạ)
- [x] **5. Xây dựng các Component giao diện:**
  - `src/lib/components/AuthModal.svelte` (Đăng nhập / Đăng ký)
  - `src/lib/components/Sidebar.svelte` (Danh sách hội thoại, tìm kiếm & danh bạ)
  - `src/lib/components/ChatArea.svelte` (Khung chat chính, header, danh sách tin nhắn, ô nhập liệu)
- [x] **6. Lắp ráp ứng dụng chính trong `src/App.svelte`**
- [x] **7. Khởi chạy và kiểm thử giao diện tại `http://localhost:5173`**

---

## Mục lục
1. [Bước 1: Khởi tạo Project Svelte với Vite](#buoc-1-khoi-tao-project-svelte)
2. [Bước 2: Xây dựng Hệ thống Giao diện & Design System (`app.css`)](#buoc-2-he-thong-giao-dien)
3. [Bước 3: Tầng Services & Svelte Stores](#buoc-3-services--stores)
4. [Bước 4: Xây dựng các Components Giao diện](#buoc-4-components-giao-dien)
5. [Bước 5: Lắp ráp Ứng dụng chính `App.svelte`](#buoc-5-lap-rap-app-svelte)
6. [Bước 6: Chạy thử nghiệm và kết nối với Backend](#buoc-6-chay-thu-nghiem)
7. [Các lỗi thường gặp và cách xử lý](#cac-loi-thuong-gap-va-cach-xu-ly)

---

<a name="buoc-1-khoi-tao-project-svelte"></a>
## Bước 1: Khởi tạo Project Svelte với Vite

Mở PowerShell tại thư mục gốc `d:\GIT\ChatRealTime`:

```powershell
# 1. Tạo project svelte bằng Vite
npm create vite@latest frontend -- --template svelte

# 2. Đi vào thư mục và cài dependencies
cd frontend
npm install

# 3. Cài bộ icon Lucide đẹp mắt
npm install lucide-svelte
```

---

<a name="buoc-2-he-thong-giao-dien"></a>
## Bước 2: Xây dựng Hệ thống Giao diện & Design System (`app.css`)

Tạo file `frontend/src/app.css` với giao diện Dark Mode hiện đại, hiệu ứng kính mờ (Glassmorphism) và gradient cao cấp:

```css
@import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap');

:root {
  --font-main: 'Plus Jakarta Sans', system-ui, -apple-system, sans-serif;
  
  /* Color Palette - Cyber Dark / Violet Neon */
  --bg-primary: #0b0f19;
  --bg-secondary: #111827;
  --bg-card: rgba(17, 24, 39, 0.75);
  --bg-glass: rgba(31, 41, 55, 0.6);
  --border-glass: rgba(255, 255, 255, 0.08);
  --border-focus: rgba(139, 92, 246, 0.5);

  --text-main: #f9fafb;
  --text-muted: #9ca3af;
  --text-dim: #6b7280;

  --accent-primary: #8b5cf6;
  --accent-hover: #7c3aed;
  --accent-gradient: linear-gradient(135deg, #8b5cf6 0%, #ec4899 100%);
  --bubble-me: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  --bubble-other: rgba(31, 41, 55, 0.85);

  --online-color: #10b981;
  --offline-color: #6b7280;
  --radius-sm: 8px;
  --radius-md: 14px;
  --radius-lg: 20px;
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: var(--font-main);
  background-color: var(--bg-primary);
  color: var(--text-main);
  min-height: 100vh;
  overflow: hidden;
  -webkit-font-smoothing: antialiased;
}

/* Custom Scrollbar */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 99px;
}
::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* Glassmorphism Card Utility */
.glass-card {
  background: var(--bg-card);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--border-glass);
}
```

---

<a name="buoc-3-services--stores"></a>
## Bước 3: Tầng Services & Svelte Stores

### 3.1. File `frontend/src/lib/services/api.js` (Gọi REST API)
```javascript
const API_BASE = 'http://localhost:8080/api';

export async function apiRequest(endpoint, method = 'GET', body = null, token = null) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const config = { method, headers };
  if (body) config.body = JSON.stringify(body);

  const res = await fetch(`${API_BASE}${endpoint}`, config);
  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.error || 'Có lỗi xảy ra');
  }
  return data;
}
```

### 3.2. File `frontend/src/lib/stores/auth.js` (Quản lý User & Token)
```javascript
import { writable } from 'svelte/store';
import { apiRequest } from '../services/api';

const storedToken = localStorage.getItem('token');
const storedUser = localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')) : null;

export const token = writable(storedToken);
export const currentUser = writable(storedUser);

export function loginSuccess(userVal, tokenVal) {
  localStorage.setItem('token', tokenVal);
  localStorage.setItem('user', JSON.stringify(userVal));
  token.set(tokenVal);
  currentUser.set(userVal);
}

export function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  token.set(null);
  currentUser.set(null);
}
```

### 3.3. File `frontend/src/lib/services/websocket.js` (WebSocket Client)
```javascript
import { writable } from 'svelte/store';

export const wsStatus = writable('DISCONNECTED'); // CONNECTED, CONNECTING, DISCONNECTED
export const incomingMessages = writable(null);

class WSClient {
  constructor() {
    this.ws = null;
    this.token = null;
    this.reconnectTimer = null;
  }

  connect(token) {
    if (!token) return;
    this.token = token;
    wsStatus.set('CONNECTING');

    const wsUrl = `ws://localhost:8080/ws?token=${token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      wsStatus.set('CONNECTED');
      console.log('🟢 WebSocket Connected');
      clearTimeout(this.reconnectTimer);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        incomingMessages.set(data);
      } catch (e) {
        console.error('Lỗi parse gói tin WS:', e);
      }
    };

    this.ws.onclose = () => {
      wsStatus.set('DISCONNECTED');
      console.log('🔴 WebSocket Disconnected, reconnecting in 3s...');
      this.reconnectTimer = setTimeout(() => this.connect(this.token), 3000);
    };

    this.ws.onerror = (err) => {
      console.error('WebSocket Error:', err);
      this.ws.close();
    };
  }

  send(event, payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ event, payload }));
    } else {
      console.warn('WebSocket chưa sẵn sàng để gửi tin');
    }
  }

  disconnect() {
    clearTimeout(this.reconnectTimer);
    if (this.ws) this.ws.close();
  }
}

export const wsService = new WSClient();
```

### 3.4. File `frontend/src/lib/stores/chat.js` (Hội thoại & Lịch sử tin nhắn)
```javascript
import { writable, get } from 'svelte/store';
import { apiRequest } from '../services/api';
import { token } from './auth';

export const conversations = writable([]);
export const activeConversation = writable(null); // { conversation, other_user }
export const currentMessages = writable([]);
export const userDirectory = writable([]);

// Tải danh sách hội thoại
export async function loadConversations() {
  const t = get(token);
  if (!t) return;
  try {
    const res = await apiRequest('/chat/conversations', 'GET', null, t);
    conversations.set(res.data || []);
  } catch (err) {
    console.error('Lỗi tải danh sách hội thoại:', err);
  }
}

// Tải lịch sử tin nhắn của cuộc trò chuyện được chọn
export async function selectConversation(convItem) {
  activeConversation.set(convItem);
  const t = get(token);
  const convID = convItem.conversation.custom_id;
  try {
    const res = await apiRequest(`/chat/messages/${convID}?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
    // Đánh dấu đã xem
    apiRequest(`/chat/messages/${convID}/read`, 'POST', null, t);
  } catch (err) {
    console.error('Lỗi tải tin nhắn:', err);
  }
}

// Thêm tin nhắn mới vào danh sách hiện tại
export function appendMessage(msg) {
  currentMessages.update((msgs) => [...msgs, msg]);
  loadConversations(); // Cập nhật lại snippet tin nhắn cuối
}
```

---

<a name="buoc-4-components-giao-dien"></a>
## Bước 4: Xây dựng các Components Giao diện

### 4.1. File `frontend/src/lib/components/AuthModal.svelte` (Đăng nhập / Đăng ký)
```svelte
<script>
  import { apiRequest } from '../services/api';
  import { loginSuccess } from '../stores/auth';
  import { loadConversations } from '../stores/chat';

  let isLogin = true;
  let username = '';
  let email = '';
  let password = '';
  let displayName = '';
  let errorMsg = '';
  let loading = false;

  async function handleSubmit() {
    errorMsg = '';
    loading = true;
    try {
      if (isLogin) {
        const res = await apiRequest('/auth/login', 'POST', { username, password });
        loginSuccess(res.data.user, res.data.token);
      } else {
        const res = await apiRequest('/auth/register', 'POST', {
          username,
          email,
          password,
          display_name: displayName || username
        });
        loginSuccess(res.data.user, res.data.token);
      }
      loadConversations();
    } catch (err) {
      errorMsg = err.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="auth-overlay">
  <div class="auth-card glass-card">
    <div class="brand">
      <div class="logo-icon">⚡</div>
      <h2>{isLogin ? 'Đăng Nhập Chat' : 'Tạo Tài Khoản'}</h2>
      <p class="subtitle">Trải nghiệm nhắn tin thời gian thực cực mượt</p>
    </div>

    {#if errorMsg}
      <div class="error-badge">{errorMsg}</div>
    {/if}

    <form on:submit|preventDefault={handleSubmit}>
      <div class="input-group">
        <label>Tên đăng nhập</label>
        <input type="text" bind:value={username} placeholder="vd: alex99" required />
      </div>

      {#if !isLogin}
        <div class="input-group">
          <label>Email</label>
          <input type="email" bind:value={email} placeholder="alex@gmail.com" required />
        </div>
        <div class="input-group">
          <label>Tên hiển thị</label>
          <input type="text" bind:value={displayName} placeholder="Alex Nguyen" />
        </div>
      {/if}

      <div class="input-group">
        <label>Mật khẩu</label>
        <input type="password" bind:value={password} placeholder="••••••••" required />
      </div>

      <button type="submit" class="btn-primary" disabled={loading}>
        {loading ? 'Đang xử lý...' : (isLogin ? 'Đăng Nhập' : 'Đăng Ký')}
      </button>
    </form>

    <div class="switch-mode">
      <span>{isLogin ? 'Chưa có tài khoản?' : 'Đã có tài khoản?'}</span>
      <button type="button" on:click={() => { isLogin = !isLogin; errorMsg = ''; }}>
        {isLogin ? 'Đăng ký ngay' : 'Đăng nhập'}
      </button>
    </div>
  </div>
</div>

<style>
  .auth-overlay {
    position: fixed;
    inset: 0;
    background: rgba(11, 15, 25, 0.85);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 999;
    padding: 20px;
  }
  .auth-card {
    width: 100%;
    max-width: 420px;
    border-radius: var(--radius-lg);
    padding: 36px 32px;
  }
  .brand { text-align: center; margin-bottom: 24px; }
  .logo-icon { font-size: 40px; margin-bottom: 8px; }
  .brand h2 { font-size: 24px; font-weight: 700; color: #fff; }
  .subtitle { font-size: 13px; color: var(--text-muted); margin-top: 4px; }
  .error-badge {
    background: rgba(239, 68, 68, 0.15);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.3);
    padding: 10px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    margin-bottom: 16px;
    text-align: center;
  }
  .input-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
  .input-group label { font-size: 13px; font-weight: 500; color: var(--text-muted); }
  .input-group input {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 12px 14px;
    color: #fff;
    outline: none;
    transition: 0.2s;
  }
  .input-group input:focus {
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 2px var(--border-focus);
  }
  .btn-primary {
    width: 100%;
    padding: 12px;
    background: var(--accent-gradient);
    border: none;
    border-radius: var(--radius-sm);
    color: #fff;
    font-weight: 600;
    cursor: pointer;
    margin-top: 8px;
    transition: 0.2s;
  }
  .btn-primary:hover { opacity: 0.95; transform: translateY(-1px); }
  .switch-mode {
    display: flex;
    justify-content: center;
    gap: 8px;
    margin-top: 20px;
    font-size: 13px;
    color: var(--text-muted);
  }
  .switch-mode button {
    background: none;
    border: none;
    color: var(--accent-primary);
    font-weight: 600;
    cursor: pointer;
  }
</style>
```

### 4.2. File `frontend/src/lib/components/Sidebar.svelte` (Danh sách Chat & Danh bạ)
```svelte
<script>
  import { conversations, activeConversation, selectConversation, userDirectory } from '../stores/chat';
  import { currentUser, logout, token } from '../stores/auth';
  import { apiRequest } from '../services/api';
  import { onMount } from 'svelte';
  import { MessageSquarePlus, LogOut, Search } from 'lucide-svelte';

  let showUsersModal = false;
  let searchQuery = '';

  async function openNewChatModal() {
    showUsersModal = true;
    try {
      const res = await apiRequest('/chat/users', 'GET', null, $token);
      userDirectory.set(res.data || []);
    } catch (e) {
      console.error(e);
    }
  }

  function startChatWithUser(user) {
    const customID = [$currentUser.id, user.id].sort().join('_');
    const newConv = {
      conversation: { custom_id: customID, members: [$currentUser.id, user.id] },
      other_user: user,
      unread_count: 0
    };
    selectConversation(newConv);
    showUsersModal = false;
  }
</script>

<aside class="sidebar glass-card">
  <!-- User Profile Header -->
  <div class="sidebar-header">
    <div class="user-info">
      <img src={$currentUser?.avatar_url} alt="avatar" class="my-avatar" />
      <div>
        <h4>{$currentUser?.display_name || $currentUser?.username}</h4>
        <span class="online-tag">● Trực tuyến</span>
      </div>
    </div>
    <div class="header-actions">
      <button class="icon-btn" title="Cuộc trò chuyện mới" on:click={openNewChatModal}>
        <MessageSquarePlus size={20} />
      </button>
      <button class="icon-btn" title="Đăng xuất" on:click={logout}>
        <LogOut size={20} />
      </button>
    </div>
  </div>

  <!-- Search Bar -->
  <div class="search-box">
    <Search size={16} class="search-icon" />
    <input type="text" placeholder="Tìm kiếm cuộc trò chuyện..." bind:value={searchQuery} />
  </div>

  <!-- Conversation List -->
  <div class="conversation-list">
    {#if $conversations.length === 0}
      <div class="empty-state">Chưa có cuộc trò chuyện nào. Bấm nút dấu cộng để bắt đầu nhắn tin!</div>
    {:else}
      {#each $conversations as item}
        <div
          class="conversation-item"
          class:active={$activeConversation?.conversation?.custom_id === item.conversation.custom_id}
          on:click={() => selectConversation(item)}
        >
          <img src={item.other_user.avatar_url} alt="avatar" class="avatar" />
          <div class="content">
            <div class="top-line">
              <span class="name">{item.other_user.display_name || item.other_user.username}</span>
              <span class="time">
                {new Date(item.conversation.last_message_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
            </div>
            <div class="bottom-line">
              <p class="snippet">{item.conversation.last_message || 'Bắt đầu cuộc trò chuyện'}</p>
              {#if item.unread_count > 0}
                <span class="unread-badge">{item.unread_count}</span>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</aside>

<!-- Modal Chọn Bạn Chat Mới -->
{#if showUsersModal}
  <div class="modal-overlay" on:click={() => (showUsersModal = false)}>
    <div class="modal-box glass-card" on:click|stopPropagation>
      <h3>Bắt đầu cuộc trò chuyện</h3>
      <div class="user-list">
        {#each $userDirectory as u}
          <div class="user-item" on:click={() => startChatWithUser(u)}>
            <img src={u.avatar_url} alt="avatar" class="avatar" />
            <div>
              <p class="u-name">{u.display_name}</p>
              <p class="u-sub">@{u.username}</p>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}

<style>
  .sidebar {
    width: 360px;
    height: 100vh;
    border-right: 1px solid var(--border-glass);
    display: flex;
    flex-direction: column;
  }
  .sidebar-header {
    padding: 18px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-glass);
  }
  .user-info { display: flex; align-items: center; gap: 12px; }
  .my-avatar { width: 44px; height: 44px; border-radius: 50%; border: 2px solid var(--accent-primary); }
  .user-info h4 { font-size: 15px; font-weight: 600; color: #fff; }
  .online-tag { font-size: 11px; color: var(--online-color); }
  .header-actions { display: flex; gap: 8px; }
  .icon-btn {
    background: rgba(255, 255, 255, 0.06);
    border: none;
    color: var(--text-muted);
    width: 36px;
    height: 36px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .icon-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.12); }
  .search-box {
    margin: 14px 20px;
    position: relative;
    display: flex;
    align-items: center;
  }
  :global(.search-icon) {
    position: absolute;
    left: 12px;
    color: var(--text-dim);
  }
  .search-box input {
    width: 100%;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    padding: 10px 14px 10px 36px;
    color: #fff;
    font-size: 13px;
    outline: none;
  }
  .conversation-list { flex: 1; overflow-y: auto; padding: 0 10px 20px; }
  .conversation-item {
    display: flex;
    gap: 12px;
    padding: 12px 14px;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: 0.2s;
    margin-bottom: 4px;
  }
  .conversation-item:hover { background: rgba(255, 255, 255, 0.05); }
  .conversation-item.active { background: rgba(139, 92, 246, 0.15); border: 1px solid rgba(139, 92, 246, 0.3); }
  .avatar { width: 46px; height: 46px; border-radius: 50%; }
  .content { flex: 1; min-width: 0; }
  .top-line { display: flex; justify-content: space-between; margin-bottom: 4px; }
  .name { font-size: 14px; font-weight: 600; color: #fff; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .time { font-size: 11px; color: var(--text-dim); }
  .bottom-line { display: flex; justify-content: space-between; align-items: center; }
  .snippet { font-size: 12px; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 200px; }
  .unread-badge {
    background: var(--accent-primary);
    color: #fff;
    font-size: 11px;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 99px;
  }
  .empty-state { padding: 30px 20px; text-align: center; color: var(--text-dim); font-size: 13px; }
  .modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.7); display: flex; align-items: center; justify-content: center; z-index: 1000;
  }
  .modal-box { width: 360px; padding: 24px; border-radius: var(--radius-lg); }
  .modal-box h3 { font-size: 16px; margin-bottom: 16px; }
  .user-list { max-height: 300px; overflow-y: auto; }
  .user-item { display: flex; gap: 12px; padding: 10px; border-radius: var(--radius-sm); cursor: pointer; }
  .user-item:hover { background: rgba(255, 255, 255, 0.08); }
  .u-name { font-size: 14px; font-weight: 600; }
  .u-sub { font-size: 12px; color: var(--text-dim); }
</style>
```

### 4.3. File `frontend/src/lib/components/ChatArea.svelte` (Khung chat chính)
```svelte
<script>
  import { activeConversation, currentMessages, appendMessage } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import { wsService } from '../services/websocket';
  import { onMount, afterUpdate } from 'svelte';
  import { Send, Smile, Paperclip } from 'lucide-svelte';

  let inputContent = '';
  let messagesContainer;

  function scrollToBottom() {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  }

  afterUpdate(() => {
    scrollToBottom();
  });

  function handleSend() {
    if (!inputContent.trim() || !$activeConversation) return;

    const receiverID = $activeConversation.other_user.id;
    wsService.send('chat:send', {
      receiver_id: receiverID,
      content: inputContent.trim(),
      type: 'text'
    });

    inputContent = '';
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }
</script>

{#if !$activeConversation}
  <div class="empty-chat glass-card">
    <div class="empty-content">
      <div class="icon-circle">💬</div>
      <h2>Chào mừng đến với Realtime Chat!</h2>
      <p>Chọn một người bạn ở danh sách bên trái để bắt đầu cuộc trò chuyện.</p>
    </div>
  </div>
{:else}
  <main class="chat-area">
    <!-- Header -->
    <div class="chat-header glass-card">
      <div class="partner-info">
        <img src={$activeConversation.other_user.avatar_url} alt="avatar" class="header-avatar" />
        <div>
          <h3>{$activeConversation.other_user.display_name || $activeConversation.other_user.username}</h3>
          <span class="status-sub">Sẵn sàng nhận tin</span>
        </div>
      </div>
    </div>

    <!-- Messages Body -->
    <div class="messages-viewport" bind:this={messagesContainer}>
      {#each $currentMessages as msg}
        {@const isMe = msg.sender_id === $currentUser.id}
        <div class="message-row" class:me={isMe}>
          <div class="bubble" class:bubble-me={isMe} class:bubble-other={!isMe}>
            <p class="text">{msg.content}</p>
            <span class="timestamp">
              {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          </div>
        </div>
      {/each}
    </div>

    <!-- Input Footer -->
    <div class="chat-footer glass-card">
      <div class="input-wrapper">
        <textarea
          rows="1"
          placeholder="Nhập tin nhắn... (Nhấn Enter để gửi)"
          bind:value={inputContent}
          on:keydown={handleKeyDown}
        ></textarea>
        <button class="send-btn" on:click={handleSend} disabled={!inputContent.trim()}>
          <Send size={18} />
        </button>
      </div>
    </div>
  </main>
{/if}

<style>
  .chat-area { flex: 1; height: 100vh; display: flex; flex-direction: column; background: var(--bg-primary); }
  .empty-chat { flex: 1; display: flex; align-items: center; justify-content: center; text-align: center; }
  .icon-circle { font-size: 50px; margin-bottom: 16px; }
  .empty-content h2 { font-size: 22px; margin-bottom: 8px; }
  .empty-content p { color: var(--text-muted); font-size: 14px; }
  .chat-header {
    padding: 16px 24px;
    border-bottom: 1px solid var(--border-glass);
    display: flex;
    align-items: center;
  }
  .partner-info { display: flex; align-items: center; gap: 14px; }
  .header-avatar { width: 44px; height: 44px; border-radius: 50%; }
  .partner-info h3 { font-size: 16px; font-weight: 600; color: #fff; }
  .status-sub { font-size: 12px; color: var(--text-dim); }
  .messages-viewport {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .message-row { display: flex; width: 100%; }
  .message-row.me { justify-content: flex-end; }
  .bubble {
    max-width: 65%;
    padding: 12px 16px;
    border-radius: var(--radius-md);
    position: relative;
    word-break: break-word;
  }
  .bubble-me {
    background: var(--bubble-me);
    color: #fff;
    border-bottom-right-radius: 4px;
  }
  .bubble-other {
    background: var(--bubble-other);
    color: #f3f4f6;
    border-bottom-left-radius: 4px;
    border: 1px solid var(--border-glass);
  }
  .text { font-size: 14px; line-height: 1.5; }
  .timestamp {
    display: block;
    font-size: 10px;
    opacity: 0.7;
    margin-top: 4px;
    text-align: right;
  }
  .chat-footer { padding: 16px 24px; border-top: 1px solid var(--border-glass); }
  .input-wrapper {
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 8px 16px;
  }
  .input-wrapper textarea {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    font-size: 14px;
    outline: none;
    resize: none;
    font-family: inherit;
    max-height: 100px;
  }
  .send-btn {
    background: var(--accent-gradient);
    border: none;
    color: #fff;
    width: 38px;
    height: 38px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .send-btn:not(:disabled):hover { transform: scale(1.05); }
</style>
```

---

<a name="buoc-5-lap-rap-app-svelte"></a>
## Bước 5: Lắp ráp Ứng dụng chính `App.svelte`

Cập nhật file `frontend/src/App.svelte`:

```svelte
<script>
  import { onMount } from 'svelte';
  import { token, currentUser } from './lib/stores/auth';
  import { wsService, incomingMessages } from './lib/services/websocket';
  import { loadConversations, appendMessage } from './lib/stores/chat';
  import AuthModal from './lib/components/AuthModal.svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import ChatArea from './lib/components/ChatArea.svelte';

  // Tự động kết nối WebSocket và nạp dữ liệu khi mở trang
  $: if ($token) {
    wsService.connect($token);
    loadConversations();
  }

  // Lắng nghe sự kiện từ WebSocket
  $: if ($incomingMessages) {
    const { event, payload } = $incomingMessages;
    if (event === 'chat:receive' || event === 'chat:ack') {
      const msg = event === 'chat:receive' ? payload : payload.message;
      appendMessage(msg);
    }
  }
</script>

<div class="app-layout">
  {#if !$currentUser}
    <AuthModal />
  {:else}
    <Sidebar />
    <ChatArea />
  {/if}
</div>

<style>
  .app-layout {
    display: flex;
    width: 100vw;
    height: 100vh;
    background: var(--bg-primary);
  }
</style>
```
### Cách hoạt động của `App.svelte` này như sau:

1. Kiểm tra trạng thái đăng nhập:
 - Nếu chưa đăng nhập, hiển thị `AuthModal` (cửa sổ Đăng nhập/Đăng ký).
 - Nếu đã đăng nhập, hiển thị `Sidebar` (cột danh sách bạn bè) và `ChatArea` (khung chat chính).

2. Kết nối WebSocket:
 - Khi có token, tự động kết nối WebSocket tới `ws://localhost:8080/ws`.
 - Lắng nghe sự kiện `chat:receive` để nhận tin nhắn real-time.

3. Hiển thị dữ liệu:
 - Nạp danh sách hội thoại từ REST API.
 - Hiển thị danh sách bạn bè bên trái và khung chat chính bên phải.

---

<a name="buoc-6-chay-thu-nghiem"></a>
## Bước 6: Chạy thử nghiệm và kết nối với Backend

1. **Khởi động Backend Go (nếu chưa bật):**
   ```powershell
   cd d:\GIT\ChatRealTime\backend
   go run cmd/server/main.go
   ```

2. **Khởi động Frontend Svelte:**
   Mở một cửa sổ Terminal mới:
   ```powershell
   cd d:\GIT\ChatRealTime\frontend
   npm run dev
   ```

3. **Mở 2 cửa sổ trình duyệt để kiểm thử:**
   - Cửa sổ 1: Mở `http://localhost:5173` $\to$ Đăng ký tài khoản `alex`.
   - Cửa sổ 2 (Ẩn danh / Incognito): Mở `http://localhost:5173` $\to$ Đăng ký tài khoản `bob`.
   - Bấm dấu `+` để tìm và gửi tin nhắn qua lại. Bạn sẽ thấy tin nhắn nhảy sang nhau **ngay lập tức trong tích tắc**!

---

<a name="cac-loi-thuong-gap-va-cach-xu-ly"></a>
## Các lỗi thường gặp và cách xử lý

1. **Lỗi `Failed to fetch` (Lỗi CORS):**
   - Đảm bảo Backend Go đã có dòng `r.Use(cors.Default())` trong `main.go`.
2. **Lỗi `WebSocket connection failed`:**
   - Kiểm tra xem token JWT có được truyền đúng qua URL `ws://localhost:8080/ws?token=...` không.
