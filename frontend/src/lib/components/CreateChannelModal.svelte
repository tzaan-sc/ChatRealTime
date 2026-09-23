<script>
  import { createChannel, createCategory, activeGroup } from '../stores/chat';
  import { Hash, Megaphone, FolderPlus, Plus, X, Loader2 } from 'lucide-svelte';

  export let initialCategoryID = '';
  export let onClose = () => {};
  export let onCreated = () => {};

  let activeTab = 'channel'; // 'channel' | 'category'

  // Form Kênh
  let channelName = '';
  let channelType = 'text'; // 'text' | 'announcement'
  let selectedCategoryID = initialCategoryID || '';
  let channelDescription = '';

  // Form Danh mục
  let categoryName = '';

  let isSubmitting = false;
  let errorMessage = '';

  $: categories = $activeGroup?.categories || [];

  function sanitizeChannelName(val) {
    return val
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9\-_]/g, '');
  }

  function handleNameInput(e) {
    channelName = sanitizeChannelName(e.target.value);
  }

  async function handleSubmit() {
    errorMessage = '';
    if (!$activeGroup) return;

    if (activeTab === 'channel') {
      if (!channelName.trim()) {
        errorMessage = 'Vui lòng nhập tên kênh';
        return;
      }

      isSubmitting = true;
      try {
        await createChannel($activeGroup.id, {
          name: channelName.trim(),
          type: channelType,
          category_id: selectedCategoryID || undefined,
          description: channelDescription.trim() || undefined
        });
        onCreated();
        onClose();
      } catch (err) {
        errorMessage = err.message || 'Không thể tạo kênh';
      } finally {
        isSubmitting = false;
      }
    } else {
      if (!categoryName.trim()) {
        errorMessage = 'Vui lòng nhập tên danh mục';
        return;
      }

      isSubmitting = true;
      try {
        await createCategory($activeGroup.id, categoryName.trim().toUpperCase());
        onCreated();
        onClose();
      } catch (err) {
        errorMessage = err.message || 'Không thể tạo danh mục';
      } finally {
        isSubmitting = false;
      }
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click|self={onClose}>
  <div class="modal-box glass-card">
    <!-- Header -->
    <div class="modal-header">
      <div class="title-wrap">
        <Plus size={20} class="header-icon" />
        <h3>Thêm mới Không gian</h3>
      </div>
      <button class="close-btn" on:click={onClose} title="Đóng">
        <X size={18} />
      </button>
    </div>

    <!-- Tabs chuyển đổi Kênh / Danh mục -->
    <div class="tab-bar">
      <button
        class="tab-btn"
        class:active={activeTab === 'channel'}
        on:click={() => { activeTab = 'channel'; errorMessage = ''; }}
      >
        <Hash size={15} />
        <span>Tạo Kênh</span>
      </button>
      <button
        class="tab-btn"
        class:active={activeTab === 'category'}
        on:click={() => { activeTab = 'category'; errorMessage = ''; }}
      >
        <FolderPlus size={15} />
        <span>Tạo Danh Mục</span>
      </button>
    </div>

    {#if errorMessage}
      <div class="error-banner">{errorMessage}</div>
    {/if}

    <form on:submit|preventDefault={handleSubmit} class="form-content">
      {#if activeTab === 'channel'}
        <!-- Chọn Loại Kênh -->
        <div class="form-group">
          <label class="group-label">Loại Kênh</label>
          <div class="type-cards">
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <div
              class="type-card"
              class:selected={channelType === 'text'}
              on:click={() => (channelType = 'text')}
            >
              <div class="type-icon-box">
                <Hash size={20} />
              </div>
              <div class="type-info">
                <div class="type-title">Kênh Thảo Luận</div>
                <div class="type-desc">Gửi tin nhắn, hình ảnh, tài liệu và thảo luận tự do</div>
              </div>
              <input type="radio" bind:group={channelType} value="text" class="radio-dot" />
            </div>

            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <div
              class="type-card"
              class:selected={channelType === 'announcement'}
              on:click={() => (channelType = 'announcement')}
            >
              <div class="type-icon-box announcement-icon">
                <Megaphone size={20} />
              </div>
              <div class="type-info">
                <div class="type-title">Kênh Thông Báo (1 Chiều)</div>
                <div class="type-desc">Chỉ Quản trị viên mới có thể đăng bài thông báo quan trọng</div>
              </div>
              <input type="radio" bind:group={channelType} value="announcement" class="radio-dot" />
            </div>
          </div>
        </div>

        <!-- Tên Kênh -->
        <div class="form-group">
          <label for="c-name" class="group-label">Tên Kênh</label>
          <div class="input-with-prefix">
            <span class="prefix">{channelType === 'announcement' ? '📢' : '#'}</span>
            <input
              id="c-name"
              type="text"
              placeholder="ten-kenh-moi"
              value={channelName}
              on:input={handleNameInput}
              required
            />
          </div>
          <span class="hint-text">Chỉ cho phép chữ thường, số, dấu gạch nối (-).</span>
        </div>

        <!-- Chọn Danh Mục Thuộc Về -->
        {#if categories.length > 0}
          <div class="form-group">
            <label for="c-cat" class="group-label">Danh Mục</label>
            <select id="c-cat" bind:value={selectedCategoryID} class="custom-select">
              <option value="">(Không thuộc danh mục nào)</option>
              {#each categories as cat}
                <option value={cat.id}>{cat.name}</option>
              {/each}
            </select>
          </div>
        {/if}

        <!-- Mô Tả Kênh -->
        <div class="form-group">
          <label for="c-desc" class="group-label">Mô Tả Kênh (Tùy chọn)</label>
          <input
            id="c-desc"
            type="text"
            placeholder="Mục đích chính của kênh này..."
            bind:value={channelDescription}
          />
        </div>

      {:else}
        <!-- Form Danh Mục -->
        <div class="form-group">
          <label for="cat-name" class="group-label">Tên Danh Mục</label>
          <input
            id="cat-name"
            type="text"
            placeholder="VÍ DỤ: DỰ ÁN X, TÀI LIỆU..."
            bind:value={categoryName}
            required
          />
          <span class="hint-text">Dùng để nhóm các kênh liên quan lại với nhau (kiểu Discord).</span>
        </div>
      {/if}

      <!-- Actions -->
      <div class="modal-actions">
        <button type="button" class="cancel-btn" on:click={onClose} disabled={isSubmitting}>
          Hủy
        </button>
        <button type="submit" class="submit-btn" disabled={isSubmitting}>
          {#if isSubmitting}
            <Loader2 size={16} class="spinner" />
            <span>Đang tạo...</span>
          {:else}
            <span>{activeTab === 'channel' ? 'Tạo Kênh' : 'Tạo Danh Mục'}</span>
          {/if}
        </button>
      </div>
    </form>
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
    max-width: 480px;
    border-radius: var(--radius-lg);
    background: rgba(18, 22, 38, 0.95);
    border: 1px solid var(--border-glass);
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
    overflow: hidden;
    animation: scaleUp 0.2s ease;
  }

  .modal-header {
    padding: 18px 22px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--border-glass);
  }

  .title-wrap {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  :global(.header-icon) {
    color: #38bdf8;
  }

  .modal-header h3 {
    font-size: 17px;
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

  /* Tab Bar */
  .tab-bar {
    display: flex;
    padding: 6px 14px;
    background: rgba(0, 0, 0, 0.25);
    border-bottom: 1px solid var(--border-glass);
    gap: 8px;
  }

  .tab-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 8px 12px;
    background: transparent;
    border: none;
    border-radius: var(--radius-md);
    color: var(--text-dim);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.2s;
  }

  .tab-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.06);
  }

  .tab-btn.active {
    color: #fff;
    background: rgba(56, 189, 248, 0.18);
    font-weight: 600;
    box-shadow: 0 2px 8px rgba(56, 189, 248, 0.2);
  }

  .error-banner {
    margin: 12px 20px 0 20px;
    padding: 8px 12px;
    border-radius: 6px;
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #fca5a5;
    font-size: 13px;
  }

  /* Form */
  .form-content {
    padding: 20px 22px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .group-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  /* Type Cards */
  .type-cards {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .type-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 14px;
    border-radius: var(--radius-md);
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
    cursor: pointer;
    transition: 0.2s;
  }

  .type-card:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.2);
  }

  .type-card.selected {
    background: rgba(56, 189, 248, 0.12);
    border-color: #38bdf8;
  }

  .type-icon-box {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.08);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #38bdf8;
    flex-shrink: 0;
  }

  .announcement-icon {
    color: #f59e0b;
  }

  .type-info {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .type-title {
    font-size: 14px;
    font-weight: 600;
    color: #fff;
  }

  .type-desc {
    font-size: 11px;
    color: var(--text-dim);
    line-height: 1.3;
  }

  .radio-dot {
    accent-color: #38bdf8;
    pointer-events: none;
  }

  /* Input with prefix */
  .input-with-prefix {
    display: flex;
    align-items: center;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .prefix {
    padding: 0 10px;
    color: var(--text-dim);
    font-size: 14px;
    user-select: none;
  }

  .input-with-prefix input {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    padding: 10px 12px 10px 0;
    font-size: 14px;
    outline: none;
  }

  input[type='text'],
  .custom-select {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    color: #fff;
    padding: 10px 12px;
    font-size: 14px;
    outline: none;
    transition: 0.2s;
  }

  input[type='text']:focus,
  .custom-select:focus,
  .input-with-prefix:focus-within {
    border-color: #38bdf8;
    box-shadow: 0 0 10px rgba(56, 189, 248, 0.25);
  }

  .custom-select option {
    background: #121626;
    color: #fff;
  }

  .hint-text {
    font-size: 11px;
    color: var(--text-muted);
  }

  /* Modal Actions */
  .modal-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 10px;
  }

  .cancel-btn {
    padding: 9px 18px;
    border-radius: var(--radius-md);
    background: transparent;
    border: 1px solid var(--border-glass);
    color: var(--text-dim);
    cursor: pointer;
    font-size: 13px;
    transition: 0.2s;
  }

  .cancel-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.08);
  }

  .submit-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 9px 20px;
    border-radius: var(--radius-md);
    background: linear-gradient(135deg, #0284c7, #0369a1);
    border: none;
    color: #fff;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: 0.2s;
  }

  .submit-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 14px rgba(2, 132, 199, 0.4);
  }

  .submit-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  :global(.spinner) {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes scaleUp {
    from { opacity: 0; transform: scale(0.96); }
    to { opacity: 1; transform: scale(1); }
  }
</style>
