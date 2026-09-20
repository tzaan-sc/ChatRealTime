<script>
  import { onMount } from 'svelte';
  import { token, currentUser } from '../stores/auth';
  import { apiRequest } from '../services/api';
  import { loadUserGroups, activeGroup } from '../stores/chat';
  import { X, Shield, UserMinus, LogOut, UserPlus, Check } from 'lucide-svelte';

  export let groupID = '';
  export let onClose = () => {};

  let groupDetails = null;
  let availableFriends = [];
  let selectedNewMemberIDs = [];
  let isAddingMembers = false;
  let showAddSection = false;
  let isLoading = true;
  let errorMsg = '';

  $: amIAdmin = groupDetails?.members?.some(
    (m) => m.id === $currentUser?.id && m.is_admin
  );

  onMount(async () => {
    await fetchGroupDetails();
  });

  async function fetchGroupDetails() {
    isLoading = true;
    try {
      const res = await apiRequest(`/groups/${groupID}`, 'GET', null, $token);
      groupDetails = res.data;
    } catch (err) {
      errorMsg = err.message || 'Không thể tải thông tin nhóm';
    } finally {
      isLoading = false;
    }
  }

  async function openAddSection() {
    showAddSection = true;
    try {
      const res = await apiRequest('/chat/users', 'GET', null, $token);
      const allUsers = res.data || [];
      const currentMemberIDs = groupDetails?.members?.map((m) => m.id) || [];
      availableFriends = allUsers.filter((u) => !currentMemberIDs.includes(u.id));
    } catch (err) {
      console.error(err);
    }
  }

  function toggleSelectNew(userId) {
    if (selectedNewMemberIDs.includes(userId)) {
      selectedNewMemberIDs = selectedNewMemberIDs.filter((id) => id !== userId);
    } else {
      selectedNewMemberIDs = [...selectedNewMemberIDs, userId];
    }
  }

  async function handleAddMembers() {
    if (selectedNewMemberIDs.length === 0) return;
    isAddingMembers = true;
    try {
      await apiRequest(
        `/groups/${groupID}/members`,
        'POST',
        { member_ids: selectedNewMemberIDs },
        $token
      );
      selectedNewMemberIDs = [];
      showAddSection = false;
      await fetchGroupDetails();
      await loadUserGroups();
    } catch (err) {
      alert('Lỗi thêm thành viên: ' + err.message);
    } finally {
      isAddingMembers = false;
    }
  }

  async function handleRemoveMember(memberId, name) {
    if (!confirm(`Bạn có chắc muốn xóa ${name} khỏi nhóm?`)) return;
    try {
      await apiRequest(`/groups/${groupID}/members/${memberId}`, 'DELETE', null, $token);
      await fetchGroupDetails();
      await loadUserGroups();
    } catch (err) {
      alert('Lỗi xóa thành viên: ' + err.message);
    }
  }

  async function handleLeaveGroup() {
    if (!confirm('Bạn có chắc chắn muốn rời khỏi nhóm này?')) return;
    try {
      await apiRequest(`/groups/${groupID}/members/${$currentUser.id}`, 'DELETE', null, $token);
      activeGroup.set(null);
      await loadUserGroups();
      onClose();
    } catch (err) {
      alert('Lỗi rời nhóm: ' + err.message);
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click={onClose}>
  <div class="modal-card glass-card" on:click={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div class="group-summary">
        <img src={groupDetails?.avatar} alt="avatar" class="group-img" />
        <div>
          <h3>{groupDetails?.name || 'Nhóm'}</h3>
          <span class="member-count">{groupDetails?.members?.length || 0} thành viên</span>
        </div>
      </div>
      <button class="close-btn" on:click={onClose} title="Đóng">
        <X size={18} />
      </button>
    </div>

    {#if isLoading}
      <div class="loading-state">Đang tải thông tin nhóm...</div>
    {:else}
      {#if errorMsg}
        <div class="error-banner">{errorMsg}</div>
      {/if}

      <!-- Admin Add Member Button -->
      {#if amIAdmin && !showAddSection}
        <button class="add-member-trigger" on:click={openAddSection}>
          <UserPlus size={16} />
          <span>Thêm thành viên mới</span>
        </button>
      {/if}

      <!-- Add Member Section -->
      {#if showAddSection}
        <div class="add-section glass-card">
          <div class="add-section-header">
            <span>Chọn bạn bè để thêm ({selectedNewMemberIDs.length})</span>
            <button class="text-btn" on:click={() => (showAddSection = false)}>Đóng</button>
          </div>

          <div class="friend-pick-list">
            {#if availableFriends.length === 0}
              <div class="empty-text">Tất cả bạn bè đã ở trong nhóm</div>
            {:else}
              {#each availableFriends as f}
                {@const isSelected = selectedNewMemberIDs.includes(f.id)}
                <div
                  class="friend-row"
                  class:selected={isSelected}
                  on:click={() => toggleSelectNew(f.id)}
                >
                  <img src={f.avatar_url} alt="" class="f-avatar" />
                  <span class="f-name">{f.display_name || f.username}</span>
                  <div class="f-check" class:checked={isSelected}>
                    {#if isSelected}
                      <Check size={12} />
                    {/if}
                  </div>
                </div>
              {/each}
            {/if}
          </div>

          {#if availableFriends.length > 0}
            <button
              class="submit-add-btn"
              on:click={handleAddMembers}
              disabled={isAddingMembers || selectedNewMemberIDs.length === 0}
            >
              {isAddingMembers ? 'Đang thêm...' : 'Xác nhận thêm'}
            </button>
          {/if}
        </div>
      {/if}

      <!-- Member List -->
      <div class="member-list">
        {#each groupDetails?.members || [] as member}
          {@const isMe = member.id === $currentUser?.id}
          <div class="member-row">
            <img src={member.avatar_url} alt="" class="member-avatar" />
            <div class="member-info">
              <div class="name-line">
                <span class="name">{member.display_name || member.username}</span>
                {#if isMe}
                  <span class="me-badge">(Bạn)</span>
                {/if}
                {#if member.is_admin}
                  <span class="admin-badge" title="Quản trị viên nhóm">
                    <Shield size={11} /> Admin
                  </span>
                {/if}
              </div>
              <span class="uname">@{member.username}</span>
            </div>

            <!-- Action buttons -->
            {#if amIAdmin && !isMe}
              <button
                class="remove-btn"
                on:click={() => handleRemoveMember(member.id, member.display_name || member.username)}
                title="Xóa khỏi nhóm"
              >
                <UserMinus size={15} />
              </button>
            {/if}
          </div>
        {/each}
      </div>

      <!-- Leave Group Button -->
      <div class="modal-footer">
        <button class="leave-btn" on:click={handleLeaveGroup}>
          <LogOut size={16} />
          <span>Rời nhóm</span>
        </button>
      </div>
    {/if}
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
    max-width: 440px;
    background: rgba(18, 22, 38, 0.94);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 22px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
    max-height: 85vh;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--border-glass);
  }
  .group-summary {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .group-img {
    width: 44px;
    height: 44px;
    border-radius: 12px;
  }
  .group-summary h3 {
    font-size: 16px;
    font-weight: 600;
    color: #fff;
  }
  .member-count {
    font-size: 12px;
    color: var(--text-dim);
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

  .loading-state {
    text-align: center;
    color: var(--text-dim);
    font-size: 14px;
    padding: 30px 0;
  }
  .error-banner {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
    font-size: 13px;
    padding: 8px 12px;
    border-radius: var(--radius-sm);
  }

  .add-member-trigger {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: rgba(167, 139, 250, 0.15);
    border: 1px solid rgba(167, 139, 250, 0.3);
    color: #c4b5fd;
    padding: 8px 14px;
    border-radius: var(--radius-md);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.2s;
  }
  .add-member-trigger:hover {
    background: rgba(167, 139, 250, 0.25);
    color: #fff;
  }

  /* Add Section */
  .add-section {
    background: rgba(0, 0, 0, 0.3);
    border-radius: var(--radius-md);
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .add-section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 12px;
    color: var(--text-muted);
  }
  .text-btn {
    background: transparent;
    border: none;
    color: #a78bfa;
    font-size: 11px;
    cursor: pointer;
  }
  .friend-pick-list {
    max-height: 120px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .empty-text {
    font-size: 12px;
    color: var(--text-dim);
    text-align: center;
    padding: 10px 0;
  }
  .friend-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    background: rgba(255, 255, 255, 0.03);
  }
  .friend-row:hover { background: rgba(255, 255, 255, 0.08); }
  .friend-row.selected { background: rgba(167, 139, 250, 0.2); }
  .f-avatar { width: 26px; height: 26px; border-radius: 50%; }
  .f-name { flex: 1; font-size: 12px; color: #fff; }
  .f-check {
    width: 18px;
    height: 18px;
    border-radius: 4px;
    border: 1px solid var(--border-glass);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .f-check.checked {
    background: #a78bfa;
    color: #fff;
    border-color: #a78bfa;
  }
  .submit-add-btn {
    background: #10b981;
    border: none;
    color: #fff;
    font-size: 12px;
    font-weight: 500;
    padding: 6px 12px;
    border-radius: var(--radius-sm);
    cursor: pointer;
  }
  .submit-add-btn:disabled { opacity: 0.4; cursor: not-allowed; }

  /* Member List */
  .member-list {
    max-height: 280px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .member-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 10px;
    border-radius: var(--radius-md);
    background: rgba(255, 255, 255, 0.03);
  }
  .member-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
  }
  .member-info {
    flex: 1;
    margin-left: 10px;
    display: flex;
    flex-direction: column;
  }
  .name-line {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .name {
    font-size: 13px;
    font-weight: 500;
    color: #fff;
  }
  .me-badge {
    font-size: 10px;
    color: #38bdf8;
  }
  .admin-badge {
    font-size: 10px;
    color: #f59e0b;
    background: rgba(245, 158, 11, 0.15);
    padding: 1px 6px;
    border-radius: 8px;
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .uname {
    font-size: 11px;
    color: var(--text-dim);
  }
  .remove-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    width: 28px;
    height: 28px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .remove-btn:hover {
    color: #f87171;
    background: rgba(239, 68, 68, 0.2);
  }

  .modal-footer {
    padding-top: 10px;
    border-top: 1px solid var(--border-glass);
    display: flex;
    justify-content: flex-end;
  }
  .leave-btn {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
    padding: 8px 16px;
    border-radius: var(--radius-md);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: 0.2s;
  }
  .leave-btn:hover {
    background: rgba(239, 68, 68, 0.3);
  }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
