<script>
  import { conversations, activeConversation, selectConversation, userDirectory } from '../stores/chat';
  import { currentUser, logout, token } from '../stores/auth';
  import { apiRequest } from '../services/api';
  import { onMount } from 'svelte';
  import { MessageSquarePlus, LogOut, Search } from 'lucide-svelte';

  let showUsersModal = false;
  let searchQuery = '';

  async function openNewChatModal() {
    showUsersModal = true;
    try {
      const res = await apiRequest('/chat/users', 'GET', null, $token);
      userDirectory.set(res.data || []);
    } catch (e) {
      console.error(e);
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
      <button class="icon-btn" title="Cuộc trò chuyện mới" on:click={openNewChatModal}>
        <MessageSquarePlus size={20} />
      </button>
      <button class="icon-btn" title="Đăng xuất" on:click={logout}>
        <LogOut size={20} />
      </button>
    </div>
  </div>

  <!-- Search Bar -->
  <div class="search-box">
    <Search size={16} class="search-icon" />
    <input type="text" placeholder="Tìm kiếm cuộc trò chuyện..." bind:value={searchQuery} />
  </div>

  <!-- Conversation List -->
  <div class="conversation-list">
    {#if $conversations.length === 0}
      <div class="empty-state">Chưa có cuộc trò chuyện nào. Bấm nút dấu cộng để bắt đầu nhắn tin!</div>
    {:else}
      {#each $conversations as item}
        <div
          class="conversation-item"
          class:active={$activeConversation?.conversation?.custom_id === item.conversation.custom_id}
          on:click={() => selectConversation(item)}
        >
          <img src={item.other_user.avatar_url} alt="avatar" class="avatar" />
          <div class="content">
            <div class="top-line">
              <span class="name">{item.other_user.display_name || item.other_user.username}</span>
              <span class="time">
                {item.conversation.last_message_at ? new Date(item.conversation.last_message_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
              </span>
            </div>
            <div class="bottom-line">
              <p class="snippet">{item.conversation.last_message || 'Bắt đầu cuộc trò chuyện'}</p>
              {#if item.unread_count > 0}
                <span class="unread-badge">{item.unread_count}</span>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</aside>

<!-- Modal Chọn Bạn Chat Mới -->
{#if showUsersModal}
  <div class="modal-overlay" on:click={() => (showUsersModal = false)}>
    <div class="modal-box glass-card" on:click|stopPropagation>
      <h3>Bắt đầu cuộc trò chuyện</h3>
      <div class="user-list">
        {#each $userDirectory as u}
          <div class="user-item" on:click={() => startChatWithUser(u)}>
            <img src={u.avatar_url} alt="avatar" class="avatar" />
            <div>
              <p class="u-name">{u.display_name}</p>
              <p class="u-sub">@{u.username}</p>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}

<style>
  .sidebar {
    width: 360px;
    height: 100vh;
    border-right: 1px solid var(--border-glass);
    display: flex;
    flex-direction: column;
  }
  .sidebar-header {
    padding: 18px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-glass);
  }
  .user-info { display: flex; align-items: center; gap: 12px; }
  .my-avatar { width: 44px; height: 44px; border-radius: 50%; border: 2px solid var(--accent-primary); }
  .user-info h4 { font-size: 15px; font-weight: 600; color: #fff; }
  .online-tag { font-size: 11px; color: var(--online-color); }
  .header-actions { display: flex; gap: 8px; }
  .icon-btn {
    background: rgba(255, 255, 255, 0.06);
    border: none;
    color: var(--text-muted);
    width: 36px;
    height: 36px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .icon-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.12); }
  .search-box {
    margin: 14px 20px;
    position: relative;
    display: flex;
    align-items: center;
  }
  :global(.search-icon) {
    position: absolute;
    left: 12px;
    color: var(--text-dim);
  }
  .search-box input {
    width: 100%;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    padding: 10px 14px 10px 36px;
    color: #fff;
    font-size: 13px;
    outline: none;
  }
  .conversation-list { flex: 1; overflow-y: auto; padding: 0 10px 20px; }
  .conversation-item {
    display: flex;
    gap: 12px;
    padding: 12px 14px;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: 0.2s;
    margin-bottom: 4px;
  }
  .conversation-item:hover { background: rgba(255, 255, 255, 0.05); }
  .conversation-item.active { background: rgba(139, 92, 246, 0.15); border: 1px solid rgba(139, 92, 246, 0.3); }
  .avatar { width: 46px; height: 46px; border-radius: 50%; }
  .content { flex: 1; min-width: 0; }
  .top-line { display: flex; justify-content: space-between; margin-bottom: 4px; }
  .name { font-size: 14px; font-weight: 600; color: #fff; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .time { font-size: 11px; color: var(--text-dim); }
  .bottom-line { display: flex; justify-content: space-between; align-items: center; }
  .snippet { font-size: 12px; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 200px; }
  .unread-badge {
    background: var(--accent-primary);
    color: #fff;
    font-size: 11px;
    font-weight: 700;
    padding: 2px 7px;
    border-radius: 99px;
  }
  .empty-state { padding: 30px 20px; text-align: center; color: var(--text-dim); font-size: 13px; }
  .modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.7); display: flex; align-items: center; justify-content: center; z-index: 1000;
  }
  .modal-box { width: 360px; padding: 24px; border-radius: var(--radius-lg); }
  .modal-box h3 { font-size: 16px; margin-bottom: 16px; }
  .user-list { max-height: 300px; overflow-y: auto; }
  .user-item { display: flex; gap: 12px; padding: 10px; border-radius: var(--radius-sm); cursor: pointer; }
  .user-item:hover { background: rgba(255, 255, 255, 0.08); }
  .u-name { font-size: 14px; font-weight: 600; }
  .u-sub { font-size: 12px; color: var(--text-dim); }
</style>
