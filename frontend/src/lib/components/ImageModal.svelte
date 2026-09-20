<script>
  import { onMount, onDestroy } from 'svelte';
  import { X, Download, ExternalLink } from 'lucide-svelte';

  export let imageUrl = '';
  export let fileName = 'image';
  export let onClose = () => {};

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeydown);
    // Khóa cuộn trang khi modal đang mở
    document.body.style.overflow = 'hidden';
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeydown);
    document.body.style.overflow = '';
  });
</script>

{#if imageUrl}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="lightbox-overlay" on:click={onClose} role="dialog" aria-modal="true" tabindex="-1">
    <div class="lightbox-container" on:click={(e) => e.stopPropagation()}>
      <div class="lightbox-header">
        <span class="file-title">{fileName}</span>
        <div class="action-buttons">
          <a
            href={imageUrl}
            download={fileName}
            target="_blank"
            rel="noopener noreferrer"
            class="action-btn"
            title="Tải ảnh về máy"
          >
            <Download size={18} />
          </a>
          <a
            href={imageUrl}
            target="_blank"
            rel="noopener noreferrer"
            class="action-btn"
            title="Mở tab mới"
          >
            <ExternalLink size={18} />
          </a>
          <button class="action-btn close-btn" on:click={onClose} title="Đóng (Esc)">
            <X size={20} />
          </button>
        </div>
      </div>

      <div class="image-wrapper">
        <img src={imageUrl} alt={fileName} class="lightbox-img" />
      </div>
    </div>
  </div>
{/if}

<style>
  .lightbox-overlay {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 20, 0.88);
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 99999;
    padding: 24px;
    animation: fadeIn 0.2s ease-out;
  }

  .lightbox-container {
    display: flex;
    flex-direction: column;
    max-width: 90vw;
    max-height: 90vh;
    border-radius: var(--radius-lg);
    background: rgba(18, 22, 38, 0.7);
    border: 1px solid var(--border-glass);
    box-shadow: 0 25px 60px rgba(0, 0, 0, 0.65), 0 0 30px rgba(99, 102, 241, 0.15);
    overflow: hidden;
  }

  .lightbox-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 18px;
    background: rgba(10, 14, 28, 0.85);
    border-bottom: 1px solid var(--border-glass);
  }

  .file-title {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-muted);
    max-width: 400px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .action-buttons {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .action-btn {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid var(--border-glass);
    color: #e2e8f0;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    text-decoration: none;
    transition: 0.2s;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.2);
    color: #fff;
    transform: scale(1.06);
  }

  .close-btn:hover {
    background: rgba(239, 68, 68, 0.3);
    color: #f87171;
    border-color: rgba(239, 68, 68, 0.4);
  }

  .image-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    overflow: auto;
    max-height: calc(90vh - 60px);
  }

  .lightbox-img {
    max-width: 100%;
    max-height: 80vh;
    border-radius: var(--radius-sm);
    object-fit: contain;
    user-select: none;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: scale(0.98); }
    to { opacity: 1; transform: scale(1); }
  }
</style>
