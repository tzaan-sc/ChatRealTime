<script>
  import { activeConversation, currentMessages, appendMessage } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import { wsService } from '../services/websocket';
  import { onMount, afterUpdate } from 'svelte';
  import { Send, Smile, Paperclip } from 'lucide-svelte';

  let inputContent = '';
  let messagesContainer;

  function scrollToBottom() {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  }

  afterUpdate(() => {
    scrollToBottom();
  });

  function handleSend() {
    if (!inputContent.trim() || !$activeConversation) return;

    const receiverID = $activeConversation.other_user.id;
    wsService.send('chat:send', {
      receiver_id: receiverID,
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
        <img src={$activeConversation.other_user.avatar_url} alt="avatar" class="header-avatar" />
        <div>
          <h3>{$activeConversation.other_user.display_name || $activeConversation.other_user.username}</h3>
          <span class="status-sub">Sẵn sàng nhận tin</span>
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
            <span class="timestamp">
              {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          </div>
        </div>
      {/each}
    </div>

    <!-- Input Footer -->
    <div class="chat-footer glass-card">
      <div class="input-wrapper">
        <textarea
          rows="1"
          placeholder="Nhập tin nhắn... (Nhấn Enter để gửi)"
          bind:value={inputContent}
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
  .header-avatar { width: 44px; height: 44px; border-radius: 50%; }
  .partner-info h3 { font-size: 16px; font-weight: 600; color: #fff; }
  .status-sub { font-size: 12px; color: var(--text-dim); }
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
  .timestamp {
    display: block;
    font-size: 10px;
    opacity: 0.7;
    margin-top: 4px;
    text-align: right;
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
