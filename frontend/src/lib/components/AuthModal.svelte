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
  let successMsg = '';
  let loading = false;

  // View switch: 'auth' | 'forgot'
  let authView = 'auth';
  let forgotEmail = '';
  let forgotOtp = '';
  let forgotNewPassword = '';
  let forgotConfirmPassword = '';
  let forgotStep = 1; // 1: Nhập email, 2: Nhập OTP & đổi pass
  let demoForgotOtp = '';

  // Trạng thái Social OAuth Modal
  let activeSocialProvider = null; // 'google', 'facebook', 'github', 'apple', 'discord'
  let socialEmail = '';
  let socialName = '';
  let socialAvatar = '';
  let socialToken = '';

  // Tài khoản mẫu demo cho từng provider để test nhanh 1-chạm
  const socialPresets = {
    google: [
      { name: 'Alex Google (Dev)', email: 'alex.google@gmail.com', avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&auto=format&fit=crop&q=80' },
      { name: 'Michael Google (Tech)', email: 'michael.g@gmail.com', avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=100&auto=format&fit=crop&q=80' }
    ],
    facebook: [
      { name: 'Sarah Facebook (Design)', email: 'sarah.meta@facebook.com', avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&auto=format&fit=crop&q=80' },
      { name: 'David Facebook (Marketing)', email: 'david.fb@meta.com', avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=100&auto=format&fit=crop&q=80' }
    ],
    github: [
      { name: 'Linus Octocat', email: 'octocat.linux@github.com', avatar: 'https://avatars.githubusercontent.com/u/583231?v=4' },
      { name: 'Elena Codecraft', email: 'elena.dev@github.com', avatar: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?w=100&auto=format&fit=crop&q=80' }
    ],
    apple: [
      { name: 'Apple VIP User', email: 'steve.apple@icloud.com', avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&auto=format&fit=crop&q=80' }
    ],
    discord: [
      { name: 'Gamer Wumpus', email: 'wumpus.chat@discord.gg', avatar: 'https://images.unsplash.com/photo-1566492031773-4f4e44671857?w=100&auto=format&fit=crop&q=80' }
    ]
  };

  async function handleSubmit() {
    errorMsg = '';
    successMsg = '';
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

  function openSocialModal(provider) {
    activeSocialProvider = provider;
    errorMsg = '';
    successMsg = '';
    // Mặc định chọn preset đầu tiên
    const presets = socialPresets[provider] || [];
    if (presets.length > 0) {
      socialName = presets[0].name;
      socialEmail = presets[0].email;
      socialAvatar = presets[0].avatar;
    } else {
      socialName = '';
      socialEmail = '';
      socialAvatar = '';
    }
    socialToken = '';
  }

  function selectPreset(preset) {
    socialName = preset.name;
    socialEmail = preset.email;
    socialAvatar = preset.avatar;
  }

  async function handleSocialLogin() {
    if (!socialEmail && !socialToken) {
      errorMsg = 'Vui lòng chọn hoặc nhập Email liên kết';
      return;
    }
    loading = true;
    errorMsg = '';
    try {
      const payload = {
        provider: activeSocialProvider,
        email: socialEmail,
        name: socialName,
        avatar_url: socialAvatar,
        provider_id: `${activeSocialProvider}_${socialEmail.replace(/[@.]/g, '_')}`,
        token: socialToken
      };

      const res = await apiRequest('/auth/oauth', 'POST', payload);
      activeSocialProvider = null;
      loginSuccess(res.data.user, res.data.token);
      loadConversations();
    } catch (err) {
      errorMsg = err.message;
    } finally {
      loading = false;
    }
  }

  async function handleSendForgotOtp() {
    if (!forgotEmail) {
      errorMsg = 'Vui lòng nhập địa chỉ email của bạn';
      return;
    }
    loading = true;
    errorMsg = '';
    successMsg = '';
    try {
      const res = await apiRequest('/auth/forgot-password', 'POST', { email: forgotEmail });
      demoForgotOtp = res.demo_otp || '';
      forgotStep = 2;
      successMsg = res.message || 'Mã OTP đặt lại mật khẩu đã được gửi!';
    } catch (err) {
      errorMsg = err.message;
    } finally {
      loading = false;
    }
  }

  async function handleResetPassword() {
    if (!forgotOtp) {
      errorMsg = 'Vui lòng nhập mã OTP 6 chữ số';
      return;
    }
    if (!forgotNewPassword || forgotNewPassword.length < 6) {
      errorMsg = 'Mật khẩu mới phải có tối thiểu 6 ký tự';
      return;
    }
    if (forgotNewPassword !== forgotConfirmPassword) {
      errorMsg = 'Mật khẩu xác nhận không khớp!';
      return;
    }

    loading = true;
    errorMsg = '';
    successMsg = '';
    try {
      const res = await apiRequest('/auth/reset-password', 'POST', {
        email: forgotEmail,
        otp: forgotOtp,
        new_password: forgotNewPassword
      });
      authView = 'auth';
      isLogin = true;
      successMsg = res.message || 'Đặt lại mật khẩu thành công! Hãy đăng nhập ngay.';
      forgotStep = 1;
      forgotOtp = '';
      forgotNewPassword = '';
      forgotConfirmPassword = '';
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
    {#if successMsg}
      <div class="success-badge">{successMsg}</div>
    {/if}

    {#if authView === 'auth'}
      <!-- Social Sign-in Buttons -->
      <div class="social-auth-section">
        <div class="social-btn-grid">
          <!-- Google -->
          <button
            type="button"
            class="social-btn btn-google"
            title="Đăng nhập bằng tài khoản Google"
            on:click={() => openSocialModal('google')}
          >
            <svg class="social-icon" viewBox="0 0 24 24">
              <path fill="#EA4335" d="M12 5c1.6 0 3 .6 4.1 1.7l3.1-3.1C17.3 1.8 14.8 1 12 1 7.5 1 3.7 3.6 1.9 7.3l3.7 2.9C6.5 7.4 9 5 12 5z"/>
              <path fill="#4285F4" d="M23.5 12.3c0-.8-.1-1.6-.2-2.3H12v4.6h6.5c-.3 1.5-1.1 2.8-2.4 3.7l3.7 2.9c2.2-2 3.7-5 3.7-8.9z"/>
              <path fill="#FBBC05" d="M5.6 14.8c-.2-.7-.4-1.5-.4-2.8s.2-2.1.4-2.8L1.9 6.3C.7 8.7 0 10.3 0 12s.7 3.3 1.9 5.7l3.7-2.9z"/>
              <path fill="#34A853" d="M12 23c3.2 0 6-1.1 8-3l-3.7-2.9c-1.1.7-2.5 1.2-4.3 1.2-3 0-5.5-2.4-6.4-5.2L1.9 16c1.8 3.7 5.6 7 10.1 7z"/>
            </svg>
            <span>Google</span>
          </button>

          <!-- Facebook -->
          <button
            type="button"
            class="social-btn btn-facebook"
            title="Đăng nhập bằng Facebook"
            on:click={() => openSocialModal('facebook')}
          >
            <svg class="social-icon" viewBox="0 0 24 24" fill="#1877F2">
              <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
            </svg>
            <span>Facebook</span>
          </button>

          <!-- GitHub -->
          <button
            type="button"
            class="social-btn btn-github"
            title="Đăng nhập bằng GitHub"
            on:click={() => openSocialModal('github')}
          >
            <svg class="social-icon" viewBox="0 0 24 24" fill="#fff">
              <path fill-rule="evenodd" clip-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
            </svg>
            <span>GitHub</span>
          </button>
        </div>

        <!-- More providers: Apple & Discord -->
        <div class="social-mini-row">
          <button
            type="button"
            class="social-mini-btn"
            title="Đăng nhập bằng Apple ID"
            on:click={() => openSocialModal('apple')}
          >
            <svg class="mini-icon" viewBox="0 0 170 170" fill="#fff">
              <path d="M150.37 130.25c-2.45 5.66-5.35 10.87-8.71 15.66-4.58 6.53-8.33 11.05-11.22 13.56-4.48 4.12-9.28 6.23-14.42 6.35-3.69 0-8.14-1.05-13.32-3.18-5.19-2.12-9.97-3.17-14.34-3.17-4.58 0-9.49 1.05-14.75 3.17-5.26 2.13-9.5 3.24-12.74 3.35-4.35.13-9.16-1.9-14.42-6.08-3.69-3.04-7.6-7.85-11.75-14.42-6.53-10.43-11.59-21.75-15.18-33.95-3.59-12.2-5.39-23.71-5.39-34.54 0-15.88 4.14-28.77 12.42-38.67 8.28-9.9 18.23-14.93 29.86-15.08 4.58 0 9.87 1.25 15.86 3.75 6 2.5 10.15 3.79 12.44 3.86 1.74 0 6.07-1.39 12.99-4.17 6.92-2.78 12.78-4.04 17.58-3.79 13.29.87 23.83 5.43 31.62 13.68-11.55 6.97-17.21 16.7-16.98 29.21.23 9.79 4.04 17.9 11.44 24.32 7.4 6.42 16.32 10.01 26.76 10.77-2.39 7.4-5.31 15.02-8.77 22.86zM119.22 31.84c0-7.39 2.67-14.34 8.01-20.85 5.34-6.51 12.02-10.66 20.04-12.45.22 1.09.33 2.18.33 3.27 0 7.18-2.73 14.19-8.19 21.03-5.46 6.84-12.27 10.98-20.44 12.42-.11-1.09-.17-2.18-.17-3.42z"/>
            </svg>
            <span>Apple</span>
          </button>
          <button
            type="button"
            class="social-mini-btn"
            title="Đăng nhập bằng Discord"
            on:click={() => openSocialModal('discord')}
          >
            <svg class="mini-icon" viewBox="0 0 24 24" fill="#5865F2">
              <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994.021-.041.001-.09-.041-.106a13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.929 1.793 8.18 1.793 12.061 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.893.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.028zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
            </svg>
            <span>Discord</span>
          </button>
        </div>
      </div>

      <!-- Hoặc qua Username & Password -->
      <div class="divider">
        <div class="line"></div>
        <span class="divider-text">hoặc tài khoản mật khẩu</span>
        <div class="line"></div>
      </div>

      <form on:submit|preventDefault={handleSubmit}>
        <div class="input-group">
          <label for="auth-username">Tên đăng nhập</label>
          <input id="auth-username" type="text" bind:value={username} placeholder="vd: alex99" required />
        </div>

        {#if !isLogin}
          <div class="input-group">
            <label for="auth-email">Email</label>
            <input id="auth-email" type="email" bind:value={email} placeholder="alex@gmail.com" required />
          </div>
          <div class="input-group">
            <label for="auth-display-name">Tên hiển thị</label>
            <input id="auth-display-name" type="text" bind:value={displayName} placeholder="Alex Nguyen" />
          </div>
        {/if}

        <div class="input-group">
          <div class="label-row">
            <label for="auth-password">Mật khẩu</label>
            {#if isLogin}
              <button
                type="button"
                class="forgot-link"
                on:click={() => { authView = 'forgot'; errorMsg = ''; successMsg = ''; }}
              >
                Quên mật khẩu?
              </button>
            {/if}
          </div>
          <input id="auth-password" type="password" bind:value={password} placeholder="••••••••" required />
        </div>

        <button type="submit" class="btn-primary" disabled={loading}>
          {loading ? 'Đang xử lý...' : (isLogin ? 'Đăng Nhập' : 'Đăng Ký')}
        </button>
      </form>

      <div class="switch-mode">
        <span>{isLogin ? 'Chưa có tài khoản?' : 'Đã có tài khoản?'}</span>
        <button type="button" on:click={() => { isLogin = !isLogin; errorMsg = ''; successMsg = ''; }}>
          {isLogin ? 'Đăng ký ngay' : 'Đăng nhập'}
        </button>
      </div>
    {:else}
      <!-- MÀN HÌNH QUÊN MẬT KHẨU / ĐẶT LẠI MẬT KHẨU QUA EMAIL OTP -->
      <div class="forgot-container">
        <div class="forgot-header-info">
          <h3>Khôi Phục Mật Khẩu</h3>
          <p class="subtitle">
            {forgotStep === 1
              ? 'Nhập địa chỉ email đăng ký để nhận mã OTP xác minh'
              : `Nhập mã OTP gửi tới ${forgotEmail} và thiết lập mật khẩu mới`}
          </p>
        </div>

        {#if demoForgotOtp}
          <div class="demo-otp-box">
            <span>✨ Mã OTP mô phỏng: <strong>{demoForgotOtp}</strong></span>
            <button type="button" class="btn-copy-otp" on:click={() => { forgotOtp = demoForgotOtp; }}>
              Tự điền
            </button>
          </div>
        {/if}

        {#if forgotStep === 1}
          <form on:submit|preventDefault={handleSendForgotOtp}>
            <div class="input-group">
              <label for="forgot-email">Địa chỉ Email</label>
              <input
                id="forgot-email"
                type="email"
                bind:value={forgotEmail}
                placeholder="vd: alex@gmail.com"
                required
              />
            </div>
            <button type="submit" class="btn-primary" disabled={loading}>
              {loading ? 'Đang gửi mã...' : 'Gửi mã OTP qua Email'}
            </button>
          </form>
        {:else}
          <form on:submit|preventDefault={handleResetPassword}>
            <div class="input-group">
              <label for="forgot-otp">Mã OTP (6 chữ số)</label>
              <input
                id="forgot-otp"
                type="text"
                maxlength="6"
                bind:value={forgotOtp}
                placeholder="123456"
                required
                style="letter-spacing: 0.25em; font-size: 18px; font-weight: 700; text-align: center;"
              />
            </div>
            <div class="input-group">
              <label for="forgot-new-pw">Mật khẩu mới (tối thiểu 6 ký tự)</label>
              <input
                id="forgot-new-pw"
                type="password"
                bind:value={forgotNewPassword}
                placeholder="••••••••"
                required
              />
            </div>
            <div class="input-group">
              <label for="forgot-confirm-pw">Xác nhận mật khẩu mới</label>
              <input
                id="forgot-confirm-pw"
                type="password"
                bind:value={forgotConfirmPassword}
                placeholder="••••••••"
                required
              />
            </div>
            <button type="submit" class="btn-primary" disabled={loading}>
              {loading ? 'Đang cập nhật...' : 'Xác nhận Đặt lại Mật khẩu'}
            </button>
            <button
              type="button"
              class="btn-resend"
              disabled={loading}
              on:click={handleSendForgotOtp}
            >
              Gửi lại mã OTP
            </button>
          </form>
        {/if}

        <div class="back-to-login">
          <button type="button" on:click={() => { authView = 'auth'; errorMsg = ''; successMsg = ''; forgotStep = 1; }}>
            ← Quay lại màn hình Đăng nhập
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Modal Giả lập / Xác thực OAuth mạng xã hội -->
{#if activeSocialProvider}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-overlay" on:click={() => (activeSocialProvider = null)}>
    <div class="social-modal glass-card" on:click={(e) => e.stopPropagation()}>
      <div class="social-modal-header">
        <div class="provider-badge {activeSocialProvider}">
          <span class="provider-title">Kết nối {activeSocialProvider.toUpperCase()} OAuth</span>
        </div>
        <button class="close-x" on:click={() => (activeSocialProvider = null)}>✕</button>
      </div>

      <p class="social-hint">
        Hệ thống hỗ trợ xác thực tài khoản 1-chạm. Bạn có thể chọn tài khoản mẫu bên dưới hoặc nhập thông tin cá nhân để đăng nhập ngay:
      </p>

      <!-- Danh sách tài khoản mẫu test 1-chạm -->
      <div class="presets-section">
        <span class="section-tag">Tài khoản đề xuất nhanh:</span>
        <div class="presets-grid">
          {#each (socialPresets[activeSocialProvider] || []) as preset}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <div
              class="preset-item"
              class:selected={socialEmail === preset.email}
              role="button"
              tabindex="0"
              on:click={() => selectPreset(preset)}
            >
              <img src={preset.avatar} alt="avatar" class="preset-avatar" />
              <div class="preset-info">
                <span class="p-name">{preset.name}</span>
                <span class="p-email">{preset.email}</span>
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Form nhập tùy chỉnh -->
      <div class="custom-oauth-form">
        <div class="input-group">
          <label for="social-email">Email tài khoản</label>
          <input id="social-email" type="email" bind:value={socialEmail} placeholder="name@domain.com" />
        </div>
        <div class="input-group">
          <label for="social-name">Họ và tên hiển thị</label>
          <input id="social-name" type="text" bind:value={socialName} placeholder="Nguyễn Văn A" />
        </div>
        <div class="input-group">
          <label for="social-avatar">Link Avatar (tùy chọn)</label>
          <input id="social-avatar" type="url" bind:value={socialAvatar} placeholder="https://..." />
        </div>
      </div>

      <div class="social-actions">
        <button type="button" class="btn-cancel" on:click={() => (activeSocialProvider = null)}>
          Hủy bỏ
        </button>
        <button
          type="button"
          class="btn-confirm-oauth {activeSocialProvider}"
          disabled={loading || !socialEmail}
          on:click={handleSocialLogin}
        >
          {loading ? 'Đang xác thực...' : `Tiếp tục với ${activeSocialProvider.toUpperCase()}`}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .auth-overlay {
    position: fixed;
    inset: 0;
    background: rgba(11, 15, 25, 0.88);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 999;
    padding: 20px;
  }
  .auth-card {
    width: 100%;
    max-width: 440px;
    border-radius: var(--radius-lg);
    padding: 32px 30px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5), 0 0 0 1px var(--border-glass);
    animation: fadeIn 0.25s ease-out;
  }
  @keyframes fadeIn {
    from { opacity: 0; transform: scale(0.96) translateY(8px); }
    to { opacity: 1; transform: scale(1) translateY(0); }
  }
  .brand { text-align: center; margin-bottom: 20px; }
  .logo-icon { font-size: 38px; margin-bottom: 6px; }
  .brand h2 { font-size: 23px; font-weight: 700; color: #fff; letter-spacing: -0.02em; }
  .subtitle { font-size: 13px; color: var(--text-muted); margin-top: 4px; }
  
  .error-badge {
    background: rgba(239, 68, 68, 0.15);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.3);
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    margin-bottom: 16px;
    text-align: center;
  }
  .success-badge {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.3);
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    margin-bottom: 16px;
    text-align: center;
  }

  /* Social Auth Styles */
  .social-auth-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 20px;
  }
  .social-btn-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
  }
  .social-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 10px 8px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    border: 1px solid var(--border-glass);
    background: rgba(255, 255, 255, 0.04);
    color: #fff;
  }
  .social-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3);
  }
  .btn-google:hover {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.25);
  }
  .btn-facebook:hover {
    background: rgba(24, 119, 242, 0.18);
    border-color: rgba(24, 119, 242, 0.4);
  }
  .btn-github:hover {
    background: rgba(255, 255, 255, 0.12);
    border-color: rgba(255, 255, 255, 0.3);
  }
  .social-icon { width: 18px; height: 18px; flex-shrink: 0; }

  .social-mini-row {
    display: flex;
    justify-content: center;
    gap: 12px;
  }
  .social-mini-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: var(--text-muted);
    font-size: 12px;
    cursor: pointer;
    transition: 0.2s;
  }
  .social-mini-btn:hover {
    color: #fff;
    border-color: rgba(255, 255, 255, 0.25);
    background: rgba(255, 255, 255, 0.08);
  }
  .mini-icon { width: 14px; height: 14px; }

  /* Divider */
  .divider {
    display: flex;
    align-items: center;
    gap: 12px;
    margin: 18px 0;
  }
  .divider .line {
    flex: 1;
    height: 1px;
    background: var(--border-glass);
  }
  .divider-text {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .input-group { margin-bottom: 14px; display: flex; flex-direction: column; gap: 6px; }
  .label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .input-group label { font-size: 13px; font-weight: 500; color: var(--text-muted); }
  .forgot-link {
    background: none;
    border: none;
    color: var(--accent-primary);
    font-size: 12px;
    cursor: pointer;
    padding: 0;
  }
  .forgot-link:hover { text-decoration: underline; }

  .input-group input {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 11px 14px;
    color: #fff;
    font-size: 14px;
    outline: none;
    transition: 0.2s;
  }
  .input-group input:focus {
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 2px var(--border-focus);
    background: rgba(255, 255, 255, 0.08);
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
    font-size: 14px;
    transition: 0.2s;
  }
  .btn-primary:hover:not(:disabled) { opacity: 0.95; transform: translateY(-1px); }
  .btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

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

  /* Social Interactive Dialog */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
  }
  .social-modal {
    width: 100%;
    max-width: 480px;
    border-radius: var(--radius-lg);
    padding: 26px;
    box-shadow: 0 25px 60px rgba(0, 0, 0, 0.6);
  }
  .social-modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }
  .provider-badge {
    padding: 6px 14px;
    border-radius: 20px;
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 0.02em;
  }
  .provider-badge.google { background: rgba(234, 67, 53, 0.2); color: #ea4335; border: 1px solid rgba(234, 67, 53, 0.4); }
  .provider-badge.facebook { background: rgba(24, 119, 242, 0.2); color: #3b82f6; border: 1px solid rgba(24, 119, 242, 0.4); }
  .provider-badge.github { background: rgba(255, 255, 255, 0.15); color: #fff; border: 1px solid rgba(255, 255, 255, 0.3); }
  .provider-badge.apple { background: rgba(255, 255, 255, 0.15); color: #f3f4f6; border: 1px solid rgba(255, 255, 255, 0.3); }
  .provider-badge.discord { background: rgba(88, 101, 242, 0.2); color: #818cf8; border: 1px solid rgba(88, 101, 242, 0.4); }

  .close-x {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 18px;
    cursor: pointer;
  }
  .close-x:hover { color: #fff; }

  .social-hint {
    font-size: 13px;
    color: var(--text-muted);
    line-height: 1.5;
    margin-bottom: 16px;
  }
  .presets-section {
    margin-bottom: 16px;
  }
  .section-tag {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-muted);
    display: block;
    margin-bottom: 8px;
  }
  .presets-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 8px;
  }
  .preset-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
    cursor: pointer;
    transition: 0.2s;
  }
  .preset-item:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.2);
  }
  .preset-item.selected {
    background: rgba(59, 130, 246, 0.15);
    border-color: var(--accent-primary);
  }
  .preset-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
  }
  .preset-info {
    display: flex;
    flex-direction: column;
  }
  .p-name { font-size: 13px; font-weight: 600; color: #fff; }
  .p-email { font-size: 12px; color: var(--text-muted); }

  .custom-oauth-form {
    border-top: 1px solid var(--border-glass);
    padding-top: 14px;
    margin-top: 14px;
  }
  .social-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 20px;
  }
  .btn-cancel {
    padding: 10px 18px;
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    color: var(--text-muted);
    cursor: pointer;
    font-size: 13px;
  }
  .btn-confirm-oauth {
    padding: 10px 22px;
    border-radius: var(--radius-sm);
    border: none;
    color: #fff;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    transition: 0.2s;
  }
  .btn-confirm-oauth.google { background: #ea4335; }
  .btn-confirm-oauth.facebook { background: #1877f2; }
  .btn-confirm-oauth.github { background: #24292e; border: 1px solid rgba(255, 255, 255, 0.3); }
  .btn-confirm-oauth.apple { background: #000; border: 1px solid rgba(255, 255, 255, 0.3); }
  .btn-confirm-oauth.discord { background: #5865f2; }
  .btn-confirm-oauth:hover:not(:disabled) {
    opacity: 0.92;
    transform: translateY(-1px);
  }
  .btn-confirm-oauth:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Forgot Password Styles */
  .forgot-container {
    animation: fadeIn 0.25s ease-out;
  }
  .forgot-header-info {
    text-align: center;
    margin-bottom: 20px;
  }
  .forgot-header-info h3 {
    font-size: 19px;
    font-weight: 700;
    color: #fff;
    margin: 0 0 6px 0;
  }
  .demo-otp-box {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: rgba(245, 158, 11, 0.12);
    border: 1px dashed rgba(245, 158, 11, 0.4);
    color: #fbbf24;
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    margin-bottom: 16px;
  }
  .btn-copy-otp {
    background: rgba(245, 158, 11, 0.25);
    border: 1px solid rgba(245, 158, 11, 0.5);
    color: #fef08a;
    font-weight: 600;
    padding: 3px 10px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 12px;
  }
  .btn-copy-otp:hover {
    background: rgba(245, 158, 11, 0.4);
  }
  .btn-resend {
    width: 100%;
    margin-top: 8px;
    background: none;
    border: 1px dashed var(--border-glass);
    color: var(--text-muted);
    padding: 10px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    cursor: pointer;
    transition: 0.2s;
  }
  .btn-resend:hover:not(:disabled) {
    color: #fff;
    border-color: rgba(255, 255, 255, 0.3);
  }
  .back-to-login {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }
  .back-to-login button {
    background: none;
    border: none;
    color: var(--accent-primary);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }
  .back-to-login button:hover {
    text-decoration: underline;
  }
</style>

