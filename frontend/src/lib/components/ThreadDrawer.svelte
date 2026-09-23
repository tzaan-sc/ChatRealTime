<script>
  import { onMount, onDestroy } from 'svelte';
  import { X, Send, MessageSquare } from 'lucide-svelte';
  import { apiRequest } from '../services/api';
  import { wsService } from '../services/websocket';
  import { currentUser, token } from '../stores/auth';
  import MarkdownRenderer from './MarkdownRenderer.svelte';

  export let rootMessage = null;
  export let conversationID = '';
  export let groupID = '';
  export let partnerID = '';
  export let onClose = () => {};

  let threadMessages = [];
  let threadInput = '';
  let isLoading = true;
  let repliesContainer;

  $: isGroup = !!groupID;

  async function loadThreadReplies() {
    if (!rootMessage?.id) return;
    isLoading = true;
    try {
      const res = await apiRequest(`/chat/messages/${rootMessage.id}/thread`, 'GET', null, $token);
      threadMessages = res.data || [];
    } catch (e) {
      console.error('Lỗi tải luồng tin nhắn:', e);
    } finally {
      isLoading = false;
      scrollToBottom();
    }
  }

  function scrollToBottom() {
    setTimeout(() => {
      if (repliesContainer) {
        repliesContainer.scrollTop = repliesContainer.scrollHeight;
      }
    }, 50);
  }

  function sendThreadReply() {
    if (!threadInput.trim() || !rootMessage) return;

    const content = threadInput.trim();
    if (isGroup) {
      wsService.send('group:send', {
        group_id: groupID,
        content: content,
        type: 'text',
        thread_root_id: rootMessage.id
      });
    } else {
      wsService.send('chat:send', {
        receiver_id: partnerID,
        content: content,
        type: 'text',
        thread_root_id: rootMessage.id
      });
    }

    threadInput = '';
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendThreadReply();
    }
  }

  // Lắng nghe tin nhắn mới trong luồng qua custom event hoặc props
  export function appendThreadReply(msg) {
    if (msg.thread_root_id === rootMessage?.id) {
      threadMessages = [...threadMessages, msg];
      scrollToBottom();
    }
  }

  onMount(() => {
    loadThreadReplies();
  });
</script>

