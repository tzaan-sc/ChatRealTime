<script>
  import { X, Search, Share2, Check, User, Users } from 'lucide-svelte';
  import { conversations, userGroups } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import { wsService } from '../services/websocket';

  export let messageToForward = null;
  export let onClose = () => {};

  let searchQuery = '';
  let withOriginalSender = true;
  let activeTab = 'chats'; // 'chats' hoặc 'groups'
  let isSent = false;
  let sentTargetName = '';

  $: filteredConversations = $conversations.filter((c) => {
    const name = (c.other_user?.display_name || c.other_user?.username || '').toLowerCase();
    return name.includes(searchQuery.toLowerCase());
  });

  $: filteredGroups = $userGroups.filter((g) => {
    return (g.name || '').toLowerCase().includes(searchQuery.toLowerCase());
  });

  function forwardTo(target, isGroup) {
    if (!messageToForward) return;

    let forwardSnippet = null;
    if (withOriginalSender) {
      forwardSnippet = {
        original_sender_id: messageToForward.sender_id,
        original_sender_name: messageToForward.sender_name || 'Người dùng',
        original_message_id: messageToForward.id,
        is_anonymous: false
      };
    } else {
      forwardSnippet = {
        is_anonymous: true
      };
    }

    if (isGroup) {
      wsService.send('group:send', {
        group_id: target.id,
        content: messageToForward.content,
        type: messageToForward.type || 'text',
        file_name: messageToForward.file_name,
        file_size: messageToForward.file_size,
        forward_from: forwardSnippet
      });
      sentTargetName = target.name;
    } else {
      wsService.send('chat:send', {
        receiver_id: target.other_user.id,
        content: messageToForward.content,
        type: messageToForward.type || 'text',
        file_name: messageToForward.file_name,
        file_size: messageToForward.file_size,
        forward_from: forwardSnippet
      });
      sentTargetName = target.other_user.display_name || target.other_user.username;
    }

    isSent = true;
    setTimeout(() => {
      onClose();
    }, 1200);
  }
</script>

