<script>
  import { X, Clock, Calendar, Check } from 'lucide-svelte';
  import { apiRequest } from '../services/api';
  import { token } from '../stores/auth';

  export let receiverID = '';
  export let groupID = '';
  export let channelID = '';
  export let content = '';
  export let isSilent = false;
  export let onScheduled = () => {};
  export let onClose = () => {};

  let scheduledDate = '';
  let isLoading = false;
  let errorMessage = '';

  // Khởi tạo ngày giờ mặc định là +30 phút
  const defaultDate = new Date(Date.now() + 30 * 60 * 1000);
  scheduledDate = formatDateTimeLocal(defaultDate);

  function formatDateTimeLocal(d) {
    const pad = (n) => (n < 10 ? '0' + n : n);
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }

  function setPreset(minutes) {
    const d = new Date(Date.now() + minutes * 60 * 1000);
    scheduledDate = formatDateTimeLocal(d);
  }

  function setTomorrowMorning() {
    const d = new Date();
    d.setDate(d.getDate() + 1);
    d.setHours(9, 0, 0, 0);
    scheduledDate = formatDateTimeLocal(d);
  }

  async function handleSchedule() {
    if (!content.trim()) {
      errorMessage = 'Nội dung tin nhắn không được để trống';
      return;
    }

    const targetTime = new Date(scheduledDate);
    if (isNaN(targetTime.getTime())) {
      errorMessage = 'Thời gian hẹn không hợp lệ';
      return;
    }

    if (targetTime.getTime() <= Date.now()) {
      errorMessage = 'Thời gian hẹn phải ở tương lai';
      return;
    }

    isLoading = true;
    errorMessage = '';

    try {
      const payload = {
        content: content,
        type: 'text',
        is_silent: isSilent,
        scheduled_at: targetTime.toISOString()
      };

      if (groupID) {
        payload.group_id = groupID;
        if (channelID) {
          payload.channel_id = channelID;
        }
      } else {
        payload.receiver_id = receiverID;
      }

      await apiRequest('/chat/scheduled', 'POST', payload, $token);
      onScheduled();
      onClose();
    } catch (e) {
      errorMessage = e.message || 'Không thể lên lịch gửi tin';
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="modal-backdrop" on:click|self={onClose}>
  <div class="modal-box glass-card">
    <div class="modal-header">
      <div class="title-wrap">
        <Clock size={18} class="clock-icon" />
        <h3>Lên lịch gửi tin nhắn</h3>
      </div>
      <button class="close-btn" on:click={onClose}>
        <X size={18} />
      </button>
    </div>

    <div class="modal-body">
      <div class="snippet-preview">
        <span class="preview-label">Tin nhắn:</span>
        <p class="preview-text">{content}</p>
      </div>

      <div class="presets-row">
        <button class="preset-btn" on:click={() => setPreset(5)}>+5 phút</button>
        <button class="preset-btn" on:click={() => setPreset(30)}>+30 phút</button>
        <button class="preset-btn" on:click={() => setPreset(60)}>+1 giờ</button>
        <button class="preset-btn" on:click={setTomorrowMorning}>Sáng mai (09:00)</button>
      </div>

      <div class="picker-group">
        <label for="schedule-time">Hoặc chọn thời gian tùy ý:</label>
        <input
          id="schedule-time"
          type="datetime-local"
          bind:value={scheduledDate}
          class="datetime-input"
        />
      </div>

      {#if errorMessage}
        <div class="error-msg">{errorMessage}</div>
      {/if}
    </div>

    <div class="modal-footer">
      <button class="btn-cancel" on:click={onClose}>Hủy</button>
      <button class="btn-confirm" disabled={isLoading} on:click={handleSchedule}>
        {#if isLoading}
          Đang lưu...
        {:else}
          <Check size={16} /> Lên lịch gửi
        {/if}
      </button>
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
    width: 400px;
    max-width: 90vw;
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
    padding: 14px 18px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .title-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .clock-icon {
    color: #f59e0b;
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

  .modal-body {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .snippet-preview {
    padding: 8px 10px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 6px;
    border-left: 3px solid #f59e0b;
  }

  .preview-label {
    font-size: 11px;
    color: #94a3b8;
  }

  .preview-text {
    margin: 2px 0 0;
    font-size: 13px;
    color: #e2e8f0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .presets-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 6px;
  }

  .preset-btn {
    padding: 6px 10px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #cbd5e1;
    border-radius: 6px;
    font-size: 12px;
    cursor: pointer;
    transition: 0.15s;
  }

  .preset-btn:hover {
    background: rgba(245, 158, 11, 0.2);
    color: #fbbf24;
    border-color: rgba(245, 158, 11, 0.4);
  }

  .picker-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .picker-group label {
    font-size: 12px;
    color: #94a3b8;
  }

  .datetime-input {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 6px;
    color: #fff;
    padding: 8px;
    font-size: 13px;
    outline: none;
    color-scheme: dark;
  }

  .error-msg {
    color: #ef4444;
    font-size: 12px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
  }

  .btn-cancel {
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #94a3b8;
    padding: 6px 14px;
    border-radius: 6px;
    font-size: 13px;
    cursor: pointer;
  }

  .btn-cancel:hover {
    color: #fff;
  }

  .btn-confirm {
    background: #f59e0b;
    border: none;
    color: #000;
    font-weight: 600;
    padding: 6px 14px;
    border-radius: 6px;
    font-size: 13px;
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    transition: 0.15s;
  }

  .btn-confirm:hover {
    background: #d97706;
  }

  .btn-confirm:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
