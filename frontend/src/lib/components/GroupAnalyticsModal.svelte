<script>
  import { onMount } from 'svelte';
  import { getGroupAnalytics } from '../stores/chat';
  import {
    X,
    TrendingUp,
    MessageSquare,
    Users,
    Activity,
    Clock,
    Hash,
    Award,
    FileText,
    Image,
    Mic,
    BarChart2,
    Calendar,
    Shield
  } from 'lucide-svelte';

  export let groupID = '';
  export let groupName = 'Cộng đồng';
  export let onClose = () => {};

  let analytics = null;
  let isLoading = true;
  let errorMessage = '';

  onMount(async () => {
    try {
      analytics = await getGroupAnalytics(groupID);
    } catch (err) {
      errorMessage = err.message || 'Không thể tải dữ liệu thống kê';
    } finally {
      isLoading = false;
    }
  });

  $: maxDailyCount = Math.max(
    ...(analytics?.daily_activity || []).map((d) => d.count),
    1
  );

  function formatDateLabel(dateStr) {
    if (!dateStr) return '';
    const parts = dateStr.split('-');
    if (parts.length === 3) {
      return `${parts[2]}/${parts[1]}`;
    }
    return dateStr;
  }

  function getTypeIcon(type) {
    switch (type) {
      case 'image':
        return Image;
      case 'file':
        return FileText;
      case 'voice':
        return Mic;
      case 'poll':
        return BarChart2;
      case 'event':
        return Calendar;
      default:
        return MessageSquare;
    }
  }

  function getTypeName(type) {
    switch (type) {
      case 'image':
        return 'Hình ảnh';
      case 'file':
        return 'Tài liệu / Tệp';
      case 'voice':
        return 'Tin nhắn thoại';
      case 'poll':
        return 'Bình chọn / Poll';
      case 'event':
        return 'Sự kiện / Lịch hẹn';
      default:
        return 'Văn bản (Text)';
    }
  }

  function getTypeColor(type) {
    switch (type) {
      case 'image':
        return '#38bdf8';
      case 'file':
        return '#a78bfa';
      case 'voice':
        return '#f43f5e';
      case 'poll':
        return '#818cf8';
      case 'event':
        return '#34d399';
      default:
        return '#60a5fa';
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click={onClose}>
  <div class="modal-card glass-card" on:click={(e) => e.stopPropagation()}>
    <!-- Modal Header -->
    <div class="modal-header">
      <div class="header-left">
        <div class="header-icon-wrap">
          <TrendingUp size={20} />
        </div>
        <div>
          <h3>Thống kê & Hoạt động nhóm</h3>
          <span class="sub-text">{analytics?.group_name || groupName}</span>
        </div>
      </div>
      <button class="close-btn" on:click={onClose} title="Đóng">
        <X size={18} />
      </button>
    </div>

    <!-- Modal Content -->
    {#if isLoading}
      <div class="loading-state">
        <div class="spinner"></div>
        <span>Đang tính toán và tổng hợp dữ liệu...</span>
      </div>
    {:else if errorMessage}
      <div class="error-banner">{errorMessage}</div>
    {:else}
      <div class="analytics-scroll">
        <!-- KPI Cards Grid -->
        <div class="kpi-grid">
          <div class="kpi-card">
            <div class="kpi-icon-wrap icon-msg">
              <MessageSquare size={18} />
            </div>
            <div class="kpi-info">
              <span class="kpi-num">{analytics?.total_messages?.toLocaleString() || 0}</span>
              <span class="kpi-label">Tổng tin nhắn</span>
            </div>
          </div>

          <div class="kpi-card">
            <div class="kpi-icon-wrap icon-members">
              <Users size={18} />
            </div>
            <div class="kpi-info">
              <span class="kpi-num">{analytics?.total_members || 0}</span>
              <span class="kpi-label">Thành viên nhóm</span>
            </div>
          </div>

          <div class="kpi-card">
            <div class="kpi-icon-wrap icon-active">
              <Activity size={18} />
            </div>
            <div class="kpi-info">
              <span class="kpi-num">{analytics?.active_members_7d || 0}</span>
              <span class="kpi-label">Tích cực (7 ngày)</span>
            </div>
          </div>

          <div class="kpi-card">
            <div class="kpi-icon-wrap icon-time">
              <Clock size={18} />
            </div>
            <div class="kpi-info">
              <span class="kpi-num">{analytics?.peak_hour}:00 - {analytics?.peak_hour + 1}:00</span>
              <span class="kpi-label">Giờ sôi nổi nhất</span>
            </div>
          </div>
        </div>

        <!-- 7-Day Activity Chart -->
        <div class="chart-section glass-card">
          <div class="chart-header">
            <h4>Biểu đồ hoạt động tin nhắn 7 ngày qua</h4>
            <span class="chart-sub">Số lượng tin nhắn được gửi mỗi ngày</span>
          </div>

          <div class="bars-container">
            {#each analytics?.daily_activity || [] as day}
              {@const heightPercent = Math.max((day.count / maxDailyCount) * 100, 4)}
              <div class="bar-col">
                <div class="bar-tooltip">{day.count} tin</div>
                <div class="bar-track">
                  <div class="bar-fill" style="height: {heightPercent}%;"></div>
                </div>
                <span class="bar-date">{formatDateLabel(day.date)}</span>
              </div>
            {/each}
          </div>
        </div>

        <!-- Two Column Section -->
        <div class="two-col-grid">
          <!-- Content Type Breakdown -->
          <div class="col-card glass-card">
            <div class="card-head">
              <h4>Phân bố loại nội dung</h4>
            </div>
            <div class="distribution-list">
              {#if (analytics?.message_type_stats || []).length === 0}
                <div class="empty-text">Chưa có dữ liệu</div>
              {:else}
                {#each analytics?.message_type_stats || [] as item}
                  {@const Icon = getTypeIcon(item.type)}
                  {@const color = getTypeColor(item.type)}
                  <div class="dist-row">
                    <div class="dist-left">
                      <svelte:component this={Icon} size={14} style="color: {color};" />
                      <span class="dist-name">{getTypeName(item.type)}</span>
                    </div>
                    <div class="dist-right">
                      <span class="dist-count">{item.count}</span>
                      <span class="dist-pct">({item.percentage}%)</span>
                    </div>
                  </div>
                  <div class="dist-bar-track">
                    <div class="dist-bar-fill" style="width: {item.percentage}%; background-color: {color};"></div>
                  </div>
                {/each}
              {/if}
            </div>
          </div>

          <!-- Channel Activity Breakdown -->
          <div class="col-card glass-card">
            <div class="card-head">
              <h4>Kênh hoạt động nhiều nhất</h4>
            </div>
            <div class="distribution-list">
              {#if (analytics?.channel_stats || []).length === 0}
                <div class="empty-text">Chưa có dữ liệu kênh</div>
              {:else}
                {#each analytics?.channel_stats || [] as ch}
                  <div class="dist-row">
                    <div class="dist-left">
                      <Hash size={14} class="chan-hash-icon" />
                      <span class="dist-name">{ch.channel_name}</span>
                    </div>
                    <div class="dist-right">
                      <span class="dist-count">{ch.message_count}</span>
                      <span class="dist-pct">({ch.percentage}%)</span>
                    </div>
                  </div>
                  <div class="dist-bar-track">
                    <div class="dist-bar-fill" style="width: {ch.percentage}%; background: #10b981;"></div>
                  </div>
                {/each}
              {/if}
            </div>
          </div>
        </div>

        <!-- Top Contributors Leaderboard -->
        <div class="leaderboard-section glass-card">
          <div class="card-head">
            <div class="head-with-icon">
              <Award size={16} class="award-icon" />
              <h4>Top thành viên tích cực nhất</h4>
            </div>
            <span class="chart-sub">Dựa trên tổng số tin nhắn đóng góp</span>
          </div>

          {#if (analytics?.top_members || []).length === 0}
            <div class="empty-text">Chưa có thành viên nào gửi tin nhắn trong nhóm</div>
          {:else}
            <div class="leaderboard-list">
              {#each analytics?.top_members || [] as member, index}
                <div class="member-rank-row" class:top-1={index === 0}>
                  <div class="rank-badge-wrap">
                    {#if index === 0}
                      <span class="medal-gold">🥇</span>
                    {:else if index === 1}
                      <span class="medal-silver">🥈</span>
                    {:else if index === 2}
                      <span class="medal-bronze">🥉</span>
                    {:else}
                      <span class="rank-num">#{index + 1}</span>
                    {/if}
                  </div>

                  <img
                    src={member.avatar_url || `https://api.dicebear.com/7.x/identicon/svg?seed=${member.user_id}`}
                    alt=""
                    class="member-avatar"
                  />

                  <div class="member-details">
                    <div class="name-line">
                      <span class="member-name">{member.display_name || member.username}</span>
                      {#if member.role === 'owner'}
                        <span class="role-pill owner">👑 Chủ phòng</span>
                      {:else if member.role === 'admin'}
                        <span class="role-pill admin">🛡️ Admin</span>
                      {:else if member.role === 'moderator'}
                        <span class="role-pill mod">⚔️ Mod</span>
                      {/if}
                    </div>
                    <span class="member-uname">@{member.username}</span>
                  </div>

                  <div class="member-score">
                    <span class="score-count">{member.message_count.toLocaleString()}</span>
                    <span class="score-unit">tin nhắn</span>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 20, 0.82);
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
    max-width: 680px;
    background: rgba(16, 20, 36, 0.96);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 22px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    box-shadow: 0 25px 60px rgba(0, 0, 0, 0.7);
    max-height: 90vh;
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--border-glass);
  }
  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .header-icon-wrap {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: rgba(99, 102, 241, 0.15);
    border: 1px solid rgba(99, 102, 241, 0.35);
    color: #818cf8;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .header-left h3 {
    margin: 0;
    font-size: 16px;
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
    padding: 5px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .close-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.08); }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 0;
    gap: 12px;
    color: var(--text-dim);
    font-size: 13px;
  }
  .spinner {
    width: 28px;
    height: 28px;
    border: 3px solid rgba(255, 255, 255, 0.1);
    border-top-color: #818cf8;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  .error-banner {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171;
    font-size: 13px;
    padding: 10px 14px;
    border-radius: 8px;
  }

  .analytics-scroll {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding-right: 4px;
    max-height: calc(90vh - 80px);
  }

  /* KPI Grid */
  .kpi-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
  }
  @media (max-width: 600px) {
    .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  }
  .kpi-card {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 10px;
    padding: 12px;
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .kpi-icon-wrap {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .icon-msg { background: rgba(99, 102, 241, 0.15); color: #818cf8; }
  .icon-members { background: rgba(16, 185, 129, 0.15); color: #34d399; }
  .icon-active { background: rgba(245, 158, 11, 0.15); color: #fbbf24; }
  .icon-time { background: rgba(236, 72, 153, 0.15); color: #f472b6; }

  .kpi-info {
    display: flex;
    flex-direction: column;
  }
  .kpi-num {
    font-size: 15px;
    font-weight: 700;
    color: #fff;
    line-height: 1.2;
  }
  .kpi-label {
    font-size: 11px;
    color: var(--text-dim);
    margin-top: 2px;
  }

  /* 7-Day Chart */
  .chart-section {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .chart-header h4, .card-head h4 {
    margin: 0;
    font-size: 13.5px;
    font-weight: 600;
    color: #fff;
  }
  .chart-sub {
    font-size: 11px;
    color: var(--text-dim);
  }

  .bars-container {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 8px;
    height: 120px;
    align-items: flex-end;
    padding-top: 18px;
  }
  .bar-col {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100%;
    justify-content: flex-end;
    gap: 6px;
  }
  .bar-track {
    width: 100%;
    max-width: 24px;
    height: 85px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 6px;
    display: flex;
    align-items: flex-end;
    overflow: hidden;
  }
  .bar-fill {
    width: 100%;
    background: linear-gradient(180deg, #818cf8 0%, #4f46e5 100%);
    border-radius: 6px 6px 0 0;
    transition: height 0.4s ease;
  }
  .bar-col:hover .bar-fill {
    background: linear-gradient(180deg, #a5b4fc 0%, #6366f1 100%);
    box-shadow: 0 0 10px rgba(99, 102, 241, 0.5);
  }
  .bar-tooltip {
    position: absolute;
    top: -4px;
    background: #1e1b2e;
    border: 1px solid var(--border-glass);
    color: #fff;
    font-size: 10px;
    padding: 2px 5px;
    border-radius: 4px;
    opacity: 0;
    transition: 0.15s;
    pointer-events: none;
    white-space: nowrap;
  }
  .bar-col:hover .bar-tooltip {
    opacity: 1;
    top: -12px;
  }
  .bar-date {
    font-size: 10.5px;
    color: #94a3b8;
  }

  /* Two Column Grid */
  .two-col-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  @media (max-width: 600px) {
    .two-col-grid { grid-template-columns: 1fr; }
  }
  .col-card {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 14px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .distribution-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .dist-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 12px;
  }
  .dist-left {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #cbd5e1;
  }
  .chan-hash-icon { color: #10b981; }
  .dist-name { font-weight: 500; }
  .dist-right {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .dist-count { font-weight: 600; color: #fff; }
  .dist-pct { color: var(--text-dim); font-size: 11px; }
  .dist-bar-track {
    width: 100%;
    height: 4px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 4px;
    overflow: hidden;
  }
  .dist-bar-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.3s ease;
  }

  /* Leaderboard */
  .leaderboard-section {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .head-with-icon {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .award-icon { color: #fbbf24; }
  .leaderboard-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .member-rank-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.04);
  }
  .member-rank-row.top-1 {
    background: rgba(251, 191, 36, 0.05);
    border-color: rgba(251, 191, 36, 0.2);
  }
  .rank-badge-wrap {
    width: 24px;
    text-align: center;
    font-size: 14px;
  }
  .rank-num { font-size: 12px; color: var(--text-dim); font-weight: 600; }
  .member-avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
  }
  .member-details {
    flex: 1;
    display: flex;
    flex-direction: column;
  }
  .name-line {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .member-name {
    font-size: 13px;
    font-weight: 500;
    color: #fff;
  }
  .member-uname {
    font-size: 11px;
    color: var(--text-dim);
  }
  .role-pill {
    font-size: 9.5px;
    padding: 1px 5px;
    border-radius: 6px;
    font-weight: 600;
  }
  .role-pill.owner { background: rgba(251, 191, 36, 0.15); color: #fbbf24; }
  .role-pill.admin { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
  .role-pill.mod { background: rgba(6, 182, 212, 0.15); color: #06b6d4; }

  .member-score {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }
  .score-count {
    font-size: 13.5px;
    font-weight: 700;
    color: #a78bfa;
  }
  .score-unit {
    font-size: 10px;
    color: var(--text-dim);
  }

  .empty-text {
    font-size: 12px;
    color: var(--text-dim);
    text-align: center;
    padding: 12px 0;
  }

  @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
