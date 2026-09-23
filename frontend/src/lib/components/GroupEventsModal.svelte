<script>
  import { onMount } from 'svelte';
  import { currentUser } from '../stores/auth';
  import {
    groupEvents,
    getGroupEvents,
    createGroupEvent,
    rsvpGroupEvent,
    deleteGroupEvent
  } from '../stores/chat';
  import {
    X,
    Calendar,
    Plus,
    Clock,
    MapPin,
    Users,
    Trash2,
    Check,
    HelpCircle,
    XCircle,
    ChevronRight
  } from 'lucide-svelte';

  export let groupID = '';
  export let channels = [];
  export let onClose = () => {};

  let activeTab = 'upcoming'; // 'upcoming' | 'create'
  let isLoading = true;
  let isCreating = false;
  let errorMessage = '';

  // Form fields
  let title = '';
  let description = '';
  let location = '';
  let startTime = '';
  let endTime = '';
  let selectedChannelID = channels?.[0]?.id || '';

  $: myId = $currentUser?.id;

  onMount(async () => {
    // Đặt mặc định thời gian bắt đầu là 1 tiếng nữa
    const now = new Date();
    now.setHours(now.getHours() + 1);
    now.setMinutes(0);
    startTime = formatDateTimeInput(now);

    const end = new Date(now);
    end.setHours(end.getHours() + 1);
    endTime = formatDateTimeInput(end);

    await loadEvents();
  });

  function formatDateTimeInput(d) {
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }

  async function loadEvents() {
    isLoading = true;
    try {
      await getGroupEvents(groupID);
    } catch (err) {
      errorMessage = err.message || 'Không thể tải sự kiện nhóm';
    } finally {
      isLoading = false;
    }
  }

  async function handleCreateEvent() {
    if (!title.trim()) {
      alert('Vui lòng nhập tiêu đề sự kiện');
      return;
    }
    if (!startTime) {
      alert('Vui lòng chọn thời gian bắt đầu');
      return;
    }

    isCreating = true;
    try {
      await createGroupEvent(groupID, {
        title: title.trim(),
        description: description.trim(),
        location: location.trim(),
        start_time: new Date(startTime).toISOString(),
        end_time: endTime ? new Date(endTime).toISOString() : null,
        channel_id: selectedChannelID
      });
      // Reset form & về danh sách
      title = '';
      description = '';
      location = '';
      activeTab = 'upcoming';
      await loadEvents();
    } catch (err) {
      alert('Lỗi tạo sự kiện: ' + err.message);
    } finally {
      isCreating = false;
    }
  }

  async function handleRSVP(event, status) {
    try {
      await rsvpGroupEvent(groupID, event.id, status);
    } catch (err) {
      alert('Lỗi phản hồi sự kiện: ' + err.message);
    }
  }

  async function handleDeleteEvent(eventId) {
    if (!confirm('Bạn có chắc muốn xóa sự kiện này?')) return;
    try {
      await deleteGroupEvent(groupID, eventId);
    } catch (err) {
      alert('Lỗi xóa sự kiện: ' + err.message);
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click={onClose}>
  <div class="modal-card glass-card" on:click={(e) => e.stopPropagation()}>
    <!-- Modal Header -->
    <div class="modal-header">
      <div class="header-title">
        <div class="icon-wrap">
          <Calendar size={18} />
        </div>
        <div>
          <h3>Sự kiện & Lịch hẹn chung</h3>
          <span class="sub-text">Quản lý các hoạt động và buổi họp của nhóm</span>
        </div>
      </div>
      <button class="close-btn" on:click={onClose} title="Đóng">
        <X size={18} />
      </button>
    </div>

    <!-- Tab Bar -->
    <div class="tab-bar">
      <button
        class="tab-btn"
        class:active={activeTab === 'upcoming'}
        on:click={() => (activeTab = 'upcoming')}
      >
        <span>Sự kiện sắp tới</span>
        <span class="tab-pill">{$groupEvents.length}</span>
      </button>

      <button
        class="tab-btn"
        class:active={activeTab === 'create'}
        on:click={() => (activeTab = 'create')}
      >
        <Plus size={14} />
        <span>Tạo sự kiện mới</span>
      </button>
    </div>

    <!-- Tab Content -->
    {#if activeTab === 'upcoming'}
      <div class="events-scroll-container">
        {#if isLoading}
          <div class="loading-state">Đang tải danh sách sự kiện...</div>
        {:else if $groupEvents.length === 0}
          <div class="empty-state">
            <div class="empty-icon-wrap">
              <Calendar size={28} />
            </div>
            <p>Chưa có sự kiện nào sắp tới trong nhóm.</p>
            <button class="create-prompt-btn" on:click={() => (activeTab = 'create')}>
              + Tạo sự kiện đầu tiên
            </button>
          </div>
        {:else}
          <div class="events-list">
            {#each $groupEvents as event}
              {@const myStatus = (event.attendees || []).find((a) => a.user_id === myId)?.status}
              {@const goingCount = (event.attendees || []).filter((a) => a.status === 'going').length}
              {@const maybeCount = (event.attendees || []).filter((a) => a.status === 'maybe').length}
              {@const startDate = event.start_time ? new Date(event.start_time) : null}
              {@const isCreator = event.created_by === myId}

              <div class="event-item glass-card">
                <div class="event-item-top">
                  <div class="date-badge">
                    <span class="month">{startDate ? startDate.toLocaleDateString([], { month: 'short' }).toUpperCase() : ''}</span>
                    <span class="day">{startDate ? startDate.getDate() : ''}</span>
                  </div>

                  <div class="item-details">
                    <div class="item-title-line">
                      <h4>{event.title}</h4>
                      {#if isCreator}
                        <button
                          class="del-event-btn"
                          on:click={() => handleDeleteEvent(event.id)}
                          title="Xóa sự kiện"
                        >
                          <Trash2 size={13} />
                        </button>
                      {/if}
                    </div>

                    <div class="item-meta">
                      <span class="meta-entry">
                        <Clock size={12} />
                        {startDate ? startDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
                      </span>
                      {#if event.location}
                        <span class="meta-entry">
                          <MapPin size={12} />
                          {event.location}
                        </span>
                      {/if}
                    </div>

                    {#if event.description}
                      <p class="item-desc">{event.description}</p>
                    {/if}
                  </div>
                </div>

                <!-- Attendees & RSVP -->
                <div class="item-bottom">
                  <div class="attendee-summary">
                    <Users size={13} />
                    <span><b>{goingCount}</b> người tham gia</span>
                  </div>

                  <div class="rsvp-btn-row">
                    <button
                      class="rsvp-mini-btn"
                      class:active={myStatus === 'going'}
                      on:click={() => handleRSVP(event, 'going')}
                    >
                      <Check size={12} />
                      <span>Tham gia</span>
                    </button>
                    <button
                      class="rsvp-mini-btn"
                      class:active={myStatus === 'maybe'}
                      on:click={() => handleRSVP(event, 'maybe')}
                    >
                      <HelpCircle size={12} />
                      <span>Có thể</span>
                    </button>
                    <button
                      class="rsvp-mini-btn"
                      class:active={myStatus === 'declined'}
                      on:click={() => handleRSVP(event, 'declined')}
                    >
                      <XCircle size={12} />
                      <span>Từ chối</span>
                    </button>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {:else}
      <!-- CREATE EVENT FORM -->
      <div class="create-form-container">
        <div class="form-group">
          <label for="event-title">Tiêu đề sự kiện / Lịch hẹn:</label>
          <input
            id="event-title"
            type="text"
            placeholder="Ví dụ: Họp Sprint Review, Voice hangout tối thứ 7..."
            bind:value={title}
          />
        </div>

        <div class="form-row-2">
          <div class="form-group">
            <label for="event-start">Thời gian bắt đầu:</label>
            <input id="event-start" type="datetime-local" bind:value={startTime} />
          </div>

          <div class="form-group">
            <label for="event-end">Thời gian kết thúc (Tùy chọn):</label>
            <input id="event-end" type="datetime-local" bind:value={endTime} />
          </div>
        </div>

        <div class="form-row-2">
          <div class="form-group">
            <label for="event-loc">Địa điểm / Phòng họp:</label>
            <input
              id="event-loc"
              type="text"
              placeholder="Kênh thoại #chung, Google Meet, Phòng A..."
              bind:value={location}
            />
          </div>

          <div class="form-group">
            <label for="event-chan">Đăng thông báo lên kênh:</label>
            <select id="event-chan" bind:value={selectedChannelID}>
              {#each channels as ch}
                <option value={ch.id}>#{ch.name}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="form-group">
          <label for="event-desc">Mô tả chi tiết nội dung sự kiện:</label>
          <textarea
            id="event-desc"
            rows="3"
            placeholder="Nội dung thảo luận, tài liệu cần chuẩn bị..."
            bind:value={description}
          ></textarea>
        </div>

        <div class="form-actions">
          <button class="back-text-btn" on:click={() => (activeTab = 'upcoming')}>Hủy</button>
          <button class="create-submit-btn" on:click={handleCreateEvent} disabled={isCreating}>
            <Calendar size={15} />
            <span>{isCreating ? 'Đang tạo...' : 'Tạo sự kiện ngay'}</span>
          </button>
        </div>
      </div>
    {/if}
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
    max-width: 520px;
    background: rgba(18, 22, 38, 0.95);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
    max-height: 88vh;
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--border-glass);
  }
  .header-title {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .icon-wrap {
    width: 36px;
    height: 36px;
    border-radius: 10px;
    background: rgba(16, 185, 129, 0.15);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: #34d399;
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
    font-size: 11px;
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

  /* Tab Bar */
  .tab-bar {
    display: flex;
    align-items: center;
    gap: 6px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    padding-bottom: 4px;
  }
  .tab-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 12.5px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.15s;
  }
  .tab-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.04); }
  .tab-btn.active {
    color: #34d399;
    background: rgba(16, 185, 129, 0.12);
    font-weight: 600;
  }
  .tab-pill {
    font-size: 11px;
    padding: 1px 6px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.08);
  }

  .events-scroll-container {
    max-height: 420px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-right: 2px;
  }

  .loading-state {
    text-align: center;
    color: var(--text-dim);
    font-size: 13px;
    padding: 30px 0;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 36px 20px;
    text-align: center;
    gap: 8px;
  }
  .empty-icon-wrap {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.04);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #64748b;
  }
  .empty-state p {
    font-size: 13px;
    color: var(--text-dim);
    margin: 0;
  }
  .create-prompt-btn {
    margin-top: 6px;
    background: rgba(16, 185, 129, 0.15);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: #34d399;
    padding: 6px 14px;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
  }
  .create-prompt-btn:hover { background: rgba(16, 185, 129, 0.25); color: #fff; }

  /* Events List */
  .events-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .event-item {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 10px;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .event-item-top {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }
  .date-badge {
    width: 44px;
    height: 46px;
    background: rgba(16, 185, 129, 0.15);
    border: 1px solid rgba(16, 185, 129, 0.35);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .date-badge .month {
    font-size: 9px;
    font-weight: 700;
    color: #34d399;
  }
  .date-badge .day {
    font-size: 16px;
    font-weight: 800;
    color: #fff;
    line-height: 1;
  }
  .item-details {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .item-title-line {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .item-title-line h4 {
    margin: 0;
    font-size: 13.5px;
    font-weight: 600;
    color: #fff;
  }
  .del-event-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    padding: 3px;
    border-radius: 4px;
  }
  .del-event-btn:hover { color: #f87171; background: rgba(239, 68, 68, 0.15); }

  .item-meta {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 11.5px;
    color: #94a3b8;
  }
  .meta-entry {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .item-desc {
    font-size: 11.5px;
    color: var(--text-dim);
    margin: 4px 0 0 0;
    line-height: 1.35;
  }

  .item-bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 6px;
    padding-top: 6px;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }
  .attendee-summary {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 11px;
    color: #94a3b8;
  }
  .rsvp-btn-row {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .rsvp-mini-btn {
    display: flex;
    align-items: center;
    gap: 3px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #cbd5e1;
    font-size: 10.5px;
    padding: 3px 7px;
    border-radius: 5px;
    cursor: pointer;
    transition: 0.15s;
  }
  .rsvp-mini-btn:hover { background: rgba(255, 255, 255, 0.1); color: #fff; }
  .rsvp-mini-btn.active {
    background: rgba(16, 185, 129, 0.2);
    border-color: rgba(16, 185, 129, 0.4);
    color: #34d399;
    font-weight: 600;
  }

  /* Create Form */
  .create-form-container {
    max-height: 420px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-right: 2px;
  }
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .form-group label {
    font-size: 11.5px;
    color: #cbd5e1;
    font-weight: 500;
  }
  .form-group input,
  .form-group select,
  .form-group textarea {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #fff;
    padding: 7px 10px;
    border-radius: 6px;
    font-size: 12px;
    outline: none;
  }
  .form-group select option {
    background: #1e1b2e;
    color: #fff;
  }
  .form-group textarea {
    resize: none;
  }
  .form-row-2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .form-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 4px;
    padding-top: 8px;
    border-top: 1px solid var(--border-glass);
  }
  .back-text-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 12px;
    cursor: pointer;
  }
  .back-text-btn:hover { color: #fff; }
  .create-submit-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    background: #10b981;
    border: none;
    color: #fff;
    font-size: 12px;
    font-weight: 500;
    padding: 7px 14px;
    border-radius: 6px;
    cursor: pointer;
    transition: 0.15s;
  }
  .create-submit-btn:hover { background: #059669; }
  .create-submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
