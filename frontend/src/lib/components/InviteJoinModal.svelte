<script>
  import { previewInvite, joinViaInvite, selectGroup, userGroups } from '../stores/chat';
  import { X, Users, Link2, ShieldCheck, Check, Loader2, AlertCircle } from 'lucide-svelte';

  export let initialCode = '';
  export let onClose = () => {};

  let inputCode = initialCode || '';
  let isLoadingPreview = false;
  let isSubmitting = false;
  let previewData = null;
  let errorMessage = '';
  let successMessage = '';

  if (initialCode) {
    loadPreview(initialCode);
  }

  function extractCode(str) {
    if (!str) return '';
    // Hỗ trợ cả url dạng http://.../join/inv_xxx hoặc mã inv_xxx
    const match = str.match(/inv_[a-zA-Z0-9]+/);
    return match ? match[0] : str.trim();
  }

  async function loadPreview(raw) {
    const code = extractCode(raw);
    if (!code) {
      previewData = null;
      return;
    }
    isLoadingPreview = true;
    errorMessage = '';
    try {
      previewData = await previewInvite(code);
    } catch (err) {
      errorMessage = err.message || 'Mã mời không hợp lệ hoặc đã hết hạn';
      previewData = null;
    } finally {
      isLoadingPreview = false;
    }
  }

  async function handleJoin() {
    if (!previewData || !previewData.code) return;
    isSubmitting = true;
    errorMessage = '';
    try {
      const res = await joinViaInvite(previewData.code);
      if (res.is_joined) {
        successMessage = 'Đã tham gia nhóm thành công!';
        setTimeout(() => {
          const joinedGroup = $userGroups.find((g) => g.id === previewData.group_id);
          if (joinedGroup) {
            selectGroup(joinedGroup);
          }
          onClose();
        }, 1000);
      } else if (res.requires_approval) {
        successMessage = 'Đã gửi yêu cầu tham gia! Vui lòng chờ Quản trị viên duyệt.';
        setTimeout(onClose, 2000);
      }
    } catch (err) {
      errorMessage = err.message || 'Không thể tham gia nhóm';
    } finally {
      isSubmitting = false;
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click|self={onClose}>
  <div class="modal-box glass-card">
    <div class="modal-header">
      <div class="title-wrap">
        <Link2 size={18} class="header-icon" />
        <h3>Tham gia Nhóm bằng Liên kết</h3>
      </div>
      <button class="close-btn" on:click={onClose} title="Đóng">
        <X size={18} />
      </button>
    </div>

    <div class="modal-body">
      <!-- Input mã hoặc dán link -->
      <div class="input-section">
        <label for="invite-link-input" class="input-label">Dán đường dẫn mời hoặc mã mời</label>
        <div class="input-row">
          <input
            id="invite-link-input"
            type="text"
            placeholder="vd: inv_1234567890 hoặc link mời..."
            bind:value={inputCode}
            on:input={() => loadPreview(inputCode)}
          />
          <button
            class="check-btn"
            on:click={() => loadPreview(inputCode)}
            disabled={!inputCode.trim() || isLoadingPreview}
          >
            {#if isLoadingPreview}
              <Loader2 size={15} class="spinner" />
            {:else}
              Kiểm tra
            {/if}
          </button>
        </div>
      </div>

      {#if errorMessage}
        <div class="alert-box error">
          <AlertCircle size={16} />
          <span>{errorMessage}</span>
        </div>
      {/if}

      {#if successMessage}
        <div class="alert-box success">
          <Check size={16} />
          <span>{successMessage}</span>
        </div>
      {/if}

      <!-- Xem trước nhóm -->
      {#if previewData}
        <div class="preview-card glass-card">
          <img
            src={previewData.group_avatar || `https://api.dicebear.com/7.x/identicon/svg?seed=${previewData.group_id}`}
            alt="Group Avatar"
            class="group-avatar"
          />
          <div class="preview-info">
            <h4 class="group-name">{previewData.group_name}</h4>
            <div class="group-stats">
              <span class="stat-item">
                <Users size={13} /> {previewData.member_count} thành viên
              </span>
              {#if previewData.inviter_name}
                <span class="inviter-tag">Mời bởi {previewData.inviter_name}</span>
              {/if}
            </div>

            {#if previewData.require_approval}
              <div class="approval-badge">
                <ShieldCheck size={13} />
                <span>Cần Quản trị viên duyệt trước khi vào</span>
              </div>
            {/if}
          </div>
        </div>

        {#if previewData.is_already_member}
          <div class="info-note">Bạn đã là thành viên của nhóm này rồi.</div>
        {:else if previewData.has_pending_req}
          <div class="info-note warning">Đơn xin gia nhập của bạn đang chờ Quản trị viên duyệt.</div>
        {/if}
      {/if}
    </div>

    <div class="modal-footer">
      <button class="cancel-btn" on:click={onClose} disabled={isSubmitting}>Hủy</button>
      {#if previewData && !previewData.is_already_member && !previewData.has_pending_req}
        <button
          class="submit-btn"
          on:click={handleJoin}
          disabled={isSubmitting || previewData.is_expired || previewData.is_maxedOut}
        >
          {#if isSubmitting}
            <Loader2 size={16} class="spinner" />
            <span>Đang tham gia...</span>
          {:else}
            <span>{previewData.require_approval ? 'Gửi yêu cầu tham gia' : 'Tham gia nhóm'}</span>
          {/if}
        </button>
      {/if}
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 20, 0.8);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
    padding: 16px;
    animation: fadeIn 0.2s ease;
  }

  .modal-box {
    width: 100%;
    max-width: 440px;
    border-radius: var(--radius-lg);
    background: rgba(18, 22, 38, 0.95);
    border: 1px solid var(--border-glass);
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
    overflow: hidden;
    animation: scaleUp 0.2s ease;
  }

  .modal-header {
    padding: 16px 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--border-glass);
  }

  .title-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  :global(.header-icon) {
    color: #38bdf8;
  }

  .modal-header h3 {
    font-size: 16px;
    font-weight: 600;
    color: #fff;
    margin: 0;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    border-radius: 50%;
    padding: 4px;
    display: flex;
  }

  .close-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .modal-body {
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .input-label {
    display: block;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 6px;
  }

  .input-row {
    display: flex;
    gap: 8px;
  }

  .input-row input {
    flex: 1;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    color: #fff;
    padding: 9px 12px;
    font-size: 13px;
    outline: none;
  }

  .input-row input:focus {
    border-color: #38bdf8;
  }

  .check-btn {
    padding: 0 16px;
    border-radius: var(--radius-md);
    background: rgba(56, 189, 248, 0.15);
    border: 1px solid rgba(56, 189, 248, 0.35);
    color: #38bdf8;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: 0.2s;
  }

  .check-btn:hover:not(:disabled) {
    background: rgba(56, 189, 248, 0.25);
  }

  .check-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .alert-box {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    border-radius: var(--radius-md);
    font-size: 13px;
  }

  .alert-box.error {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #fca5a5;
  }

  .alert-box.success {
    background: rgba(16, 185, 129, 0.15);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: #6ee7b7;
  }

  /* Preview Card */
  .preview-card {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 14px;
    border-radius: var(--radius-lg);
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
  }

  .group-avatar {
    width: 54px;
    height: 54px;
    border-radius: 14px;
    object-fit: cover;
  }

  .preview-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .group-name {
    font-size: 16px;
    font-weight: 700;
    color: #fff;
    margin: 0;
  }

  .group-stats {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 12px;
    color: var(--text-dim);
  }

  .stat-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .inviter-tag {
    color: #a78bfa;
  }

  .approval-badge {
    margin-top: 4px;
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 11px;
    font-weight: 600;
    color: #f59e0b;
  }

  .info-note {
    font-size: 12px;
    text-align: center;
    color: #38bdf8;
    padding: 8px;
    border-radius: 6px;
    background: rgba(56, 189, 248, 0.1);
  }

  .info-note.warning {
    color: #fbbf24;
    background: rgba(245, 158, 11, 0.1);
  }

  /* Footer */
  .modal-footer {
    padding: 14px 20px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    border-top: 1px solid var(--border-glass);
  }

  .cancel-btn {
    padding: 8px 16px;
    border-radius: var(--radius-md);
    background: transparent;
    border: 1px solid var(--border-glass);
    color: var(--text-dim);
    cursor: pointer;
    font-size: 13px;
  }

  .submit-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 18px;
    border-radius: var(--radius-md);
    background: var(--accent-gradient);
    border: none;
    color: #fff;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }

  .submit-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  :global(.spinner) {
    animation: spin 1s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
  @keyframes scaleUp { from { opacity: 0; transform: scale(0.96); } to { opacity: 1; transform: scale(1); } }
</style>
