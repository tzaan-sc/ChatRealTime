<script>
  import { onMount } from 'svelte';
  import { token, currentUser } from '../stores/auth';
  import { apiRequest } from '../services/api';
  import {
    loadUserGroups,
    activeGroup,
    updateMemberRole,
    muteMember,
    updateGroupSettings,
    createGroupInvite,
    getGroupInvites,
    revokeGroupInvite,
    getPendingJoinRequests,
    approveJoinRequest,
    rejectJoinRequest
  } from '../stores/chat';
  import {
    X,
    Shield,
    UserMinus,
    LogOut,
    UserPlus,
    Check,
    Link2,
    UserCheck,
    Clock,
    VolumeX,
    Volume2,
    Copy,
    Trash2,
    Search,
    ShieldAlert,
    Sliders
  } from 'lucide-svelte';

  export let groupID = '';
  export let onClose = () => {};

  let groupDetails = null;
  let availableFriends = [];
  let selectedNewMemberIDs = [];
  let isAddingMembers = false;
  let showAddSection = false;
  let isLoading = true;
  let errorMsg = '';
  let successMsg = '';

  // Tabs: 'members' | 'invites' | 'approvals'
  let activeTab = 'members';

  // Member search
  let memberSearchQuery = '';

  // Mute modal state
  let activeMuteTarget = null;
  let selectedMuteDuration = 60; // minutes

  // Invites state
  let groupInvites = [];
  let isLoadingInvites = false;
  let inviteMaxUses = 0;
  let inviteExpireHours = 24; // 1 day
  let isCreatingInvite = false;
  let copiedInviteCode = '';

  // Join requests state
  let joinRequests = [];
  let isLoadingRequests = false;
  let isTogglingApproval = false;

  $: myRole = groupDetails?.members?.find((m) => m.id === $currentUser?.id)?.role || 'member';
  $: isOwner = myRole === 'owner';
  $: isAdmin = isOwner || myRole === 'admin';
  $: isModerator = isAdmin || myRole === 'moderator';

  $: filteredMembers = (groupDetails?.members || []).filter((m) => {
    if (!memberSearchQuery.trim()) return true;
    const q = memberSearchQuery.toLowerCase();
    return (
      (m.display_name && m.display_name.toLowerCase().includes(q)) ||
      (m.username && m.username.toLowerCase().includes(q))
    );
  });

  onMount(async () => {
    await fetchGroupDetails();
  });

  async function fetchGroupDetails() {
    isLoading = true;
    errorMsg = '';
    try {
      const res = await apiRequest(`/groups/${groupID}`, 'GET', null, $token);
      groupDetails = res.data;
      if (isModerator) {
        loadInvites();
        loadJoinRequests();
      }
    } catch (err) {
      errorMsg = err.message || 'Không thể tải thông tin nhóm';
    } finally {
      isLoading = false;
    }
  }

  async function loadInvites() {
    isLoadingInvites = true;
    try {
      groupInvites = await getGroupInvites(groupID);
    } catch (err) {
      console.error('Lỗi tải danh sách mã mời:', err);
    } finally {
      isLoadingInvites = false;
    }
  }

  async function loadJoinRequests() {
    isLoadingRequests = true;
    try {
      joinRequests = await getPendingJoinRequests(groupID);
    } catch (err) {
      console.error('Lỗi tải yêu cầu tham gia:', err);
    } finally {
      isLoadingRequests = false;
    }
  }

  // --- Add Member Section ---
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
      alert('Lỗi thao tác: ' + err.message);
    }
  }

  // --- RBAC Role Updates ---
  function canChangeRole(targetMember) {
    if (!targetMember || targetMember.id === $currentUser?.id) return false;
    if (targetMember.role === 'owner') return false;
    if (isOwner) return true;
    if (isAdmin && targetMember.role !== 'admin') return true;
    return false;
  }

  async function handleRoleChange(targetMember, newRole) {
    if (!canChangeRole(targetMember)) return;
    try {
      await updateMemberRole(groupID, targetMember.id, newRole);
      targetMember.role = newRole;
      targetMember.is_admin = newRole === 'admin' || newRole === 'owner';
      groupDetails = { ...groupDetails };
    } catch (err) {
      alert('Lỗi phân quyền: ' + (err.message || 'Không thể cập nhật vai trò'));
    }
  }

  // --- Muting ---
  function canMute(targetMember) {
    if (!targetMember || targetMember.id === $currentUser?.id) return false;
    if (targetMember.role === 'owner') return false;
    if (isOwner) return true;
    if (isAdmin && targetMember.role !== 'admin') return true;
    if (myRole === 'moderator' && targetMember.role === 'member') return true;
    return false;
  }

  function openMuteModal(member) {
    activeMuteTarget = member;
    selectedMuteDuration = 60;
  }

  async function submitMute() {
    if (!activeMuteTarget) return;
    try {
      await muteMember(groupID, activeMuteTarget.id, selectedMuteDuration);
      activeMuteTarget = null;
      await fetchGroupDetails();
    } catch (err) {
      alert('Lỗi thao tác cấm chat: ' + err.message);
    }
  }

  async function handleUnmute(member) {
    try {
      await muteMember(groupID, member.id, 0);
      await fetchGroupDetails();
    } catch (err) {
      alert('Lỗi hủy cấm chat: ' + err.message);
    }
  }

  // --- Slow Mode ---
  async function handleUpdateSlowMode(seconds) {
    try {
      await apiRequest(`/groups/${groupID}/slowmode`, 'PATCH', { seconds }, $token);
      groupDetails.slow_mode_seconds = seconds;
      activeGroup.update((g) => (g ? { ...g, slow_mode_seconds: seconds } : g));
      loadUserGroups();
    } catch (e) {
      alert('Không thể cập nhật chế độ chậm: ' + e.message);
    }
  }

  // --- Leave Group ---
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

  // --- Tab 2: Invites ---
  async function handleCreateInvite() {
    isCreatingInvite = true;
    errorMsg = '';
    try {
      await createGroupInvite(groupID, {
        max_uses: Number(inviteMaxUses),
        expire_hours: Number(inviteExpireHours)
      });
      await loadInvites();
    } catch (err) {
      alert('Lỗi tạo mã mời: ' + err.message);
    } finally {
      isCreatingInvite = false;
    }
  }

  async function handleRevokeInvite(code) {
    if (!confirm('Bạn có chắc muốn thu hồi mã mời này?')) return;
    try {
      await revokeGroupInvite(groupID, code);
      groupInvites = groupInvites.filter((inv) => inv.code !== code);
    } catch (err) {
      alert('Lỗi thu hồi mã mời: ' + err.message);
    }
  }

  function copyInviteLink(code) {
    const origin = window.location.origin;
    const url = `${origin}/join/${code}`;
    navigator.clipboard.writeText(url);
    copiedInviteCode = code;
    setTimeout(() => {
      if (copiedInviteCode === code) copiedInviteCode = '';
    }, 2000);
  }

  // --- Tab 3: Approvals ---
  async function handleToggleApproval() {
    if (isTogglingApproval) return;
    isTogglingApproval = true;
    const nextVal = !groupDetails.require_approval;
    try {
      await updateGroupSettings(groupID, { require_approval: nextVal });
      groupDetails.require_approval = nextVal;
    } catch (err) {
      alert('Lỗi cập nhật cài đặt: ' + err.message);
    } finally {
      isTogglingApproval = false;
    }
  }

  async function handleApproveRequest(reqId) {
    try {
      await approveJoinRequest(groupID, reqId);
      joinRequests = joinRequests.filter((r) => r.id !== reqId);
      await fetchGroupDetails();
      await loadUserGroups();
    } catch (err) {
      alert('Lỗi duyệt thành viên: ' + err.message);
    }
  }

  async function handleRejectRequest(reqId) {
    try {
      await rejectJoinRequest(groupID, reqId);
      joinRequests = joinRequests.filter((r) => r.id !== reqId);
    } catch (err) {
      alert('Lỗi từ chối yêu cầu: ' + err.message);
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click={onClose}>
  <div class="modal-card glass-card" on:click={(e) => e.stopPropagation()}>
    <!-- Modal Header -->
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

    <!-- Tab Bar -->
    <div class="tab-bar">
      <button
        class="tab-btn"
        class:active={activeTab === 'members'}
        on:click={() => (activeTab = 'members')}
      >
        <span>Thành viên</span>
        <span class="tab-pill">{groupDetails?.members?.length || 0}</span>
      </button>

      {#if isModerator}
        <button
          class="tab-btn"
          class:active={activeTab === 'invites'}
          on:click={() => {
            activeTab = 'invites';
            loadInvites();
          }}
        >
          <Link2 size={13} />
          <span>Liên kết mời</span>
        </button>

        <button
          class="tab-btn"
          class:active={activeTab === 'approvals'}
          on:click={() => {
            activeTab = 'approvals';
            loadJoinRequests();
          }}
        >
          <UserCheck size={13} />
          <span>Duyệt yêu cầu</span>
          {#if joinRequests.length > 0}
            <span class="tab-pill-alert">{joinRequests.length}</span>
          {/if}
        </button>
      {/if}
    </div>

    {#if isLoading}
      <div class="loading-state">Đang tải thông tin...</div>
    {:else}
      {#if errorMsg}
        <div class="error-banner">{errorMsg}</div>
      {/if}

      <!-- TAB 1: THÀNH VIÊN & PHÂN QUYỀN -->
      {#if activeTab === 'members'}
        <!-- Search bar & Add Button -->
        <div class="members-subhead">
          <div class="member-search-box">
            <Search size={14} class="s-icon" />
            <input
              type="text"
              placeholder="Tìm theo tên hoặc @username..."
              bind:value={memberSearchQuery}
            />
          </div>
          {#if isAdmin && !showAddSection}
            <button class="add-member-trigger" on:click={openAddSection}>
              <UserPlus size={14} />
              <span>Thêm</span>
            </button>
          {/if}
        </div>

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

        <!-- Slow Mode Setting for Admins -->
        {#if isAdmin}
          <div class="slowmode-setting-box">
            <div class="slowmode-setting-header">
              <span>⏱️ Chế độ chậm (Slow Mode chống spam):</span>
            </div>
            <div class="slowmode-btn-group">
              {#each [0, 5, 10, 30, 60, 120] as sec}
                <button
                  class="sm-btn"
                  class:active={(groupDetails?.slow_mode_seconds || 0) === sec}
                  on:click={() => handleUpdateSlowMode(sec)}
                >
                  {sec === 0 ? 'Tắt' : `${sec}s`}
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Member List -->
        <div class="member-list">
          {#if filteredMembers.length === 0}
            <div class="empty-state-text">Không tìm thấy thành viên phù hợp.</div>
          {:else}
            {#each filteredMembers as member}
              {@const isMe = member.id === $currentUser?.id}
              {@const userCanChangeRole = canChangeRole(member)}
              {@const userCanMute = canMute(member)}
              {@const userCanKick = (isAdmin && member.role !== 'owner' && (isOwner || member.role !== 'admin')) && !isMe}

              <div class="member-row" class:muted-border={member.is_muted}>
                <div class="avatar-wrap">
                  <img src={member.avatar_url} alt="" class="member-avatar" />
                  {#if member.is_muted}
                    <div class="avatar-mute-badge" title="Đang bị cấm chat">🔇</div>
                  {/if}
                </div>

                <div class="member-info">
                  <div class="name-line">
                    <span class="name">{member.display_name || member.username}</span>
                    {#if isMe}
                      <span class="me-badge">(Bạn)</span>
                    {/if}

                    <!-- Role Badge -->
                    {#if member.role === 'owner'}
                      <span class="role-badge role-owner" title="Chủ sở hữu nhóm">👑 Chủ phòng</span>
                    {:else if member.role === 'admin'}
                      <span class="role-badge role-admin" title="Quản trị viên">🛡️ Admin</span>
                    {:else if member.role === 'moderator'}
                      <span class="role-badge role-mod" title="Điều hành viên">⚔️ Mod</span>
                    {:else}
                      <span class="role-badge role-member">Thành viên</span>
                    {/if}

                    {#if member.is_muted}
                      <span class="mute-tag" title="Tạm thời không thể gửi tin nhắn">🔇 Cấm chat</span>
                    {/if}
                  </div>
                  <span class="uname">@{member.username}</span>
                </div>

                <!-- Controls & Actions -->
                <div class="member-actions">
                  <!-- Role Select Dropdown (for Owner/Admin) -->
                  {#if userCanChangeRole}
                    <select
                      class="role-select"
                      value={member.role}
                      on:change={(e) => handleRoleChange(member, e.target.value)}
                    >
                      {#if isOwner}
                        <option value="admin">Admin</option>
                      {/if}
                      <option value="moderator">Mod</option>
                      <option value="member">Thành viên</option>
                    </select>
                  {/if}

                  <!-- Mute / Unmute Button -->
                  {#if userCanMute}
                    {#if member.is_muted}
                      <button
                        class="action-icon-btn unmute-btn"
                        on:click={() => handleUnmute(member)}
                        title="Bỏ cấm chat"
                      >
                        <Volume2 size={14} />
                      </button>
                    {:else}
                      <button
                        class="action-icon-btn mute-btn"
                        on:click={() => openMuteModal(member)}
                        title="Cấm chat thành viên"
                      >
                        <VolumeX size={14} />
                      </button>
                    {/if}
                  {/if}

                  <!-- Kick Button -->
                  {#if userCanKick}
                    <button
                      class="action-icon-btn remove-btn"
                      on:click={() => handleRemoveMember(member.id, member.display_name || member.username)}
                      title="Xóa khỏi nhóm"
                    >
                      <UserMinus size={14} />
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          {/if}
        </div>

        <!-- Leave Group Button -->
        <div class="modal-footer">
          <button class="leave-btn" on:click={handleLeaveGroup}>
            <LogOut size={15} />
            <span>Rời nhóm</span>
          </button>
        </div>
      {/if}

      <!-- TAB 2: LIÊN KẾT MỜI (INVITE LINKS) -->
      {#if activeTab === 'invites' && isModerator}
        <div class="tab-content-container">
          <!-- Create Invite Box -->
          <div class="create-invite-box glass-card">
            <h4>Tạo liên kết mời mới</h4>
            <div class="invite-controls-row">
              <div class="control-group">
                <label>Thời hạn:</label>
                <select bind:value={inviteExpireHours} class="control-select">
                  <option value={1}>1 giờ</option>
                  <option value={6}>6 giờ</option>
                  <option value={24}>24 giờ (1 ngày)</option>
                  <option value={168}>7 ngày</option>
                  <option value={720}>30 ngày</option>
                  <option value={0}>Vô thời hạn</option>
                </select>
              </div>

              <div class="control-group">
                <label>Số lượt tối đa:</label>
                <select bind:value={inviteMaxUses} class="control-select">
                  <option value={0}>Không giới hạn</option>
                  <option value={1}>1 lượt dùng</option>
                  <option value={5}>5 lượt dùng</option>
                  <option value={10}>10 lượt dùng</option>
                  <option value={25}>25 lượt dùng</option>
                  <option value={50}>50 lượt dùng</option>
                </select>
              </div>
            </div>

            <button
              class="create-invite-btn"
              on:click={handleCreateInvite}
              disabled={isCreatingInvite}
            >
              <Link2 size={15} />
              <span>{isCreatingInvite ? 'Đang tạo...' : 'Tạo liên kết'}</span>
            </button>
          </div>

          <!-- Existing Invites List -->
          <div class="invites-list-title">Liên kết đang hoạt động ({groupInvites.length})</div>
          {#if isLoadingInvites}
            <div class="loading-state">Đang tải danh sách liên kết...</div>
          {:else if groupInvites.length === 0}
            <div class="empty-state-text">Chưa có liên kết mời nào được tạo.</div>
          {:else}
            <div class="invites-list">
              {#each groupInvites as inv}
                {@const isCopied = copiedInviteCode === inv.code}
                <div class="invite-card">
                  <div class="invite-card-left">
                    <span class="invite-code">{inv.code}</span>
                    <div class="invite-meta">
                      <span class="meta-item">
                        Lượt dùng: <b>{inv.used_count}/{inv.max_uses === 0 ? '∞' : inv.max_uses}</b>
                      </span>
                      <span class="meta-dot">•</span>
                      <span class="meta-item">
                        Hạn: <b>{inv.expires_at ? new Date(inv.expires_at).toLocaleDateString() : 'Vĩnh viễn'}</b>
                      </span>
                    </div>
                  </div>

                  <div class="invite-card-actions">
                    <button
                      class="copy-btn"
                      class:copied={isCopied}
                      on:click={() => copyInviteLink(inv.code)}
                      title="Sao chép link mời"
                    >
                      {#if isCopied}
                        <Check size={14} />
                        <span>Đã chép</span>
                      {:else}
                        <Copy size={14} />
                        <span>Chép link</span>
                      {/if}
                    </button>

                    <button
                      class="revoke-btn"
                      on:click={() => handleRevokeInvite(inv.code)}
                      title="Thu hồi liên kết này"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}

      <!-- TAB 3: DUYỆT THÀNH VIÊN & CÀI ĐẶT (APPROVALS) -->
      {#if activeTab === 'approvals' && isModerator}
        <div class="tab-content-container">
          <!-- Toggle Requirement -->
          <div class="approval-setting-card glass-card">
            <div class="setting-text">
              <div class="setting-title">Yêu cầu xét duyệt thành viên</div>
              <div class="setting-desc">
                Khi bật, người dùng tham gia qua link mời sẽ cần được Admin hoặc Mod phê duyệt trước khi vào nhóm.
              </div>
            </div>
            <label class="switch-toggle">
              <input
                type="checkbox"
                checked={groupDetails?.require_approval}
                on:change={handleToggleApproval}
                disabled={isTogglingApproval}
              />
              <span class="slider round"></span>
            </label>
          </div>

          <!-- Pending Requests List -->
          <div class="invites-list-title">Yêu cầu đang chờ duyệt ({joinRequests.length})</div>

          {#if isLoadingRequests}
            <div class="loading-state">Đang tải yêu cầu...</div>
          {:else if joinRequests.length === 0}
            <div class="empty-state-text">Không có yêu cầu tham gia nào đang chờ duyệt.</div>
          {:else}
            <div class="requests-list">
              {#each joinRequests as req}
                <div class="request-card">
                  <img src={req.user_avatar || 'https://api.dicebear.com/7.x/identicon/svg?seed=' + req.user_id} alt="" class="req-avatar" />
                  <div class="req-info">
                    <span class="req-name">{req.display_name || req.username}</span>
                    <span class="req-uname">@{req.username}</span>
                    <span class="req-time">Gửi lúc: {new Date(req.created_at).toLocaleString([], { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit' })}</span>
                  </div>

                  <div class="req-actions">
                    <button
                      class="approve-btn"
                      on:click={() => handleApproveRequest(req.id)}
                      title="Chấp thuận vào nhóm"
                    >
                      <Check size={14} />
                      <span>Duyệt</span>
                    </button>
                    <button
                      class="reject-btn"
                      on:click={() => handleRejectRequest(req.id)}
                      title="Từ chối yêu cầu"
                    >
                      <X size={14} />
                      <span>Từ chối</span>
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    {/if}
  </div>
</div>

<!-- Inline Mute Modal Dialog -->
{#if activeMuteTarget}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="sub-modal-backdrop" on:click={() => (activeMuteTarget = null)}>
    <div class="sub-modal-card glass-card" on:click={(e) => e.stopPropagation()}>
      <div class="sub-modal-header">
        <h4>🔇 Cấm chat: {activeMuteTarget.display_name || activeMuteTarget.username}</h4>
        <button class="close-btn" on:click={() => (activeMuteTarget = null)}><X size={16} /></button>
      </div>
      <p class="sub-modal-desc">
        Thành viên này sẽ không thể gửi tin nhắn trong tất cả các kênh của nhóm trong thời gian được chọn:
      </p>

      <div class="duration-options">
        {#each [
          { val: 15, label: '15 phút' },
          { val: 60, label: '1 giờ' },
          { val: 1440, label: '24 giờ' },
          { val: 10080, label: '7 ngày' }
        ] as opt}
          <button
            class="duration-chip"
            class:active={selectedMuteDuration === opt.val}
            on:click={() => (selectedMuteDuration = opt.val)}
          >
            {opt.label}
          </button>
        {/each}
      </div>

      <div class="sub-modal-footer">
        <button class="cancel-text-btn" on:click={() => (activeMuteTarget = null)}>Hủy</button>
        <button class="confirm-mute-btn" on:click={submitMute}>
          Xác nhận cấm chat
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 20, 0.78);
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
    max-width: 500px;
    background: rgba(18, 22, 38, 0.95);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 22px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
    max-height: 88vh;
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 10px;
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
    object-fit: cover;
  }
  .group-summary h3 {
    font-size: 16px;
    font-weight: 600;
    color: #fff;
    margin: 0;
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
    padding: 5px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .close-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.1); }

  /* Tab Bar */
  .tab-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    padding-bottom: 4px;
  }
  .tab-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 12.5px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.15s;
  }
  .tab-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.04);
  }
  .tab-btn.active {
    color: #a78bfa;
    background: rgba(167, 139, 250, 0.12);
    font-weight: 600;
  }
  .tab-pill {
    font-size: 11px;
    padding: 1px 6px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.08);
  }
  .tab-pill-alert {
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 10px;
    background: #ef4444;
    color: #fff;
    font-weight: 600;
  }

  .loading-state, .empty-state-text {
    text-align: center;
    color: var(--text-dim);
    font-size: 13px;
    padding: 24px 0;
  }
  .error-banner {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
    font-size: 12.5px;
    padding: 8px 12px;
    border-radius: var(--radius-sm);
  }

  /* Members Tab Subhead */
  .members-subhead {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .member-search-box {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 6px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 6px 10px;
  }
  .member-search-box input {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    font-size: 12px;
    outline: none;
  }
  .add-member-trigger {
    display: flex;
    align-items: center;
    gap: 5px;
    background: rgba(167, 139, 250, 0.15);
    border: 1px solid rgba(167, 139, 250, 0.3);
    color: #c4b5fd;
    padding: 6px 12px;
    border-radius: 8px;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.2s;
    white-space: nowrap;
  }
  .add-member-trigger:hover {
    background: rgba(167, 139, 250, 0.25);
    color: #fff;
  }

  /* Add Section */
  .add-section {
    background: rgba(0, 0, 0, 0.3);
    border-radius: var(--radius-md);
    padding: 10px;
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
    max-height: 110px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .empty-text {
    font-size: 11px;
    color: var(--text-dim);
    text-align: center;
    padding: 8px 0;
  }
  .friend-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 8px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    background: rgba(255, 255, 255, 0.03);
  }
  .friend-row:hover { background: rgba(255, 255, 255, 0.08); }
  .friend-row.selected { background: rgba(167, 139, 250, 0.2); }
  .f-avatar { width: 24px; height: 24px; border-radius: 50%; }
  .f-name { flex: 1; font-size: 12px; color: #fff; }
  .f-check {
    width: 16px;
    height: 16px;
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

  /* Slow Mode Setting */
  .slowmode-setting-box {
    padding: 8px 12px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 8px;
  }
  .slowmode-setting-header {
    font-size: 11.5px;
    color: #94a3b8;
    margin-bottom: 5px;
  }
  .slowmode-btn-group {
    display: flex;
    gap: 5px;
  }
  .sm-btn {
    padding: 3px 8px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #cbd5e1;
    border-radius: 5px;
    font-size: 11px;
    cursor: pointer;
    transition: 0.15s;
  }
  .sm-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.12); }
  .sm-btn.active {
    background: #6366f1;
    color: #fff;
    border-color: #818cf8;
    font-weight: 600;
  }

  /* Member List */
  .member-list {
    max-height: 290px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding-right: 2px;
  }
  .member-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 10px;
    border-radius: var(--radius-md);
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid transparent;
    transition: 0.15s;
  }
  .member-row:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  .member-row.muted-border {
    border-color: rgba(239, 68, 68, 0.3);
    background: rgba(239, 68, 68, 0.04);
  }
  .avatar-wrap {
    position: relative;
    width: 36px;
    height: 36px;
  }
  .member-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
  }
  .avatar-mute-badge {
    position: absolute;
    bottom: -2px;
    right: -2px;
    font-size: 11px;
    background: #1e1b2e;
    border-radius: 50%;
    padding: 1px;
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
    flex-wrap: wrap;
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
  .uname {
    font-size: 11px;
    color: var(--text-dim);
  }

  /* Role Badges */
  .role-badge {
    font-size: 10px;
    padding: 1px 7px;
    border-radius: 10px;
    font-weight: 600;
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .role-owner {
    background: rgba(251, 191, 36, 0.15);
    color: #fbbf24;
    border: 1px solid rgba(251, 191, 36, 0.3);
  }
  .role-admin {
    background: rgba(245, 158, 11, 0.15);
    color: #f59e0b;
    border: 1px solid rgba(245, 158, 11, 0.3);
  }
  .role-mod {
    background: rgba(6, 182, 212, 0.15);
    color: #06b6d4;
    border: 1px solid rgba(6, 182, 212, 0.3);
  }
  .role-member {
    background: rgba(148, 163, 184, 0.1);
    color: #94a3b8;
  }
  .mute-tag {
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 8px;
    background: rgba(239, 68, 68, 0.15);
    color: #f87171;
    font-weight: 500;
  }

  /* Member Actions */
  .member-actions {
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .role-select {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #cbd5e1;
    border-radius: 6px;
    font-size: 11px;
    padding: 3px 6px;
    outline: none;
    cursor: pointer;
  }
  .role-select option {
    background: #1e1b2e;
    color: #fff;
  }
  .action-icon-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    width: 26px;
    height: 26px;
    border-radius: 6px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.15s;
  }
  .action-icon-btn:hover { background: rgba(255, 255, 255, 0.08); color: #fff; }
  .mute-btn:hover { color: #f59e0b; background: rgba(245, 158, 11, 0.15); }
  .unmute-btn { color: #10b981; }
  .unmute-btn:hover { color: #34d399; background: rgba(16, 185, 129, 0.15); }
  .remove-btn:hover { color: #f87171; background: rgba(239, 68, 68, 0.15); }

  .modal-footer {
    padding-top: 10px;
    border-top: 1px solid var(--border-glass);
    display: flex;
    justify-content: flex-end;
  }
  .leave-btn {
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.25);
    color: #f87171;
    padding: 7px 14px;
    border-radius: var(--radius-md);
    font-size: 12.5px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 7px;
    transition: 0.2s;
  }
  .leave-btn:hover { background: rgba(239, 68, 68, 0.25); }

  /* Tab 2: Invites */
  .tab-content-container {
    max-height: 380px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding-right: 2px;
  }
  .create-invite-box {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .create-invite-box h4 {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: #fff;
  }
  .invite-controls-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }
  .control-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .control-group label {
    font-size: 11px;
    color: #94a3b8;
  }
  .control-select {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: #fff;
    padding: 6px 8px;
    border-radius: 6px;
    font-size: 12px;
    outline: none;
    cursor: pointer;
  }
  .control-select option {
    background: #1e1b2e;
    color: #fff;
  }
  .create-invite-btn {
    background: #6366f1;
    border: none;
    color: #fff;
    padding: 8px 14px;
    border-radius: 6px;
    font-size: 12.5px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: 0.2s;
  }
  .create-invite-btn:hover { background: #4f46e5; }
  .create-invite-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  .invites-list-title {
    font-size: 12px;
    font-weight: 600;
    color: #94a3b8;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-top: 4px;
  }
  .invites-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .invite-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
    padding: 8px 12px;
  }
  .invite-code {
    font-family: monospace;
    font-size: 13px;
    color: #c4b5fd;
    font-weight: 600;
  }
  .invite-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-dim);
    margin-top: 2px;
  }
  .meta-dot { color: rgba(255, 255, 255, 0.2); }
  .invite-card-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .copy-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    background: rgba(167, 139, 250, 0.15);
    border: 1px solid rgba(167, 139, 250, 0.25);
    color: #c4b5fd;
    font-size: 11.5px;
    padding: 4px 8px;
    border-radius: 6px;
    cursor: pointer;
    transition: 0.15s;
  }
  .copy-btn:hover { background: rgba(167, 139, 250, 0.25); }
  .copy-btn.copied {
    background: rgba(16, 185, 129, 0.2);
    border-color: rgba(16, 185, 129, 0.4);
    color: #34d399;
  }
  .revoke-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    padding: 5px;
    border-radius: 6px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .revoke-btn:hover { color: #f87171; background: rgba(239, 68, 68, 0.15); }

  /* Tab 3: Approvals */
  .approval-setting-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 10px;
    padding: 12px 14px;
  }
  .setting-title {
    font-size: 13px;
    font-weight: 600;
    color: #fff;
  }
  .setting-desc {
    font-size: 11.5px;
    color: #94a3b8;
    margin-top: 3px;
    line-height: 1.4;
  }
  /* Toggle Switch */
  .switch-toggle {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    flex-shrink: 0;
  }
  .switch-toggle input { opacity: 0; width: 0; height: 0; }
  .slider {
    position: absolute;
    cursor: pointer;
    inset: 0;
    background-color: rgba(255, 255, 255, 0.15);
    transition: 0.3s;
    border-radius: 24px;
  }
  .slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
  }
  input:checked + .slider { background-color: #6366f1; }
  input:checked + .slider:before { transform: translateX(20px); }

  .requests-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .request-card {
    display: flex;
    align-items: center;
    gap: 10px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 8px;
    padding: 8px 12px;
  }
  .req-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
  }
  .req-info {
    flex: 1;
    display: flex;
    flex-direction: column;
  }
  .req-name {
    font-size: 13px;
    font-weight: 500;
    color: #fff;
  }
  .req-uname {
    font-size: 11px;
    color: var(--text-dim);
  }
  .req-time {
    font-size: 10px;
    color: #64748b;
    margin-top: 1px;
  }
  .req-actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .approve-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    background: #10b981;
    border: none;
    color: #fff;
    font-size: 11.5px;
    font-weight: 500;
    padding: 5px 10px;
    border-radius: 6px;
    cursor: pointer;
    transition: 0.15s;
  }
  .approve-btn:hover { background: #059669; }
  .reject-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
    font-size: 11.5px;
    font-weight: 500;
    padding: 5px 10px;
    border-radius: 6px;
    cursor: pointer;
    transition: 0.15s;
  }
  .reject-btn:hover { background: rgba(239, 68, 68, 0.25); }

  /* Sub-Modal (Mute Duration Picker) */
  .sub-modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.65);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
    padding: 20px;
  }
  .sub-modal-card {
    width: 100%;
    max-width: 360px;
    background: #171b2e;
    border: 1px solid var(--border-glass);
    border-radius: 12px;
    padding: 18px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    box-shadow: 0 15px 40px rgba(0, 0, 0, 0.7);
  }
  .sub-modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .sub-modal-header h4 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: #fff;
  }
  .sub-modal-desc {
    font-size: 12px;
    color: #94a3b8;
    margin: 0;
    line-height: 1.4;
  }
  .duration-options {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }
  .duration-chip {
    padding: 8px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #cbd5e1;
    border-radius: 8px;
    font-size: 12px;
    cursor: pointer;
    transition: 0.15s;
    text-align: center;
  }
  .duration-chip:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
  }
  .duration-chip.active {
    background: #6366f1;
    border-color: #818cf8;
    color: #fff;
    font-weight: 600;
  }
  .sub-modal-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 10px;
    margin-top: 4px;
  }
  .cancel-text-btn {
    background: transparent;
    border: none;
    color: #94a3b8;
    font-size: 12px;
    cursor: pointer;
  }
  .cancel-text-btn:hover { color: #fff; }
  .confirm-mute-btn {
    background: #ef4444;
    border: none;
    color: #fff;
    font-size: 12px;
    font-weight: 600;
    padding: 7px 14px;
    border-radius: 6px;
    cursor: pointer;
    transition: 0.15s;
  }
  .confirm-mute-btn:hover { background: #dc2626; }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
