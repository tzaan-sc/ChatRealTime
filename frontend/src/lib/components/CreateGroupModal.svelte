<script>
  import { onMount } from 'svelte';
  import { token } from '../stores/auth';
  import { apiRequest } from '../services/api';
  import { loadUserGroups, selectGroup } from '../stores/chat';
  import { X, Users, Check, Search } from 'lucide-svelte';

  export let onClose = () => {};

  let groupName = '';
  let availableUsers = [];
  let selectedMemberIDs = [];
  let searchQuery = '';
  let isCreating = false;
  let errorMsg = '';

  onMount(async () => {
    try {
      const res = await apiRequest('/chat/users', 'GET', null, $token);
      availableUsers = res.data || [];
    } catch (err) {
      console.error('Lỗi tải danh sách bạn bè:', err);
    }
  });

  $: filteredUsers = availableUsers.filter((u) => {
    const q = searchQuery.toLowerCase();
    return (
      (u.username && u.username.toLowerCase().includes(q)) ||
      (u.display_name && u.display_name.toLowerCase().includes(q))
    );
  });

  function toggleSelect(userId) {
    if (selectedMemberIDs.includes(userId)) {
      selectedMemberIDs = selectedMemberIDs.filter((id) => id !== userId);
    } else {
      selectedMemberIDs = [...selectedMemberIDs, userId];
    }
  }

  async function handleCreateGroup() {
    if (!groupName.trim()) {
      errorMsg = 'Vui lòng nhập tên nhóm';
      return;
    }

    isCreating = true;
    errorMsg = '';

    try {
      const res = await apiRequest(
        '/groups',
        'POST',
        {
          name: groupName.trim(),
          member_ids: selectedMemberIDs
        },
        $token
      );

      await loadUserGroups();
      if (res.data) {
        selectGroup(res.data);
      }
      onClose();
    } catch (err) {
      errorMsg = err.message || 'Không thể tạo nhóm';
    } finally {
      isCreating = false;
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click={onClose}>
  <div class="modal-card glass-card" on:click={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div class="header-title">
        <Users size={20} class="icon-accent" />
        <h3>Tạo Nhóm Trò Chuyện Mới</h3>
      </div>
      <button class="close-btn" on:click={onClose} title="Đóng">
        <X size={18} />
      </button>
    </div>

    {#if errorMsg}
      <div class="error-banner">{errorMsg}</div>
    {/if}

    <div class="form-group">
      <label for="group-name-input">Tên nhóm</label>
      <input
        id="group-name-input"
        type="text"
        placeholder="Ví dụ: Team Dự Án, Hội Bạn Thân..."
        bind:value={groupName}
        maxlength="60"
        required
      />
    </div>

    <div class="members-section">
      <div class="members-header">
        <label for="search-friend-input">Mời bạn bè tham gia ({selectedMemberIDs.length} đã chọn)</label>
        <div class="search-wrap">
          <Search size={14} />
          <input
            id="search-friend-input"
            type="text"
            placeholder="Tìm kiếm bạn bè..."
            bind:value={searchQuery}
          />
        </div>
      </div>

      <div class="user-select-list">
        {#if filteredUsers.length === 0}
          <div class="empty-list">Không tìm thấy người dùng nào</div>
        {:else}
          {#each filteredUsers as u}
            {@const isSelected = selectedMemberIDs.includes(u.id)}
            <div
              class="user-select-row"
              class:selected={isSelected}
              on:click={() => toggleSelect(u.id)}
            >
              <img src={u.avatar_url} alt={u.username} class="user-avatar" />
              <div class="user-meta">
                <span class="name">{u.display_name || u.username}</span>
                <span class="uname">@{u.username}</span>
              </div>
              <div class="check-box" class:checked={isSelected}>
                {#if isSelected}
                  <Check size={14} />
                {/if}
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <div class="modal-actions">
      <button class="btn-secondary" on:click={onClose} type="button">Hủy</button>
      <button
        class="btn-primary"
        on:click={handleCreateGroup}
        disabled={isCreating || !groupName.trim()}
      >
        {isCreating ? 'Đang tạo...' : 'Tạo Nhóm'}
      </button>
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 20, 0.75);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    padding: 20px;
    animation: fadeIn 0.2s ease;
  }

  .modal-card {
    width: 100%;
    max-width: 460px;
    background: rgba(18, 22, 38, 0.94);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .header-title {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  :global(.icon-accent) { color: #a78bfa; }
  .modal-header h3 {
    font-size: 17px;
    font-weight: 600;
    color: #fff;
  }
  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    border-radius: 50%;
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .close-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.1); }

  .error-banner {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
    font-size: 13px;
    padding: 8px 12px;
    border-radius: var(--radius-sm);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  label {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-muted);
  }
  input[type="text"] {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    padding: 10px 14px;
    color: #fff;
    font-size: 14px;
    outline: none;
    transition: border-color 0.2s;
  }
  input[type="text"]:focus {
    border-color: #a78bfa;
  }

  .members-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .members-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .search-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 6px 10px;
    color: var(--text-dim);
  }
  .search-wrap input {
    background: transparent;
    border: none;
    outline: none;
    color: #fff;
    font-size: 12px;
    width: 100%;
  }

  .user-select-list {
    max-height: 200px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding-right: 4px;
  }
  .empty-list {
    text-align: center;
    color: var(--text-dim);
    font-size: 13px;
    padding: 20px 0;
  }
  .user-select-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-radius: var(--radius-md);
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid transparent;
    cursor: pointer;
    transition: 0.15s;
  }
  .user-select-row:hover {
    background: rgba(255, 255, 255, 0.08);
  }
  .user-select-row.selected {
    background: rgba(167, 139, 250, 0.12);
    border-color: rgba(167, 139, 250, 0.35);
  }
  .user-avatar {
    width: 34px;
    height: 34px;
    border-radius: 50%;
  }
  .user-meta {
    flex: 1;
    margin-left: 10px;
    display: flex;
    flex-direction: column;
  }
  .name {
    font-size: 13px;
    font-weight: 500;
    color: #fff;
  }
  .uname {
    font-size: 11px;
    color: var(--text-dim);
  }
  .check-box {
    width: 22px;
    height: 22px;
    border-radius: 6px;
    border: 1px solid var(--border-glass);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .check-box.checked {
    background: #a78bfa;
    border-color: #a78bfa;
    color: #fff;
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 8px;
  }
  .btn-secondary {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid var(--border-glass);
    color: #fff;
    padding: 8px 16px;
    border-radius: var(--radius-md);
    cursor: pointer;
  }
  .btn-secondary:hover { background: rgba(255, 255, 255, 0.14); }
  .btn-primary {
    background: var(--accent-gradient);
    border: none;
    color: #fff;
    padding: 8px 20px;
    border-radius: var(--radius-md);
    font-weight: 500;
    cursor: pointer;
    transition: 0.2s;
  }
  .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-primary:not(:disabled):hover { transform: scale(1.03); }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
