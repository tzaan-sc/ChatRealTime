<script>
  import { Pin, ChevronLeft, ChevronRight, X, List } from 'lucide-svelte';

  export let pinnedMessages = [];
  export let onScrollTo = (id) => {};
  export let onUnpin = (id) => {};

  let currentIndex = 0;
  let showAllModal = false;

  $: if (currentIndex >= pinnedMessages.length && pinnedMessages.length > 0) {
    currentIndex = pinnedMessages.length - 1;
  }

  $: currentPin = pinnedMessages[currentIndex] || null;

  function prevPin() {
    if (currentIndex > 0) {
      currentIndex--;
    } else {
      currentIndex = pinnedMessages.length - 1;
    }
  }

  function nextPin() {
    if (currentIndex < pinnedMessages.length - 1) {
      currentIndex++;
    } else {
      currentIndex = 0;
    }
  }

  function getSnippet(msg) {
    if (!msg) return '';
    if (msg.type === 'image') return '🖼️ [Hình ảnh]';
    if (msg.type === 'voice') return '🎙️ [Tin nhắn thoại]';
    if (msg.type === 'file') return `📎 [Tệp] ${msg.file_name || ''}`;
    return msg.content || '';
  }
</script>

{#if pinnedMessages.length > 0 && currentPin}
  <div class="pinned-bar glass-card">
    <div class="pin-icon-wrap">
      <Pin size={16} class="pin-icon" />
    </div>

    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="pin-content"
      role="button"
      tabindex="0"
      on:click={() => onScrollTo(currentPin.id)}
      title="Bấm để cuộn tới tin nhắn này"
    >
      <div class="pin-meta">
        <span class="pin-title">Tin nhắn đã ghim</span>
        {#if pinnedMessages.length > 1}
          <span class="pin-count">({currentIndex + 1}/{pinnedMessages.length})</span>
        {/if}
        <span class="pin-author">• {currentPin.sender_name || 'Người gửi'}</span>
      </div>
      <div class="pin-text">{getSnippet(currentPin)}</div>
    </div>

    <div class="pin-actions">
      {#if pinnedMessages.length > 1}
        <button class="pin-nav-btn" on:click={prevPin} title="Tin ghim trước">
          <ChevronLeft size={16} />
        </button>
        <button class="pin-nav-btn" on:click={nextPin} title="Tin ghim kế tiếp">
          <ChevronRight size={16} />
        </button>
        <button class="pin-nav-btn" on:click={() => (showAllModal = true)} title="Xem tất cả tin ghim">
          <List size={16} />
        </button>
      {/if}

      <button class="pin-nav-btn unpin-btn" on:click={() => onUnpin(currentPin.id)} title="Gỡ ghim tin nhắn này">
        <X size={15} />
      </button>
    </div>
  </div>

  <!-- Modal Xem Toàn Bộ Tin Đã Ghim -->
  {#if showAllModal}
    <div class="modal-overlay" on:click|self={() => (showAllModal = false)}>
      <div class="pinned-modal glass-card">
        <div class="modal-header">
          <div class="modal-title-row">
            <Pin size={18} class="pin-icon" />
            <h3>Danh sách tin nhắn đã ghim ({pinnedMessages.length})</h3>
          </div>
          <button class="close-btn" on:click={() => (showAllModal = false)}>
            <X size={18} />
          </button>
        </div>

        <div class="pinned-list">
          {#each pinnedMessages as msg, idx}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <div
              class="pinned-item"
              role="button"
              tabindex="0"
              on:click={() => {
                showAllModal = false;
                onScrollTo(msg.id);
              }}
            >
              <div class="pinned-item-header">
                <strong>{msg.sender_name || 'Thành viên'}</strong>
                <span class="pinned-date">
                  {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
              <p class="pinned-item-content">{getSnippet(msg)}</p>
              <div class="pinned-item-footer">
                <button
                  class="item-unpin-btn"
                  on:click|stopPropagation={() => onUnpin(msg.id)}
                >
                  Gỡ ghim
                </button>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
{/if}

<style>
  .pinned-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 16px;
    background: rgba(15, 23, 42, 0.75);
    border-bottom: 1px solid var(--border-glass, rgba(255, 255, 255, 0.08));
    backdrop-filter: blur(8px);
    z-index: 10;
  }

  .pin-icon-wrap {
    color: #f59e0b;
    display: flex;
    align-items: center;
  }

  .pin-content {
    flex: 1;
    min-width: 0;
    cursor: pointer;
    text-align: left;
  }

  .pin-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: #f59e0b;
    font-weight: 600;
  }

  .pin-author {
    color: var(--text-dim, #94a3b8);
    font-weight: 400;
  }

  .pin-text {
    font-size: 13px;
    color: #fff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .pin-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .pin-nav-btn {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: var(--text-muted, #94a3b8);
    width: 28px;
    height: 28px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: 0.15s;
  }

  .pin-nav-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.15);
  }

  .unpin-btn:hover {
    color: #ef4444;
    background: rgba(239, 68, 68, 0.15);
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    backdrop-filter: blur(4px);
  }

  .pinned-modal {
    width: 480px;
    max-width: 90vw;
    max-height: 80vh;
    background: #0f172a;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .modal-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .modal-title-row h3 {
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

  .pinned-list {
    padding: 16px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .pinned-item {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 12px;
    cursor: pointer;
    transition: 0.15s;
    text-align: left;
  }

  .pinned-item:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(245, 158, 11, 0.4);
  }

  .pinned-item-header {
    display: flex;
    justify-content: space-between;
    font-size: 13px;
    color: #a78bfa;
    margin-bottom: 6px;
  }

  .pinned-date {
    font-size: 11px;
    color: #64748b;
  }

  .pinned-item-content {
    margin: 0 0 8px 0;
    font-size: 13.5px;
    color: #e2e8f0;
    word-break: break-word;
  }

  .pinned-item-footer {
    display: flex;
    justify-content: flex-end;
  }

  .item-unpin-btn {
    background: transparent;
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #ef4444;
    font-size: 11px;
    padding: 3px 8px;
    border-radius: 4px;
    cursor: pointer;
    transition: 0.15s;
  }

  .item-unpin-btn:hover {
    background: rgba(239, 68, 68, 0.15);
  }
</style>