<div class="thread-drawer glass-card">
  <div class="drawer-header">
    <div class="header-title">
      <MessageSquare size={18} class="thread-icon" />
      <div>
        <h3>Luồng thảo luận</h3>
        <span class="sub-title">Phản hồi riêng cho tin nhắn này</span>
      </div>
    </div>
    <button class="close-btn" on:click={onClose} title="Đóng luồng">
      <X size={18} />
    </button>
  </div>

  <!-- Tin Nhắn Gốc (Root Message) -->
  <div class="root-message-card">
    <div class="root-header">
      <div class="author-info">
        {#if rootMessage.sender_avatar}
          <img src={rootMessage.sender_avatar} alt="avatar" class="author-avatar" />
        {/if}
        <span class="author-name">{rootMessage.sender_name || 'Người gửi'}</span>
      </div>
      <span class="root-time">
        {new Date(rootMessage.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
      </span>
    </div>

    <div class="root-content">
      {#if rootMessage.type === 'text'}
        <MarkdownRenderer content={rootMessage.content} />
      {:else if rootMessage.type === 'image'}
        <img src={rootMessage.content} alt="ảnh gốc" class="root-image" />
      {:else}
        <p class="root-fallback">{rootMessage.content}</p>
      {/if}
    </div>
  </div>

  <div class="replies-divider">
    <span>{threadMessages.length} phản hồi</span>
  </div>

  <!-- Danh Sách Phản Hồi Trong Luồng -->
  <div class="replies-viewport" bind:this={repliesContainer}>
    {#if isLoading}
      <div class="thread-loading">Đang tải phản hồi...</div>
    {:else if threadMessages.length === 0}
      <div class="thread-empty">
        <p>Chưa có câu trả lời nào trong luồng này.</p>
        <span>Hãy là người đầu tiên bắt đầu thảo luận!</span>
      </div>
    {:else}
      {#each threadMessages as reply (reply.id || reply.created_at)}
        {@const isMe = reply.sender_id === $currentUser?.id}
        <div class="reply-row" class:me={isMe}>
          <div class="reply-bubble" class:bubble-me={isMe}>
            {#if !isMe}
              <div class="reply-author-row">
                <span class="reply-author">{reply.sender_name || 'Thành viên'}</span>
              </div>
            {/if}
            <div class="reply-text">
              <MarkdownRenderer content={reply.content} />
            </div>
            <span class="reply-time">
              {new Date(reply.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          </div>
        </div>
      {/each}
    {/if}
  </div>

  <!-- Ô Nhập Tin Trong Luồng -->
  <div class="thread-input-area">
    <div class="input-box">
      <input
        type="text"
        placeholder="Trả lời trong luồng... (Enter)"
        bind:value={threadInput}
        on:keydown={handleKeyDown}
      />
      <button
        class="send-btn"
        disabled={!threadInput.trim()}
        on:click={sendThreadReply}
        title="Gửi phản hồi"
      >
        <Send size={16} />
      </button>
    </div>
  </div>
</div>

<style>
  .thread-drawer {
    width: 360px;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #0b1120;
    border-left: 1px solid var(--border-glass, rgba(255, 255, 255, 0.1));
    z-index: 20;
    animation: slideIn 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideIn {
    from { transform: translateX(100%); }
    to { transform: translateX(0); }
  }

  .drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 18px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .thread-icon {
    color: #38bdf8;
  }

  .header-title h3 {
    margin: 0;
    font-size: 15px;
    color: #fff;
  }

  .sub-title {
    font-size: 11px;
    color: #94a3b8;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    border-radius: 6px;
    padding: 4px;
  }

  .close-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .root-message-card {
    margin: 12px;
    padding: 12px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 10px;
  }

  .root-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .author-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .author-avatar {
    width: 22px;
    height: 22px;
    border-radius: 50%;
  }

  .author-name {
    font-size: 12.5px;
    font-weight: 600;
    color: #a78bfa;
  }

  .root-time {
    font-size: 11px;
    color: #64748b;
  }

  .root-image {
    max-width: 100%;
    max-height: 180px;
    border-radius: 6px;
  }

  .replies-divider {
    display: flex;
    align-items: center;
    margin: 0 16px;
    font-size: 11px;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .replies-divider::before,
  .replies-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: rgba(255, 255, 255, 0.08);
  }

  .replies-divider span {
    padding: 0 8px;
  }

  .replies-viewport {
    flex: 1;
    overflow-y: auto;
    padding: 12px 16px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .thread-loading,
  .thread-empty {
    text-align: center;
    color: #64748b;
    font-size: 13px;
    margin-top: 30px;
  }

  .thread-empty span {
    font-size: 11.5px;
    color: #475569;
  }

  .reply-row {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .reply-row.me {
    align-items: flex-end;
  }

  .reply-bubble {
    max-width: 85%;
    padding: 8px 12px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.07);
    color: #f1f5f9;
  }

  .reply-bubble.bubble-me {
    background: linear-gradient(135deg, #6366f1, #8b5cf6);
    color: #fff;
  }

  .reply-author-row {
    margin-bottom: 2px;
  }

  .reply-author {
    font-size: 11px;
    font-weight: 600;
    color: #38bdf8;
  }

  .reply-time {
    display: block;
    font-size: 10px;
    color: rgba(255, 255, 255, 0.5);
    text-align: right;
    margin-top: 3px;
  }

  .thread-input-area {
    padding: 12px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(15, 23, 42, 0.5);
  }

  .input-box {
    display: flex;
    align-items: center;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 4px 8px;
  }

  .input-box input {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    padding: 6px;
    font-size: 13px;
    outline: none;
  }

  .input-box .send-btn {
    background: #6366f1;
    border: none;
    color: #fff;
    width: 30px;
    height: 30px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: 0.15s;
  }

  .input-box .send-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