<div class="modal-backdrop" on:click|self={onClose}>
  <div class="modal-box glass-card">
    <div class="modal-header">
      <div class="title-wrap">
        <Share2 size={18} class="header-icon" />
        <h3>Chuyển tiếp tin nhắn</h3>
      </div>
      <button class="close-btn" on:click={onClose}>
        <X size={18} />
      </button>
    </div>

    <!-- Xem trước tin chuyển tiếp -->
    <div class="preview-box">
      <span class="preview-label">Nội dung chuyển tiếp:</span>
      <p class="preview-content">{messageToForward?.content || '[Tệp đính kèm]'}</p>
    </div>

    <!-- Tùy chọn nguồn gửi -->
    <div class="forward-options">
      <!-- svelte-ignore a11y_label_has_associated_control -->
      <label class="toggle-option">
        <input type="checkbox" bind:checked={withOriginalSender} />
        <span>Kèm tên người gửi gốc (Ghi nhận nguồn)</span>
      </label>
      <span class="option-hint">
        {withOriginalSender ? 'Người nhận sẽ thấy tin nhắn được chuyển tiếp từ ' + (messageToForward?.sender_name || 'người gửi') : 'Chuyển tiếp ẩn danh (không hiện tên người gửi ban đầu)'}
      </span>
    </div>

    <!-- Tìm kiếm người nhận -->
    <div class="search-box">
      <Search size={16} class="search-icon" />
      <input
        type="text"
        placeholder="Tìm bạn bè hoặc nhóm..."
        bind:value={searchQuery}
      />
    </div>

    <!-- Tabs chuyển danh sách -->
    <div class="tabs-row">
      <button
        class="tab-btn"
        class:active={activeTab === 'chats'}
        on:click={() => (activeTab = 'chats')}
      >
        <User size={14} /> Bạn bè ({filteredConversations.length})
      </button>
      <button
        class="tab-btn"
        class:active={activeTab === 'groups'}
        on:click={() => (activeTab = 'groups')}
      >
        <Users size={14} /> Nhóm ({filteredGroups.length})
      </button>
    </div>

    <!-- Danh sách đích gửi -->
    <div class="targets-list">
      {#if isSent}
        <div class="sent-success">
          <Check size={28} class="check-icon" />
          <p>Đã chuyển tiếp tới <strong>{sentTargetName}</strong> thành công!</p>
        </div>
      {:else if activeTab === 'chats'}
        {#if filteredConversations.length === 0}
          <div class="empty-list">Không tìm thấy cuộc trò chuyện nào</div>
        {:else}
          {#each filteredConversations as c}
            <button class="target-item" on:click={() => forwardTo(c, false)}>
              <img
                src={c.other_user?.avatar_url || `https://api.dicebear.com/7.x/identicon/svg?seed=${c.other_user?.id}`}
                alt="avatar"
                class="target-avatar"
              />
              <div class="target-info">
                <span class="target-name">{c.other_user?.display_name || c.other_user?.username}</span>
                <span class="target-sub">@{c.other_user?.username}</span>
              </div>
              <span class="send-badge">Gửi</span>
            </button>
          {/each}
        {/if}
      {:else}
        {#if filteredGroups.length === 0}
          <div class="empty-list">Không tìm thấy nhóm nào</div>
        {:else}
          {#each filteredGroups as g}
            <button class="target-item" on:click={() => forwardTo(g, true)}>
              <img
                src={g.avatar || `https://api.dicebear.com/7.x/identicon/svg?seed=${g.id}`}
                alt="avatar"
                class="target-avatar"
              />
              <div class="target-info">
                <span class="target-name">{g.name}</span>
                <span class="target-sub">Nhóm chat</span>
              </div>
              <span class="send-badge">Gửi</span>
            </button>
          {/each}
        {/if}
      {/if}
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    backdrop-filter: blur(4px);
  }

  .modal-box {
    width: 440px;
    max-width: 92vw;
    max-height: 85vh;
    background: #0f172a;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 18px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .title-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .header-icon {
    color: #38bdf8;
  }

  .title-wrap h3 {
    margin: 0;
    font-size: 16px;
    color: #fff;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
  }

  .close-btn:hover {
    color: #fff;
  }

  .preview-box {
    margin: 12px 16px 8px;
    padding: 10px;
    background: rgba(255, 255, 255, 0.04);
    border-left: 3px solid #38bdf8;
    border-radius: 4px;
  }

  .preview-label {
    font-size: 11px;
    color: #94a3b8;
    display: block;
    margin-bottom: 3px;
  }

  .preview-content {
    margin: 0;
    font-size: 13px;
    color: #e2e8f0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .forward-options {
    padding: 8px 16px;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .toggle-option {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: #fff;
    cursor: pointer;
  }

  .toggle-option input {
    accent-color: #6366f1;
    width: 16px;
    height: 16px;
  }

  .option-hint {
    font-size: 11px;
    color: #94a3b8;
    margin-left: 24px;
  }

  .search-box {
    margin: 8px 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
  }

  .search-icon {
    color: #64748b;
  }

  .search-box input {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    font-size: 13px;
    outline: none;
  }

  .tabs-row {
    display: flex;
    margin: 0 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .tab-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 8px;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: #94a3b8;
    font-size: 12.5px;
    cursor: pointer;
    transition: 0.15s;
  }

  .tab-btn:hover {
    color: #fff;
  }

  .tab-btn.active {
    color: #38bdf8;
    border-bottom-color: #38bdf8;
    font-weight: 600;
  }

  .targets-list {
    flex: 1;
    overflow-y: auto;
    padding: 10px 16px;
    min-height: 180px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .target-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 8px;
    cursor: pointer;
    text-align: left;
    transition: 0.15s;
    width: 100%;
  }

  .target-item:hover {
    background: rgba(255, 255, 255, 0.06);
    border-color: rgba(255, 255, 255, 0.1);
  }

  .target-avatar {
    width: 34px;
    height: 34px;
    border-radius: 50%;
    object-fit: cover;
  }

  .target-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .target-name {
    font-size: 13px;
    color: #fff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .target-sub {
    font-size: 11px;
    color: #64748b;
  }

  .send-badge {
    padding: 4px 10px;
    background: rgba(99, 102, 241, 0.15);
    color: #818cf8;
    border-radius: 6px;
    font-size: 11.5px;
    font-weight: 600;
  }

  .target-item:hover .send-badge {
    background: #6366f1;
    color: #fff;
  }

  .empty-list {
    text-align: center;
    color: #64748b;
    font-size: 13px;
    margin-top: 40px;
  }

  .sent-success {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    margin: 40px 0;
    color: #10b981;
    font-size: 14px;
  }

  .check-icon {
    color: #10b981;
  }
</style>
