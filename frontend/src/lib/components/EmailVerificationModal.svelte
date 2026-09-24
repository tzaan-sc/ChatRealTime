<script>
  import { currentUser, token, loginSuccess } from '../stores/auth';
  import { apiRequest } from '../services/api';
  import { MailCheck, CheckCircle2, AlertCircle, X, Send } from 'lucide-svelte';

  export let onClose = () => {};

  let otp = '';
  let loading = false;
  let errorMsg = '';
  let successMsg = '';
  let demoOtp = '';
  let sentOtp = false;

  async function handleSendOTP() {
    loading = true;
    errorMsg = '';
    successMsg = '';
    try {
      const res = await apiRequest('/auth/send-verification', 'POST', {
        email: $currentUser.email
      });
      demoOtp = res.demo_otp || '';
      sentOtp = true;
      successMsg = res.message || 'Mã xác thực đã được gửi tới email của bạn!';
    } catch (err) {
      errorMsg = err.message;
    } finally {
      loading = false;
    }
  }

  async function handleVerify() {
    if (!otp || otp.length !== 6) {
      errorMsg = 'Vui lòng nhập đầy đủ mã OTP 6 chữ số';
      return;
    }
    loading = true;
    errorMsg = '';
    successMsg = '';
    try {
      const res = await apiRequest('/auth/verify-email', 'POST', {
        email: $currentUser.email,
        otp: otp
      });

      // Update current user store with email_verified = true
      const updatedUser = { ...$currentUser, email_verified: true };
      loginSuccess(updatedUser, $token);
      successMsg = res.message || 'Xác thực email thành công!';
      setTimeout(() => {
        onClose();
      }, 1500);
    } catch (err) {
      errorMsg = err.message;
    } finally {
      loading = false;
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-overlay" on:click={onClose}>
  <div class="modal-box glass-card" on:click={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div class="header-left">
        <MailCheck size={22} class="mail-icon" />
        <div>
          <h3>Xác thực địa chỉ Email</h3>
          <p class="subtitle">Kích hoạt bảo mật tài khoản và nhận thông báo</p>
        </div>
      </div>
      <button class="close-btn" on:click={onClose}>
        <X size={18} />
      </button>
    </div>

    {#if errorMsg}
      <div class="error-badge">{errorMsg}</div>
    {/if}
    {#if successMsg}
      <div class="success-badge">{successMsg}</div>
    {/if}

    <div class="verification-body">
      <div class="email-info-card">
        <span class="label">Địa chỉ email tài khoản:</span>
        <span class="email-text">{$currentUser?.email}</span>
        {#if $currentUser?.email_verified}
          <div class="verified-badge">
            <CheckCircle2 size={15} /> Đã kích hoạt bảo mật
          </div>
        {:else}
          <div class="unverified-badge">
            <AlertCircle size={15} /> Chưa được xác thực
          </div>
        {/if}
      </div>

      {#if !$currentUser?.email_verified}
        {#if demoOtp}
          <div class="demo-otp-box">
            <span>✨ Mã OTP kích hoạt: <strong>{demoOtp}</strong></span>
            <button type="button" class="btn-copy-otp" on:click={() => { otp = demoOtp; }}>
              Tự điền
            </button>
          </div>
        {/if}

        {#if !sentOtp}
          <div class="send-prompt">
            <p>Bấm nút bên dưới để nhận mã OTP xác thực qua email. Mã có hiệu lực trong vòng 24 giờ.</p>
            <button
              type="button"
              class="btn-primary"
              disabled={loading}
              on:click={handleSendOTP}
            >
              <Send size={16} /> {loading ? 'Đang gửi...' : 'Gửi mã OTP kích hoạt'}
            </button>
          </div>
        {:else}
          <form on:submit|preventDefault={handleVerify} class="otp-form">
            <div class="input-group">
              <label for="verify-email-otp">Nhập mã OTP (6 chữ số)</label>
              <input
                id="verify-email-otp"
                type="text"
                maxlength="6"
                bind:value={otp}
                placeholder="123456"
                required
                style="letter-spacing: 0.25em; font-size: 18px; font-weight: 700; text-align: center;"
              />
            </div>
            <button type="submit" class="btn-primary" disabled={loading || otp.length !== 6}>
              {loading ? 'Đang xác thực...' : 'Xác nhận Kích hoạt Email'}
            </button>
            <button
              type="button"
              class="btn-resend"
              disabled={loading}
              on:click={handleSendOTP}
            >
              Gửi lại mã OTP
            </button>
          </form>
        {/if}
      {:else}
        <div class="already-verified-box">
          <p>Email này đã được xác thực an toàn. Bạn có thể sử dụng đầy đủ các tính năng khôi phục tài khoản và bảo mật nâng cao!</p>
        </div>
      {/if}
    </div>

    <div class="modal-footer">
      <button class="btn-done" on:click={onClose}>Đóng</button>
    </div>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(5px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
  }
  .modal-box {
    width: 100%;
    max-width: 480px;
    border-radius: var(--radius-lg);
    padding: 26px;
    box-shadow: 0 25px 60px rgba(0, 0, 0, 0.6);
  }
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    border-bottom: 1px solid var(--border-glass);
    padding-bottom: 14px;
  }
  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  :global(.mail-icon) {
    color: var(--accent-primary);
  }
  .modal-header h3 {
    font-size: 17px;
    font-weight: 700;
    color: #fff;
    margin: 0;
  }
  .subtitle {
    font-size: 12px;
    color: var(--text-muted);
    margin: 2px 0 0 0;
  }
  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px;
    border-radius: 50%;
  }
  .close-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.08); }

  .error-badge {
    background: rgba(239, 68, 68, 0.15);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.3);
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    margin-bottom: 16px;
  }
  .success-badge {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.3);
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    margin-bottom: 16px;
  }

  .email-info-card {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 14px 16px;
    margin-bottom: 18px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .email-info-card .label {
    font-size: 12px;
    color: var(--text-muted);
  }
  .email-info-card .email-text {
    font-size: 15px;
    font-weight: 600;
    color: #fff;
  }
  .verified-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    color: #34d399;
    background: rgba(16, 185, 129, 0.15);
    padding: 4px 10px;
    border-radius: 20px;
    width: fit-content;
    margin-top: 6px;
  }
  .unverified-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    font-weight: 600;
    color: #fbbf24;
    background: rgba(245, 158, 11, 0.15);
    padding: 4px 10px;
    border-radius: 20px;
    width: fit-content;
    margin-top: 6px;
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

  .send-prompt p {
    font-size: 13px;
    color: var(--text-muted);
    line-height: 1.5;
    margin-bottom: 14px;
  }
  .input-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 14px;
  }
  .input-group label {
    font-size: 13px;
    color: var(--text-muted);
  }
  .input-group input {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 12px;
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
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    font-size: 14px;
    transition: 0.2s;
  }
  .btn-primary:hover:not(:disabled) {
    opacity: 0.95;
    transform: translateY(-1px);
  }
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-resend {
    width: 100%;
    margin-top: 8px;
    background: none;
    border: 1px dashed var(--border-glass);
    color: var(--text-muted);
    padding: 9px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    cursor: pointer;
    transition: 0.2s;
  }
  .btn-resend:hover:not(:disabled) {
    color: #fff;
    border-color: rgba(255, 255, 255, 0.3);
  }

  .already-verified-box {
    background: rgba(16, 185, 129, 0.08);
    border: 1px solid rgba(16, 185, 129, 0.2);
    border-radius: var(--radius-sm);
    padding: 14px;
    color: #6ee7b7;
    font-size: 13px;
    line-height: 1.5;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid var(--border-glass);
    padding-top: 16px;
    margin-top: 16px;
  }
  .btn-done {
    padding: 8px 22px;
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid var(--border-glass);
    color: #fff;
    font-weight: 600;
    cursor: pointer;
    font-size: 13px;
  }
  .btn-done:hover {
    background: rgba(255, 255, 255, 0.14);
  }
</style>
