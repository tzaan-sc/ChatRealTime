<script>
  import { createPoll } from '../stores/chat';
  import { X, Plus, Trash2, BarChart2, CheckCircle2 } from 'lucide-svelte';

  export let groupID = '';
  export let channelID = '';
  export let onClose = () => {};

  let question = '';
  let options = ['', ''];
  let multipleChoice = false;
  let isAnonymous = false;
  let isSubmitting = false;
  let errorMessage = '';

  function addOption() {
    if (options.length < 8) {
      options = [...options, ''];
    }
  }

  function removeOption(index) {
    if (options.length > 2) {
      options = options.filter((_, i) => i !== index);
    }
  }

  async function handleSubmit() {
    if (!question.trim()) {
      errorMessage = 'Vui lòng nhập câu hỏi thăm dò ý kiến';
      return;
    }

    const filteredOptions = options.map((o) => o.trim()).filter(Boolean);
    if (filteredOptions.length < 2) {
      errorMessage = 'Cần ít nhất 2 phương án lựa chọn';
      return;
    }

    isSubmitting = true;
    errorMessage = '';

    try {
      await createPoll(groupID, {
        question: question.trim(),
        options: filteredOptions,
        multiple_choice: multipleChoice,
        is_anonymous: isAnonymous,
        channel_id: channelID || ''
      });
      onClose();
    } catch (err) {
      errorMessage = err.message || 'Không thể tạo cuộc bình chọn';
    } finally {
      isSubmitting = false;
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click={onClose}>
  <div class="modal-card glass-card" on:click={(e) => e.stopPropagation()}>
    <!-- Header -->
    <div class="modal-header">
      <div class="header-title">
        <div class="header-icon-wrap">
          <BarChart2 size={18} />
        </div>
        <div>
          <h3>Tạo cuộc bình chọn</h3>
          <span class="sub-text">Khảo sát ý kiến thành viên trong kênh</span>
        </div>
      </div>
      <button class="close-btn" on:click={onClose} title="Đóng">
        <X size={18} />
      </button>
    </div>

    <!-- Error Banner -->
    {#if errorMessage}
      <div class="error-banner">{errorMessage}</div>
    {/if}

    <!-- Form -->
    <div class="modal-body">
      <!-- Question -->
      <div class="form-group">
        <label for="poll-question">Câu hỏi / Chủ đề bình chọn:</label>
        <textarea
          id="poll-question"
          rows="2"
          placeholder="Ví dụ: Trưa nay ăn gì? / Thống nhất lịch họp sprint..."
          bind:value={question}
        ></textarea>
      </div>

      <!-- Options -->
      <div class="form-group">
        <label>Các phương án lựa chọn:</label>
        <div class="options-list">
          {#each options as opt, index}
            <div class="option-row">
              <span class="opt-num">{index + 1}</span>
              <input
                type="text"
                placeholder={`Phương án ${index + 1}...`}
                bind:value={options[index]}
              />
              {#if options.length > 2}
                <button
                  type="button"
                  class="remove-opt-btn"
                  on:click={() => removeOption(index)}
                  title="Xóa phương án này"
                >
                  <Trash2 size={14} />
                </button>
              {/if}
            </div>
          {/each}
        </div>

        {#if options.length < 8}
          <button type="button" class="add-opt-btn" on:click={addOption}>
            <Plus size={14} />
            <span>Thêm phương án</span>
          </button>
        {/if}
      </div>

      <!-- Settings / Toggles -->
      <div class="settings-box">
        <label class="toggle-row">
          <div class="toggle-info">
            <span class="toggle-label">Cho phép chọn nhiều phương án</span>
            <span class="toggle-desc">Thành viên có thể tick chọn nhiều hơn 1 câu trả lời</span>
          </div>
          <input type="checkbox" bind:checked={multipleChoice} class="checkbox-custom" />
        </label>

        <label class="toggle-row">
          <div class="toggle-info">
            <span class="toggle-label">Bình chọn ẩn danh</span>
            <span class="toggle-desc">Không hiển thị tên và avatar của người tham gia bầu chọn</span>
          </div>
          <input type="checkbox" bind:checked={isAnonymous} class="checkbox-custom" />
        </label>
      </div>
    </div>

    <!-- Footer -->
    <div class="modal-footer">
      <button class="cancel-btn" on:click={onClose} disabled={isSubmitting}>Hủy</button>
      <button class="submit-btn" on:click={handleSubmit} disabled={isSubmitting}>
        {#if isSubmitting}
          <span>Đang tạo...</span>
        {:else}
          <CheckCircle2 size={16} />
          <span>Tạo bình chọn</span>
        {/if}
      </button>
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 20, 0.78);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    padding: 20px;
    animation: fadeIn 0.2s ease;
  }

  .modal-card {
    width: 100%;
    max-width: 460px;
    background: rgba(18, 22, 38, 0.95);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--border-glass);
  }
  .header-title {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .header-icon-wrap {
    width: 36px;
    height: 36px;
    border-radius: 10px;
    background: rgba(99, 102, 241, 0.15);
    border: 1px solid rgba(99, 102, 241, 0.3);
    color: #818cf8;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .header-title h3 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: #fff;
  }
  .sub-text {
    font-size: 11.5px;
    color: var(--text-dim);
  }
  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    border-radius: 50%;
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .close-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.08); }

  .error-banner {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
    font-size: 12.5px;
    padding: 8px 12px;
    border-radius: 6px;
  }

  .modal-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .form-group label {
    font-size: 12px;
    color: #cbd5e1;
    font-weight: 500;
  }
  .form-group textarea {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #fff;
    padding: 8px 10px;
    border-radius: 8px;
    font-size: 13px;
    outline: none;
    resize: none;
  }
  .form-group textarea:focus {
    border-color: #818cf8;
  }

  .options-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .option-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .opt-num {
    font-size: 11px;
    color: var(--text-dim);
    width: 16px;
    text-align: center;
  }
  .option-row input {
    flex: 1;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #fff;
    padding: 7px 10px;
    border-radius: 6px;
    font-size: 12.5px;
    outline: none;
  }
  .option-row input:focus {
    border-color: #818cf8;
  }
  .remove-opt-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .remove-opt-btn:hover { color: #f87171; background: rgba(239, 68, 68, 0.1); }

  .add-opt-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px dashed rgba(255, 255, 255, 0.15);
    color: #a78bfa;
    padding: 7px;
    border-radius: 6px;
    font-size: 12px;
    cursor: pointer;
    transition: 0.15s;
    margin-top: 2px;
  }
  .add-opt-btn:hover {
    background: rgba(167, 139, 250, 0.1);
    border-color: #a78bfa;
  }

  /* Settings Box */
  .settings-box {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 8px;
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .toggle-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    cursor: pointer;
  }
  .toggle-info {
    display: flex;
    flex-direction: column;
  }
  .toggle-label {
    font-size: 12px;
    font-weight: 500;
    color: #e2e8f0;
  }
  .toggle-desc {
    font-size: 10.5px;
    color: var(--text-dim);
  }
  .checkbox-custom {
    width: 16px;
    height: 16px;
    accent-color: #6366f1;
    cursor: pointer;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--border-glass);
  }
  .cancel-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 12.5px;
    padding: 7px 14px;
    border-radius: 6px;
    cursor: pointer;
  }
  .cancel-btn:hover { color: #fff; }
  .submit-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    background: #6366f1;
    border: none;
    color: #fff;
    font-size: 12.5px;
    font-weight: 500;
    padding: 7px 16px;
    border-radius: 6px;
    cursor: pointer;
    transition: 0.15s;
  }
  .submit-btn:hover { background: #4f46e5; }
  .submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
