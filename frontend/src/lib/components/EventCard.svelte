<script>
  import { currentUser } from '../stores/auth';
  import { rsvpGroupEvent } from '../stores/chat';
  import { Calendar, Clock, MapPin, Users, Check, HelpCircle, XCircle } from 'lucide-svelte';

  export let message = {};

  let event = message.event || null;
  let isUpdating = false;

  $: if (message.event) {
    event = message.event;
  }

  $: myId = $currentUser?.id;
  $: myStatus = (event?.attendees || []).find((a) => a.user_id === myId)?.status;

  $: goingCount = (event?.attendees || []).filter((a) => a.status === 'going').length;
  $: maybeCount = (event?.attendees || []).filter((a) => a.status === 'maybe').length;
  $: declinedCount = (event?.attendees || []).filter((a) => a.status === 'declined').length;

  $: startTime = event?.start_time ? new Date(event.start_time) : null;
  $: isPast = startTime ? startTime < new Date() : false;

  async function handleRSVP(status) {
    if (!event?.id || !message.group_id || isUpdating) return;
    isUpdating = true;
    try {
      const updated = await rsvpGroupEvent(message.group_id, event.id, status);
      if (updated) {
        event = updated;
      }
    } catch (err) {
      alert('Lỗi cập nhật phản hồi: ' + err.message);
    } finally {
      isUpdating = false;
    }
  }

  function formatEventTime(date) {
    if (!date) return '';
    return date.toLocaleDateString([], {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
</script>

<div class="event-card glass-card">
  <!-- Top Banner / Date Box -->
  <div class="event-top">
    <div class="date-badge">
      <span class="month">{startTime ? startTime.toLocaleDateString([], { month: 'short' }).toUpperCase() : 'SỰ KIỆN'}</span>
      <span class="day">{startTime ? startTime.getDate() : '📅'}</span>
    </div>

    <div class="event-main-info">
      <div class="event-header-line">
        <span class="event-tag">📅 Lịch hẹn nhóm</span>
        {#if isPast}
          <span class="past-tag">Đã qua</span>
        {/if}
      </div>
      <h4 class="event-title">{event?.title || message.content}</h4>
      <div class="meta-row">
        <Clock size={13} class="meta-icon" />
        <span>{formatEventTime(startTime)}</span>
      </div>
      {#if event?.location}
        <div class="meta-row">
          <MapPin size={13} class="meta-icon" />
          <span>{event.location}</span>
        </div>
      {/if}
    </div>
  </div>

  {#if event?.description}
    <p class="event-desc">{event.description}</p>
  {/if}

  <!-- Attendees count -->
  <div class="attendees-bar">
    <div class="attendee-avatars">
      <Users size={14} class="users-icon" />
      <span><b>{goingCount}</b> người tham gia</span>
    </div>

    <!-- RSVP Buttons -->
    <div class="rsvp-actions">
      <button
        class="rsvp-btn going-btn"
        class:active={myStatus === 'going'}
        on:click={() => handleRSVP('going')}
        disabled={isUpdating}
        title="Tôi sẽ tham gia"
      >
        <Check size={13} />
        <span>Tham gia ({goingCount})</span>
      </button>

      <button
        class="rsvp-btn maybe-btn"
        class:active={myStatus === 'maybe'}
        on:click={() => handleRSVP('maybe')}
        disabled={isUpdating}
        title="Có thể tham gia"
      >
        <HelpCircle size={13} />
        <span>Có thể ({maybeCount})</span>
      </button>

      <button
        class="rsvp-btn decline-btn"
        class:active={myStatus === 'declined'}
        on:click={() => handleRSVP('declined')}
        disabled={isUpdating}
        title="Không thể tham gia"
      >
        <XCircle size={13} />
        <span>Từ chối</span>
      </button>
    </div>
  </div>
</div>

<style>
  .event-card {
    width: 100%;
    max-width: 480px;
    background: rgba(22, 27, 46, 0.85);
    border: 1px solid rgba(16, 185, 129, 0.3);
    border-radius: 12px;
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin: 4px 0;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  }

  .event-top {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }

  .date-badge {
    width: 52px;
    height: 54px;
    background: rgba(16, 185, 129, 0.15);
    border: 1px solid rgba(16, 185, 129, 0.4);
    border-radius: 10px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .date-badge .month {
    font-size: 10px;
    font-weight: 700;
    color: #34d399;
  }
  .date-badge .day {
    font-size: 18px;
    font-weight: 800;
    color: #fff;
    line-height: 1;
  }

  .event-main-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .event-header-line {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .event-tag {
    font-size: 10.5px;
    font-weight: 600;
    color: #34d399;
  }
  .past-tag {
    font-size: 10px;
    padding: 1px 5px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.1);
    color: #94a3b8;
  }

  .event-title {
    margin: 0;
    font-size: 14.5px;
    font-weight: 600;
    color: #fff;
    line-height: 1.3;
  }
  .meta-row {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #cbd5e1;
  }
  .meta-icon {
    color: #38bdf8;
    flex-shrink: 0;
  }

  .event-desc {
    font-size: 12.5px;
    color: #94a3b8;
    margin: 0;
    line-height: 1.4;
    background: rgba(255, 255, 255, 0.03);
    padding: 6px 10px;
    border-radius: 6px;
  }

  /* Attendees Bar & RSVP */
  .attendees-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 8px;
    padding-top: 8px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
  }
  .attendee-avatars {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11.5px;
    color: var(--text-dim);
  }
  .users-icon {
    color: #818cf8;
  }

  .rsvp-actions {
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .rsvp-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 500;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #cbd5e1;
    cursor: pointer;
    transition: 0.15s;
  }
  .rsvp-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
  }
  .going-btn.active {
    background: rgba(16, 185, 129, 0.2);
    border-color: rgba(16, 185, 129, 0.5);
    color: #34d399;
    font-weight: 600;
  }
  .maybe-btn.active {
    background: rgba(245, 158, 11, 0.2);
    border-color: rgba(245, 158, 11, 0.5);
    color: #fbbf24;
    font-weight: 600;
  }
  .decline-btn.active {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.5);
    color: #f87171;
    font-weight: 600;
  }
</style>
