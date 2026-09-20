<script>
  import { activeConversation, currentMessages, onlineUsers, typingUsers } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import { wsService } from '../services/websocket';
  import { afterUpdate } from 'svelte';
  import { Send } from 'lucide-svelte';

  let inputContent = '';
  let messagesContainer;
  let typingTimeout;
  let isTyping = false;

  // Kiểm tra đối phương có online không
  $: partnerID = $activeConversation?.other_user?.id;
  $: isPartnerOnline = partnerID ? $onlineUsers.has(partnerID) : false;
  $: isPartnerTyping = partnerID ? $typingUsers.has(partnerID) : false;

  // Tự động cuộn xuống đáy khi có tin nhắn mới
  afterUpdate(() => {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  });

  // Bắt sự kiện gõ phím để bắn Typing Indicator
  function handleInput() {
    if (!$activeConversation || !partnerID) return;

    if (!isTyping) {
      isTyping = true;
      wsService.send('typing:start', { receiver_id: partnerID });
    }

    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => {
      isTyping = false;
      wsService.send('typing:stop', { receiver_id: partnerID });
    }, 2000);
  }

  // Gửi tin nhắn
  function handleSend() {
    if (!inputContent.trim() || !$activeConversation || !partnerID) return;

    // Dừng typing khi gửi tin
    clearTimeout(typingTimeout);
    if (isTyping) {
      isTyping = false;
      wsService.send('typing:stop', { receiver_id: partnerID });
    }

    wsService.send('chat:send', {
      receiver_id: partnerID,
      content: inputContent.trim(),
      type: 'text'
    });

    inputContent = '';
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  // Gửi event xác nhận đã đọc khi đang mở xem khung chat
  $: if ($activeConversation && partnerID) {
    wsService.send('chat:read', {
      conversation_id: $activeConversation.conversation.custom_id,
      partner_id: partnerID
    });
  }
</script>

{#if !$activeConversation}
  <div class="empty-chat glass-card">
    <div class="empty-content">
      <div class="icon-circle">💬</div>
      <h2>Chào mừng đến với Realtime Chat!</h2>
      <p>Chọn một người bạn ở danh sách bên trái để bắt đầu cuộc trò chuyện.</p>
    </div>
  </div>
{:else}
  <main class="chat-area">
    <!-- Header -->
    <div class="chat-header glass-card">
      <div class="partner-info">
        <div class="avatar-wrap">
          <img src={$activeConversation.other_user.avatar_url} alt="avatar" class="header-avatar" />
          <span class="status-dot" class:online={isPartnerOnline}></span>
        </div>
        <div>
          <h3>{$activeConversation.other_user.display_name || $activeConversation.other_user.username}</h3>
          <span class="status-sub">
            {#if isPartnerTyping}
              <span class="typing-text">đang soạn tin...</span>
            {:else if isPartnerOnline}
              <span class="online-text">Đang hoạt động</span>
            {:else}
              <span class="offline-text">Ngoại tuyến</span>
            {/if}
          </span>
        </div>
      </div>
    </div>

    <!-- Messages Body -->
    <div class="messages-viewport" bind:this={messagesContainer}>
      {#each $currentMessages as msg}
        {@const isMe = msg.sender_id === $currentUser.id}
        <div class="message-row" class:me={isMe}>
          <div class="bubble" class:bubble-me={isMe} class:bubble-other={!isMe}>
            <p class="text">{msg.content}</p>
            <div class="msg-meta">
              <span class="timestamp">
                {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
              {#if isMe}
                <span class="receipt-icon" class:seen={msg.is_read} title={msg.is_read ? 'Đã xem' : 'Đã gửi'}>
                  {msg.is_read ? '✓✓' : '✓'}
                </span>
              {/if}
            </div>
          </div>
        </div>
      {/each}

      <!-- Animation Đang Gõ Phím (Typing Dots) -->
      {#if isPartnerTyping}
        <div class="message-row">
          <div class="typing-bubble glass-card">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
          </div>
        </div>
      {/if}
    </div>

    <!-- Input Footer -->
    <div class="chat-footer glass-card">
      <div class="input-wrapper">
        <textarea
          rows="1"
          placeholder="Nhập tin nhắn... (Nhấn Enter để gửi)"
          bind:value={inputContent}
          on:input={handleInput}
          on:keydown={handleKeyDown}
        ></textarea>
        <button class="send-btn" on:click={handleSend} disabled={!inputContent.trim()}>
          <Send size={18} />
        </button>
      </div>
    </div>
  </main>
{/if}

<style>
  .chat-area { flex: 1; height: 100vh; display: flex; flex-direction: column; background: var(--bg-primary); }
  .empty-chat { flex: 1; display: flex; align-items: center; justify-content: center; text-align: center; }
  .icon-circle { font-size: 50px; margin-bottom: 16px; }
  .empty-content h2 { font-size: 22px; margin-bottom: 8px; }
  .empty-content p { color: var(--text-muted); font-size: 14px; }
  .chat-header {
    padding: 16px 24px;
    border-bottom: 1px solid var(--border-glass);
    display: flex;
    align-items: center;
  }
  .partner-info { display: flex; align-items: center; gap: 14px; }
  .avatar-wrap { position: relative; }
  .header-avatar { width: 44px; height: 44px; border-radius: 50%; }
  .status-dot {
    position: absolute;
    bottom: 2px;
    right: 2px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--offline-color);
    border: 2px solid var(--bg-primary);
  }
  .status-dot.online { background: var(--online-color); }
  .partner-info h3 { font-size: 16px; font-weight: 600; color: #fff; }
  .status-sub { font-size: 12px; }
  .online-text { color: var(--online-color); }
  .offline-text { color: var(--text-dim); }
  .typing-text { color: #a78bfa; font-style: italic; animation: pulse 1.5s infinite; }
  @keyframes pulse { 0%, 100% { opacity: 0.6; } 50% { opacity: 1; } }

  .messages-viewport {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .message-row { display: flex; width: 100%; }
  .message-row.me { justify-content: flex-end; }
  .bubble {
    max-width: 65%;
    padding: 12px 16px;
    border-radius: var(--radius-md);
    position: relative;
    word-break: break-word;
  }
  .bubble-me {
    background: var(--bubble-me);
    color: #fff;
    border-bottom-right-radius: 4px;
  }
  .bubble-other {
    background: var(--bubble-other);
    color: #f3f4f6;
    border-bottom-left-radius: 4px;
    border: 1px solid var(--border-glass);
  }
  .text { font-size: 14px; line-height: 1.5; }
  .msg-meta {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    margin-top: 4px;
  }
  .timestamp { font-size: 10px; opacity: 0.75; }
  .receipt-icon { font-size: 11px; font-weight: 700; opacity: 0.6; }
  .receipt-icon.seen { color: #38bdf8; opacity: 1; }

  /* Typing Indicator Animation */
  .typing-bubble {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 10px 16px;
    border-radius: var(--radius-md);
    border-bottom-left-radius: 4px;
  }
  .typing-dot {
    width: 6px;
    height: 6px;
    background: #a78bfa;
    border-radius: 50%;
    animation: typingBounce 1.4s infinite ease-in-out both;
  }
  .typing-dot:nth-child(1) { animation-delay: -0.32s; }
  .typing-dot:nth-child(2) { animation-delay: -0.16s; }
  @keyframes typingBounce {
    0%, 80%, 100% { transform: scale(0); }
    40% { transform: scale(1); }
  }

  .chat-footer { padding: 16px 24px; border-top: 1px solid var(--border-glass); }
  .input-wrapper {
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 8px 16px;
  }
  .input-wrapper textarea {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    font-size: 14px;
    outline: none;
    resize: none;
    font-family: inherit;
    max-height: 100px;
  }
  .send-btn {
    background: var(--accent-gradient);
    border: none;
    color: #fff;
    width: 38px;
    height: 38px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .send-btn:not(:disabled):hover { transform: scale(1.05); }
</style>
