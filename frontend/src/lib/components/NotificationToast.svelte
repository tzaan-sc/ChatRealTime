<script>
  import { activeToasts, removeToast } from '../stores/notification';
  import { conversations, selectConversation, userGroups, selectGroup } from '../stores/chat';
  import { X, MessageSquare, ArrowRight } from 'lucide-svelte';

  function handleClickToast(toast) {
    if (toast.conversationId) {
      const conv = $conversations.find((c) => c.conversation?.custom_id === toast.conversationId);
      if (conv) {
        selectConversation(conv);
      }
    } else if (toast.groupId) {
      const grp = $userGroups.find((g) => g.id === toast.groupId);
      if (grp) {
        selectGroup(grp);
      }
    }
    removeToast(toast.id);
  }
</script>

{#if $activeToasts.length > 0}
  <div class="toast-container">
    {#each $activeToasts as toast (toast.id)}
      <div class="toast-card glass-toast">
        <div class="toast-avatar-box">
          {#if toast.senderAvatar}
            <img src={toast.senderAvatar} alt="avatar" class="toast-avatar" />
          {:else}
            <div class="toast-avatar-fallback">
              <MessageSquare size={16} />
            </div>
          {/if}
        </div>

        <div class="toast-body" role="button" tabindex="0" on:click={() => handleClickToast(toast)} on:keydown={(e) => e.key === 'Enter' && handleClickToast(toast)}>
          <div class="toast-header">
            <span class="toast-title">{toast.title || 'Tin nhắn mới'}</span>
            <span class="toast-sender">{toast.senderName}</span>
          </div>
          <p class="toast-content">{toast.content}</p>
        </div>

        <div class="toast-actions">
          <button class="toast-action-btn" title="Xem tin nhắn" on:click={() => handleClickToast(toast)}>
            <span>Xem</span>
            <ArrowRight size={13} />
          </button>
          <button class="toast-close-btn" title="Đóng" on:click|stopPropagation={() => removeToast(toast.id)}>
            <X size={15} />
          </button>
        </div>

        <div class="toast-progress-bar"></div>
      </div>
    {/each}
  </div>
{/if}

<style>
  .toast-container {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 10px;
    pointer-events: none;
  }

  .toast-card {
    pointer-events: auto;
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 320px;
    max-width: 420px;
    padding: 12px 14px;
    border-radius: 14px;
    overflow: hidden;
    animation: toastSlideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4), 0 0 0 1px rgba(255, 255, 255, 0.1);
  }

  .glass-toast {
    background: rgba(17, 24, 39, 0.92);
    backdrop-filter: blur(16px);
    border: 1px solid rgba(139, 92, 246, 0.25);
  }

  .glass-toast:hover {
    border-color: rgba(139, 92, 246, 0.5);
  }

  @keyframes toastSlideIn {
    from {
      opacity: 0;
      transform: translateX(60px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateX(0) scale(1);
    }
  }

  .toast-avatar-box {
    flex-shrink: 0;
  }

  .toast-avatar {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid var(--accent-primary);
  }

  .toast-avatar-fallback {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    background: rgba(139, 92, 246, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--accent-primary);
  }

  .toast-body {
    flex: 1;
    min-width: 0;
    cursor: pointer;
  }

  .toast-header {
    display: flex;
    align-items: baseline;
    gap: 6px;
    margin-bottom: 2px;
  }

  .toast-title {
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--accent-primary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .toast-sender {
    font-size: 0.84rem;
    font-weight: 600;
    color: var(--text-main);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .toast-content {
    font-size: 0.8rem;
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: 1.3;
  }

  .toast-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .toast-action-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 10px;
    border-radius: 8px;
    border: none;
    background: var(--accent-gradient);
    color: #fff;
    font-size: 0.76rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .toast-action-btn:hover {
    opacity: 0.9;
    transform: translateY(-1px);
  }

  .toast-close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border-radius: 50%;
    border: none;
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.2s;
  }

  .toast-close-btn:hover {
    background: rgba(255, 255, 255, 0.18);
    color: var(--text-main);
  }

  .toast-progress-bar {
    position: absolute;
    bottom: 0;
    left: 0;
    height: 2.5px;
    width: 100%;
    background: var(--accent-gradient);
    animation: toastProgress 4.5s linear forwards;
  }

  @keyframes toastProgress {
    from {
      width: 100%;
    }
    to {
      width: 0%;
    }
  }
</style>
