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
