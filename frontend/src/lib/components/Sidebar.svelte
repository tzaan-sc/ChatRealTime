<script>
  import {
    conversations,
    activeConversation,
    selectConversation,
    userGroups,
    activeGroup,
    selectGroup,
    userDirectory,
    onlineUsers,
    groupUnreadCounts,
    markConversationAsUnread,
    markConversationAsRead,
    drafts,
    openSavedMessages
  } from '../stores/chat';
  import { currentUser, logout, token } from '../stores/auth';
  import { notificationSettings, updateNotificationSettings } from '../stores/notification';
  import { requestNotificationPermission, getNotificationPermission } from '../services/notification';
  import { apiRequest } from '../services/api';
  import {
    MessageSquarePlus,
    Users,
    UserPlus,
    LogOut,
    Search,
    MessageSquare,
    Bell,
    BellOff,
    MoreVertical,
    Eye,
    EyeOff,
    Bookmark,
    Link2,
    ShieldCheck,
    ShieldAlert
  } from 'lucide-svelte';
  import { onMount, onDestroy } from 'svelte';
  import CreateGroupModal from './CreateGroupModal.svelte';
  import InviteJoinModal from './InviteJoinModal.svelte';
  import SocialAccountsModal from './SocialAccountsModal.svelte';
  import EmailVerificationModal from './EmailVerificationModal.svelte';
  import SecuritySessionsModal from './SecuritySessionsModal.svelte';

  let showUsersModal = false;
  let showCreateGroupModal = false;
  let showInviteJoinModal = false;
  let showSocialModal = false;
  let showEmailVerifyModal = false;
  let showSecurityModal = false;
  let loadingUsers = false;
  let usersError = '';
  let searchQuery = '';
  let activeTab = 'direct'; // 'direct' hoặc 'groups'
  let activeContextMenuConvId = null;

  // Tính tổng số tin nhắn chưa đọc
  $: totalDirectUnread = $conversations.reduce((sum, item) => sum + (item.unread_count || 0), 0);
  $: totalGroupUnread = Object.values($groupUnreadCounts).reduce((sum, count) => sum + (count || 0), 0);

  async function toggleSound() {
    const isEnabled = $notificationSettings.soundEnabled;
    updateNotificationSettings({ soundEnabled: !isEnabled });
    if (!isEnabled && getNotificationPermission() === 'default') {
      await requestNotificationPermission();
    }
  }

  function toggleContextMenu(convId, event) {
    event.stopPropagation();
    activeContextMenuConvId = activeContextMenuConvId === convId ? null : convId;
  }

  function handleMarkUnread(convItem, event) {
    event.stopPropagation();
    markConversationAsUnread(convItem.conversation.custom_id);
    activeContextMenuConvId = null;
  }

  function handleMarkRead(convItem, event) {
    event.stopPropagation();
    markConversationAsRead(convItem.conversation.custom_id, convItem.other_user?.id);
    activeContextMenuConvId = null;
  }

  function closeMenu() {
    activeContextMenuConvId = null;
  }

  onMount(() => {
    window.addEventListener('click', closeMenu);
    return () => {
      window.removeEventListener('click', closeMenu);
    };
  });

  // Lọc danh sách hội thoại 1-1
  $: filteredConversations = $conversations.filter((item) => {
    if (!searchQuery.trim()) return true;
    const name = (item.other_user.display_name || item.other_user.username || '').toLowerCase();
    return name.includes(searchQuery.toLowerCase());
  });

  // Lọc danh sách nhóm
  $: filteredGroups = $userGroups.filter((g) => {
    if (!searchQuery.trim()) return true;
    return (g.name || '').toLowerCase().includes(searchQuery.toLowerCase());
  });

  async function openNewChatModal() {
    showUsersModal = true;
    loadingUsers = true;
    usersError = '';
    try {
      const res = await apiRequest('/chat/users', 'GET', null, $token);
      userDirectory.set(res.data || []);
    } catch (e) {
      console.error('Lỗi tải danh bạ:', e);
      usersError = e.message || 'Không thể tải danh sách người dùng';
    } finally {
      loadingUsers = false;
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
      <button
        class="icon-btn"
        class:muted={!$notificationSettings.soundEnabled}
        title={$notificationSettings.soundEnabled ? 'Chuông thông báo: BẬT (Bấm để tắt)' : 'Chuông thông báo: TẮT (Bấm để bật)'}
        on:click={toggleSound}
      >
        {#if $notificationSettings.soundEnabled}
          <Bell size={18} />
        {:else}
          <BellOff size={18} style="color: #ef4444;" />
        {/if}
      </button>
      <button class="icon-btn" title="Liên kết tài khoản mạng xã hội (OAuth)" on:click={() => (showSocialModal = true)}>
        <ShieldCheck size={19} />
      </button>
      <button class="icon-btn" title="Bảo mật phiên & Thiết bị đăng nhập" on:click={() => (showSecurityModal = true)}>
        <ShieldAlert size={19} />
      </button>
      <button class="icon-btn" title="Tham gia nhóm bằng liên kết mời" on:click={() => (showInviteJoinModal = true)}>
        <Link2 size={19} />
      </button>
      <button class="icon-btn" title="Tạo nhóm trò chuyện mới" on:click={() => (showCreateGroupModal = true)}>
        <Users size={19} />
      </button>
      <button class="icon-btn" title="Nhắn tin cá nhân mới" on:click={openNewChatModal}>
        <MessageSquarePlus size={20} />
      </button>
      <button class="icon-btn" title="Đăng xuất" on:click={logout}>
        <LogOut size={19} />
      </button>
    </div>
  </div>

  <!-- Banner cảnh báo chưa xác thực Email -->
  {#if $currentUser && !$currentUser.email_verified}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="unverified-banner" on:click={() => (showEmailVerifyModal = true)} role="button" tabindex="0">
      <span class="warning-icon">⚠️</span>
      <span class="banner-text">Email chưa xác thực. <u>Kích hoạt ngay</u></span>
    </div>
  {/if}

  <!-- Search Bar -->
  <div class="search-box">
    <Search size={16} class="search-icon" />
    <input
      type="text"
      placeholder={activeTab === 'direct' ? 'Tìm bạn bè...' : 'Tìm nhóm trò chuyện...'}
      bind:value={searchQuery}
    />
  </div>

  <!-- Tab Switcher (Cá nhân / Nhóm) -->
  <div class="sidebar-tabs">
    <button
      class="tab-btn"
      class:active={activeTab === 'direct'}
      on:click={() => (activeTab = 'direct')}
    >
      <MessageSquare size={15} />
      <span>Cá nhân ({$conversations.length})</span>
      {#if totalDirectUnread > 0}
        <span class="tab-badge">{totalDirectUnread > 99 ? '99+' : totalDirectUnread}</span>
      {/if}
    </button>
    <button
      class="tab-btn"
      class:active={activeTab === 'groups'}
      on:click={() => (activeTab = 'groups')}
    >
      <Users size={15} />
      <span>Nhóm ({$userGroups.length})</span>
      {#if totalGroupUnread > 0}
        <span class="tab-badge">{totalGroupUnread > 99 ? '99+' : totalGroupUnread}</span>
      {/if}
    </button>
  </div>

  <!-- List View -->
  <div class="conversation-list">
    {#if activeTab === 'direct'}
      <!-- Hộp thư lưu trữ cá nhân (Saved Messages / Cloud Notebook) -->
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div
        class="conversation-item saved-messages-item"
        class:active={$activeConversation?.is_saved}
        role="button"
        tabindex="0"
        on:click={openSavedMessages}
        title="Hộp thư lưu trữ cá nhân & Ghi chú đám mây"
      >
        <div class="avatar-container saved-avatar-wrap">
          <Bookmark size={18} class="saved-icon" />
        </div>
        <div class="content">
          <div class="top-line">
            <span class="name saved-name">Tin nhắn đã lưu</span>
            <span class="time">Cloud</span>
          </div>
          <div class="bottom-line">
            <p class="snippet saved-snippet">Ghi chú & Lưu trữ cá nhân</p>
          </div>
        </div>
      </div>

      <!-- Danh sách Chat 1-1 -->
      {#if filteredConversations.length === 0}
        <div class="empty-state">
          {searchQuery.trim() ? 'Không tìm thấy cuộc trò chuyện phù hợp' : 'Chưa có cuộc trò chuyện nào. Bấm nút dấu cộng để bắt đầu nhắn tin!'}
        </div>
      {:else}
        {#each filteredConversations as item}
          {@const draftText = $drafts[item.conversation?.custom_id]}
          <div
            class="conversation-item"
            class:active={$activeConversation?.conversation?.custom_id === item.conversation.custom_id}
            role="button"
            tabindex="0"
            on:click={() => selectConversation(item)}
            on:keydown={(e) => e.key === 'Enter' && selectConversation(item)}
            on:contextmenu={(e) => { e.preventDefault(); toggleContextMenu(item.conversation.custom_id, e); }}
          >
            <div class="avatar-container">
              <img src={item.other_user.avatar_url} alt="avatar" class="avatar" />
              <span class="user-status-dot" class:online={$onlineUsers.has(item.other_user.id)}></span>
            </div>
            <div class="content">
              <div class="top-line">
                <span class="name">{item.other_user.display_name || item.other_user.username}</span>
                <span class="time">
                  {item.conversation.last_message_at ? new Date(item.conversation.last_message_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
                </span>
              </div>
              <div class="bottom-line">
                <p class="snippet">
                  {#if draftText}
                    <span class="draft-tag">[Bản nháp]</span> {draftText}
                  {:else}
                    {item.conversation.last_message || 'Bắt đầu cuộc trò chuyện'}
                  {/if}
                </p>
                <div class="badge-and-menu">
                  {#if item.unread_count > 0}
                    <span class="unread-badge animate-pulse">{item.unread_count > 99 ? '99+' : item.unread_count}</span>
                  {/if}
                  <button
                    class="item-more-btn"
                    title="Tùy chọn"
                    on:click={(e) => toggleContextMenu(item.conversation.custom_id, e)}
                  >
                    <MoreVertical size={14} />
                  </button>
                </div>
              </div>
            </div>

            <!-- Context Menu Dropdown -->
            {#if activeContextMenuConvId === item.conversation.custom_id}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div class="item-context-menu" on:click|stopPropagation>
                {#if item.unread_count > 0}
                  <button class="context-item" on:click={(e) => handleMarkRead(item, e)}>
                    <Eye size={14} />
                    <span>Đánh dấu là đã đọc</span>
                  </button>
                {:else}
                  <button class="context-item" on:click={(e) => handleMarkUnread(item, e)}>
                    <EyeOff size={14} />
                    <span>Đánh dấu là chưa đọc</span>
                  </button>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      {/if}
    {:else}
      <!-- Danh sách Chat Nhóm -->
      {#if filteredGroups.length === 0}
        <div class="empty-state">
          <p>{searchQuery.trim() ? 'Không tìm thấy nhóm nào phù hợp' : 'Bạn chưa tham gia nhóm nào.'}</p>
          <div class="empty-group-btns">
            <button class="create-group-prompt-btn" on:click={() => (showCreateGroupModal = true)}>
              + Tạo nhóm ngay
            </button>
            <button class="join-link-prompt-btn" on:click={() => (showInviteJoinModal = true)}>
              🔗 Tham gia bằng link
            </button>
          </div>
        </div>
      {:else}
        {#each filteredGroups as group}
          {@const draftText = $drafts[group.id]}
          <div
            class="conversation-item group-item"
            class:active={$activeGroup?.id === group.id}
            role="button"
            tabindex="0"
            on:click={() => selectGroup(group)}
            on:keydown={(e) => e.key === 'Enter' && selectGroup(group)}
          >
            <div class="avatar-container">
              <img src={group.avatar} alt="group-avatar" class="avatar group-avatar-shape" />
            </div>
            <div class="content">
              <div class="top-line">
                <span class="name group-title">{group.name}</span>
                <span class="time">
                  {group.updated_at ? new Date(group.updated_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
                </span>
              </div>
              <div class="bottom-line">
                <p class="snippet member-tag">
                  {#if draftText}
                    <span class="draft-tag">[Bản nháp]</span> {draftText}
                  {:else}
                    {group.member_ids?.length || 0} thành viên
                  {/if}
                </p>
                {#if ($groupUnreadCounts[group.id] || 0) > 0}
                  <span class="unread-badge group-unread-badge">{$groupUnreadCounts[group.id]}</span>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      {/if}
    {/if}
  </div>
</aside>

<!-- Modal Chọn Bạn Chat Mới -->
{#if showUsersModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-overlay" on:click={() => (showUsersModal = false)}>
    <div class="modal-box glass-card" on:click={(e) => e.stopPropagation()}>
      <div class="modal-top">
        <h3>Bắt đầu cuộc trò chuyện</h3>
        <button class="close-btn" on:click={() => (showUsersModal = false)}>✕</button>
      </div>

      {#if loadingUsers}
        <p class="empty-state">Đang tải danh bạ người dùng...</p>
      {:else if usersError}
        <p class="error-msg">{usersError}</p>
      {:else if $userDirectory.length === 0}
        <p class="empty-state">Chưa có người dùng nào khác trong hệ thống.</p>
      {:else}
        <div class="user-list">
          {#each $userDirectory as u}
            <div
              class="user-item"
              role="button"
              tabindex="0"
              on:click={() => startChatWithUser(u)}
              on:keydown={(e) => e.key === 'Enter' && startChatWithUser(u)}
            >
              <div class="avatar-container">
                <img src={u.avatar_url} alt="avatar" class="avatar" />
                <span class="user-status-dot" class:online={$onlineUsers.has(u.id)}></span>
              </div>
              <div>
                <p class="u-name">{u.display_name || u.username}</p>
                <p class="u-sub">@{u.username}</p>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}

<!-- Modal Tạo Nhóm Mới -->
{#if showCreateGroupModal}
  <CreateGroupModal onClose={() => (showCreateGroupModal = false)} />
{/if}

<!-- Modal Tham gia bằng liên kết mời -->
{#if showInviteJoinModal}
  <InviteJoinModal onClose={() => (showInviteJoinModal = false)} />
{/if}

<!-- Modal Quản lý liên kết OAuth -->
{#if showSocialModal}
  <SocialAccountsModal onClose={() => (showSocialModal = false)} />
{/if}

<!-- Modal Xác thực Email -->
{#if showEmailVerifyModal}
  <EmailVerificationModal onClose={() => (showEmailVerifyModal = false)} />
{/if}

<!-- Modal Bảo mật phiên & Thiết bị đăng nhập -->
<SecuritySessionsModal isOpen={showSecurityModal} onClose={() => (showSecurityModal = false)} />

<style>
  .sidebar {
    width: 320px;
    height: 100vh;
    border-right: 1px solid var(--border-glass);
    display: flex;
    flex-direction: column;
    background: var(--bg-surface);
  }

  .sidebar-header {
    padding: 16px 20px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--border-glass);
  }

  .unverified-banner {
    background: rgba(245, 158, 11, 0.14);
    border-bottom: 1px solid rgba(245, 158, 11, 0.3);
    color: #fde68a;
    padding: 7px 16px;
    font-size: 12px;
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    transition: 0.2s;
  }
  .unverified-banner:hover {
    background: rgba(245, 158, 11, 0.22);
  }
  .unverified-banner u {
    font-weight: 600;
    color: #fbbf24;
  }

  .user-info { display: flex; align-items: center; gap: 12px; }
  .my-avatar { width: 40px; height: 40px; border-radius: 50%; }
  .user-info h4 { font-size: 15px; font-weight: 600; color: #fff; }
  .online-tag { font-size: 11px; color: var(--online-color); }

  .header-actions { display: flex; gap: 6px; }
  .icon-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    border-radius: 50%;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .icon-btn:hover { background: rgba(255, 255, 255, 0.1); color: #fff; }

  .search-box {
    margin: 12px 16px 6px 16px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    padding: 8px 12px;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  :global(.search-icon) { color: var(--text-dim); }
  .search-box input {
    background: transparent;
    border: none;
    color: #fff;
    font-size: 13px;
    width: 100%;
    outline: none;
  }

  /* Tabs Switcher */
  .sidebar-tabs {
    display: flex;
    gap: 4px;
    padding: 4px 16px 8px 16px;
    border-bottom: 1px solid var(--border-glass);
  }
  .tab-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 6px 10px;
    border-radius: var(--radius-sm);
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.2s;
  }
  .tab-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.05);
  }
  .tab-btn.active {
    background: rgba(167, 139, 250, 0.18);
    color: #c4b5fd;
    border: 1px solid rgba(167, 139, 250, 0.35);
  }

  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .conversation-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: 0.2s;
  }
  .conversation-item:hover { background: rgba(255, 255, 255, 0.05); }
  .conversation-item.active { background: rgba(99, 102, 241, 0.2); border: 1px solid rgba(99, 102, 241, 0.4); }

  .saved-messages-item {
    background: rgba(56, 189, 248, 0.08);
    border: 1px dashed rgba(56, 189, 248, 0.3);
    margin-bottom: 8px;
  }
  .saved-messages-item:hover {
    background: rgba(56, 189, 248, 0.15);
  }
  .saved-avatar-wrap {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: linear-gradient(135deg, #0284c7, #38bdf8);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .saved-icon {
    color: #fff;
  }
  .saved-name {
    color: #38bdf8;
    font-weight: 600;
  }
  .saved-snippet {
    color: #94a3b8;
    font-size: 11.5px;
  }
  .draft-tag {
    color: #f43f5e;
    font-weight: 600;
  }

  .avatar-container { position: relative; }
  .avatar { width: 44px; height: 44px; border-radius: 50%; }
  .group-avatar-shape { border-radius: 12px; }
  .user-status-dot {
    position: absolute;
    bottom: 2px;
    right: 2px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--offline-color);
    border: 2px solid var(--bg-surface);
  }
  .user-status-dot.online { background: var(--online-color); }

  .content { flex: 1; overflow: hidden; }
  .top-line { display: flex; justify-content: space-between; margin-bottom: 4px; }
  .name { font-size: 14px; font-weight: 500; color: #fff; }
  .group-title { color: #e0e7ff; font-weight: 600; }
  .time { font-size: 11px; color: var(--text-dim); }
  .bottom-line { display: flex; justify-content: space-between; align-items: center; }
  .snippet {
    font-size: 12px;
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 180px;
  }
  .member-tag { color: #a78bfa; font-size: 11px; }
  .unread-badge {
    background: var(--accent-gradient);
    color: #fff;
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 10px;
    font-weight: 700;
  }
  .group-unread-badge {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }
  .tab-badge {
    background: #ef4444;
    color: #fff;
    font-size: 10px;
    font-weight: 700;
    padding: 1px 6px;
    border-radius: 99px;
    margin-left: 4px;
    animation: badgePulse 2.5s infinite;
  }
  .badge-and-menu {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-left: auto;
  }
  .item-more-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: var(--text-dim);
    opacity: 0;
    transition: all 0.2s;
    cursor: pointer;
    padding: 2px;
    border-radius: 4px;
  }
  .conversation-item:hover .item-more-btn {
    opacity: 1;
  }
  .item-more-btn:hover {
    color: var(--text-main);
    background: rgba(255, 255, 255, 0.12);
  }
  .item-context-menu {
    position: absolute;
    right: 14px;
    bottom: -32px;
    z-index: 99;
    background: rgba(17, 24, 39, 0.96);
    backdrop-filter: blur(14px);
    border: 1px solid rgba(139, 92, 246, 0.35);
    border-radius: 8px;
    padding: 4px;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
    display: flex;
    flex-direction: column;
    min-width: 175px;
    animation: menuFadeIn 0.15s ease-out;
  }
  @keyframes menuFadeIn {
    from { opacity: 0; transform: translateY(-4px); }
    to { opacity: 1; transform: translateY(0); }
  }
  .context-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 7px 10px;
    border: none;
    background: transparent;
    color: var(--text-main);
    font-size: 12px;
    border-radius: 6px;
    cursor: pointer;
    text-align: left;
    width: 100%;
    transition: background 0.15s;
  }
  .context-item:hover {
    background: rgba(139, 92, 246, 0.25);
    color: #fff;
  }
  .animate-pulse {
    animation: badgePulse 2s infinite;
  }
  @keyframes badgePulse {
    0%, 100% { transform: scale(1); }
    50% { transform: scale(1.08); }
  }

  .empty-state {
    padding: 30px 20px;
    text-align: center;
    color: var(--text-dim);
    font-size: 13px;
    line-height: 1.5;
  }
  .empty-group-btns {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: 14px;
    width: 100%;
    max-width: 200px;
    align-items: center;
  }
  .create-group-prompt-btn {
    width: 100%;
    background: var(--accent-gradient);
    border: none;
    color: #fff;
    padding: 7px 14px;
    border-radius: var(--radius-sm);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
  }
  .join-link-prompt-btn {
    width: 100%;
    background: rgba(167, 139, 250, 0.15);
    border: 1px solid rgba(167, 139, 250, 0.3);
    color: #c4b5fd;
    padding: 7px 14px;
    border-radius: var(--radius-sm);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.2s;
  }
  .join-link-prompt-btn:hover {
    background: rgba(167, 139, 250, 0.25);
    color: #fff;
  }

  /* Modal Overlay */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  .modal-box {
    width: 380px;
    background: var(--bg-surface);
    border-radius: var(--radius-lg);
    padding: 20px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }
  .modal-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
  .modal-top h3 { font-size: 16px; color: #fff; }
  .close-btn { background: transparent; border: none; color: var(--text-dim); cursor: pointer; font-size: 16px; }
  .user-list { overflow-y: auto; display: flex; flex-direction: column; gap: 8px; }
  .user-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px;
    border-radius: var(--radius-sm);
    cursor: pointer;
  }
  .user-item:hover { background: rgba(255, 255, 255, 0.05); }
  .u-name { font-size: 14px; color: #fff; font-weight: 500; }
  .u-sub { font-size: 12px; color: var(--text-dim); }
  .error-msg { color: #f87171; font-size: 13px; text-align: center; }
</style>
