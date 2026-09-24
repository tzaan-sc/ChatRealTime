<script>
  import { currentUser, token, loginSuccess } from '../stores/auth';
  import { apiRequest } from '../services/api';
  import { ShieldCheck, Check, Plus, Trash2, X, ExternalLink } from 'lucide-svelte';

  export let onClose = () => {};

  let loading = false;
  let errorMsg = '';
  let successMsg = '';

  const allProviders = [
    { id: 'google', name: 'Google', desc: 'Đăng nhập 1-chạm & bảo mật xác thực', color: '#ea4335' },
    { id: 'facebook', name: 'Facebook', desc: 'Đồng bộ danh tính mạng xã hội Meta', color: '#1877f2' },
    { id: 'github', name: 'GitHub', desc: 'Tài khoản nhà phát triển & mã nguồn', color: '#24292e' },
    { id: 'apple', name: 'Apple ID', desc: 'Bảo mật quyền riêng tư chuẩn Apple', color: '#ffffff' },
    { id: 'discord', name: 'Discord', desc: 'Cộng đồng game & kênh trò chuyện', color: '#5865f2' }
  ];

  // Kiểm tra tài khoản đã liên kết provider nào
  $: linkedMap = ($currentUser?.oauth_accounts || []).reduce((acc, item) => {
    acc[item.provider] = item;
    return acc;
  }, {});

  async function handleLink(providerId) {
    loading = true;
    errorMsg = '';
    successMsg = '';
    try {
      const email = `${$currentUser.username}.${providerId}@gmail.com`;
      const payload = {
        provider: providerId,
        email: email,
        name: $currentUser.display_name || $currentUser.username,
        avatar_url: $currentUser.avatar_url,
        provider_id: `${providerId}_${$currentUser.id}`
      };

      await apiRequest('/auth/oauth/link', 'POST', payload, $token);
      
      // Refresh user profile
      const meRes = await apiRequest('/auth/me', 'GET', null, $token);
      loginSuccess(meRes.data, $token);
      successMsg = `Đã liên kết thành công với tài khoản ${providerId.toUpperCase()}!`;
    } catch (err) {
      errorMsg = err.message;
    } finally {
      loading = false;
    }
  }

  async function handleUnlink(providerId) {
    if (!confirm(`Bạn có chắc chắn muốn gỡ liên kết tài khoản ${providerId.toUpperCase()}?`)) return;
    loading = true;
    errorMsg = '';
    successMsg = '';
    try {
      await apiRequest(`/auth/oauth/unlink/${providerId}`, 'DELETE', null, $token);
      
      // Refresh user profile
      const meRes = await apiRequest('/auth/me', 'GET', null, $token);
      loginSuccess(meRes.data, $token);
      successMsg = `Đã gỡ liên kết tài khoản ${providerId.toUpperCase()} thành công.`;
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
        <ShieldCheck size={22} class="shield-icon" />
        <div>
          <h3>Liên kết mạng xã hội & Danh tính</h3>
          <p class="subtitle">Quản lý các tài khoản bên thứ 3 dùng để đăng nhập</p>
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

    <div class="provider-list">
      {#each allProviders as prov}
        {@const isLinked = !!linkedMap[prov.id]}
        {@const linkedItem = linkedMap[prov.id]}
        <div class="provider-row" class:is-linked={isLinked}>
          <div class="prov-info">
            <div class="prov-avatar-badge" style="border-color: {prov.color}">
              <span class="prov-name-char">{prov.name.charAt(0)}</span>
            </div>
            <div>
              <div class="prov-title-line">
                <span class="prov-title">{prov.name}</span>
                {#if isLinked}
                  <span class="badge-linked">
                    <Check size={12} /> Đã liên kết
                  </span>
                {/if}
              </div>
              <p class="prov-desc">
                {isLinked ? `Email: ${linkedItem.email}` : prov.desc}
              </p>
            </div>
          </div>

          <div class="prov-action">
            {#if isLinked}
              <button
                class="btn-unlink"
                title="Gỡ liên kết"
                disabled={loading}
                on:click={() => handleUnlink(prov.id)}
              >
                <Trash2 size={15} /> Gỡ
              </button>
            {:else}
              <button
                class="btn-link"
                disabled={loading}
                on:click={() => handleLink(prov.id)}
              >
                <Plus size={15} /> Liên kết
              </button>
            {/if}
          </div>
        </div>
      {/each}
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
    max-width: 520px;
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
  :global(.shield-icon) {
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

  .provider-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 20px;
  }
  .provider-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px;
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--border-glass);
    transition: 0.2s;
  }
  .provider-row.is-linked {
    background: rgba(16, 185, 129, 0.05);
    border-color: rgba(16, 185, 129, 0.2);
  }
  .prov-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .prov-avatar-badge {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.06);
    border: 2px solid;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    color: #fff;
    font-size: 14px;
  }
  .prov-title-line {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .prov-title {
    font-size: 14px;
    font-weight: 600;
    color: #fff;
  }
  .badge-linked {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    font-weight: 600;
    color: #34d399;
    background: rgba(16, 185, 129, 0.15);
    padding: 2px 8px;
    border-radius: 12px;
  }
  .prov-desc {
    font-size: 12px;
    color: var(--text-muted);
    margin: 3px 0 0 0;
  }

  .btn-link {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 7px 14px;
    border-radius: var(--radius-sm);
    background: rgba(59, 130, 246, 0.15);
    border: 1px solid rgba(59, 130, 246, 0.3);
    color: #60a5fa;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: 0.2s;
  }
  .btn-link:hover:not(:disabled) {
    background: rgba(59, 130, 246, 0.25);
    color: #fff;
  }
  .btn-unlink {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 7px 12px;
    border-radius: var(--radius-sm);
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.25);
    color: #f87171;
    font-size: 12px;
    cursor: pointer;
    transition: 0.2s;
  }
  .btn-unlink:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.2);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid var(--border-glass);
    padding-top: 16px;
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
