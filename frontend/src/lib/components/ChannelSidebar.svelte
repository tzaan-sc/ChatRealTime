<script>
  import {
    activeGroup,
    activeChannel,
    selectChannel,
    deleteChannel,
    deleteCategory,
    channelUnreadCounts,
    groupEvents,
    getGroupEvents
  } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import {
    Hash,
    Megaphone,
    ChevronDown,
    ChevronRight,
    Plus,
    Trash2,
    Settings,
    Users,
    Calendar
  } from 'lucide-svelte';
  import CreateChannelModal from './CreateChannelModal.svelte';
  import GroupEventsModal from './GroupEventsModal.svelte';

  export let onOpenMembersModal = () => {};

  let showCreateModal = false;
  let showEventsModal = false;
  let targetCategoryID = '';
  let collapsedCategories = {}; // map catID -> boolean

  $: if ($activeGroup?.id) {
    getGroupEvents($activeGroup.id);
  }

  $: amIAdmin =
    $activeGroup?.creator_id === $currentUser?.id ||
    $activeGroup?.members?.some((m) => m.id === $currentUser?.id && m.is_admin);

  $: categories = $activeGroup?.categories || [];
  $: channels = $activeGroup?.channels || [];

  // Phân nhóm channels theo category_id
  $: categorizedChannels = categories.map((cat) => ({
    category: cat,
    channels: channels.filter((c) => c.category_id === cat.id)
  }));

  $: uncategorizedChannels = channels.filter(
    (c) => !c.category_id || !categories.some((cat) => cat.id === c.category_id)
  );

  function toggleCollapse(catId) {
    collapsedCategories[catId] = !collapsedCategories[catId];
  }

  function openCreateModal(catId = '') {
    targetCategoryID = catId;
    showCreateModal = true;
  }

  async function handleDeleteChannel(channel, e) {
    e.stopPropagation();
    if (!confirm(`Bạn có chắc chắn muốn xóa kênh #${channel.name}? Toàn bộ tin nhắn trong kênh này sẽ bị xóa.`)) {
      return;
    }
    try {
      await deleteChannel($activeGroup.id, channel.id);
    } catch (err) {
      alert(err.message || 'Không thể xóa kênh');
    }
  }

  async function handleDeleteCategory(category, e) {
    e.stopPropagation();
    if (!confirm(`Bạn có chắc chắn muốn xóa danh mục "${category.name}"?`)) {
      return;
    }
    try {
      await deleteCategory($activeGroup.id, category.id);
    } catch (err) {
      alert(err.message || 'Không thể xóa danh mục');
    }
  }
</script>

<aside class="channel-sidebar">
  <!-- Group Header Banner -->
  <div class="group-header">
    <div class="header-left">
      <img
        src={$activeGroup?.avatar || `https://api.dicebear.com/7.x/identicon/svg?seed=${$activeGroup?.id}`}
        alt="avatar"
        class="group-badge"
      />
      <div class="header-info">
        <h2 class="group-name" title={$activeGroup?.name}>{$activeGroup?.name || 'Cộng đồng'}</h2>
        <span class="group-members-count">
          {$activeGroup?.member_count || $activeGroup?.members?.length || 0} thành viên
        </span>
      </div>
    </div>
    <div class="header-tools">
      <button class="tool-btn" on:click={onOpenMembersModal} title="Quản lý thành viên & Cài đặt">
        <Settings size={16} />
      </button>
      {#if amIAdmin}
        <button class="tool-btn add-btn" on:click={() => openCreateModal('')} title="Thêm Kênh / Danh Mục mới">
          <Plus size={16} />
        </button>
      {/if}
    </div>
  </div>

  <!-- Group Events Banner / Trigger -->
  <button class="events-trigger-btn" on:click={() => (showEventsModal = true)}>
    <div class="events-trigger-left">
      <Calendar size={14} class="event-cal-icon" />
      <span>Sự kiện nhóm</span>
    </div>
    {#if $groupEvents.length > 0}
      <span class="events-count-badge">{$groupEvents.length}</span>
    {/if}
  </button>

  <!-- Channel Navigation Tree -->
  <div class="channel-tree">
    <!-- Uncategorized Channels nếu có -->
    {#if uncategorizedChannels.length > 0}
      <div class="category-block">
        <div class="category-header">
          <span class="cat-title">KÊNH CHUNG</span>
        </div>
        <div class="channel-list">
          {#each uncategorizedChannels as chan}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div
              class="channel-item"
              class:active={$activeChannel?.id === chan.id}
              class:announcement={chan.type === 'announcement'}
              on:click={() => selectChannel(chan)}
            >
              <div class="chan-left">
                {#if chan.type === 'announcement'}
                  <Megaphone size={15} class="chan-icon announcement-icon" />
                {:else}
                  <Hash size={15} class="chan-icon" />
                {/if}
                <span class="chan-name">{chan.name}</span>
              </div>
              <div class="chan-right">
                {#if ($channelUnreadCounts[chan.id] || 0) > 0}
                  <span class="unread-pill">{$channelUnreadCounts[chan.id]}</span>
                {/if}
                {#if amIAdmin && uncategorizedChannels.length + channels.length > 1}
                  <button
                    class="chan-action-btn"
                    on:click={(e) => handleDeleteChannel(chan, e)}
                    title="Xóa kênh"
                  >
                    <Trash2 size={13} />
                  </button>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Categorized Channels -->
    {#each categorizedChannels as item}
      {@const isCollapsed = collapsedCategories[item.category.id]}
      <div class="category-block">
        <div class="category-header" on:click={() => toggleCollapse(item.category.id)}>
          <div class="cat-left">
            {#if isCollapsed}
              <ChevronRight size={13} class="arrow-icon" />
            {:else}
              <ChevronDown size={13} class="arrow-icon" />
            {/if}
            <span class="cat-title">{item.category.name}</span>
          </div>

          {#if amIAdmin}
            <div class="cat-actions" on:click|stopPropagation>
              <button
                class="cat-action-btn"
                on:click={() => openCreateModal(item.category.id)}
                title="Tạo kênh trong danh mục này"
              >
                <Plus size={14} />
              </button>
              {#if item.channels.length === 0}
                <button
                  class="cat-action-btn danger"
                  on:click={(e) => handleDeleteCategory(item.category, e)}
                  title="Xóa danh mục trống"
                >
                  <Trash2 size={13} />
                </button>
              {/if}
            </div>
          {/if}
        </div>

        {#if !isCollapsed}
          <div class="channel-list">
            {#if item.channels.length === 0}
              <div class="empty-channel-hint">Chưa có kênh nào</div>
            {:else}
              {#each item.channels as chan}
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  class="channel-item"
                  class:active={$activeChannel?.id === chan.id}
                  class:announcement={chan.type === 'announcement'}
                  on:click={() => selectChannel(chan)}
                >
                  <div class="chan-left">
                    {#if chan.type === 'announcement'}
                      <Megaphone size={15} class="chan-icon announcement-icon" />
                    {:else}
                      <Hash size={15} class="chan-icon" />
                    {/if}
                    <span class="chan-name">{chan.name}</span>
                  </div>
                  <div class="chan-right">
                    {#if ($channelUnreadCounts[chan.id] || 0) > 0}
                      <span class="unread-pill">{$channelUnreadCounts[chan.id]}</span>
                    {/if}
                    {#if amIAdmin && channels.length > 1}
                      <button
                        class="chan-action-btn"
                        on:click={(e) => handleDeleteChannel(chan, e)}
                        title="Xóa kênh"
                      >
                        <Trash2 size={13} />
                      </button>
                    {/if}
                  </div>
                </div>
              {/each}
            {/if}
          </div>
        {/if}
      </div>
    {/each}
  </div>

  <!-- Footer Quick Bar -->
  <div class="sidebar-footer">
    <button class="footer-btn" on:click={onOpenMembersModal}>
      <Users size={15} />
      <span>Thành viên ({$activeGroup?.member_count || $activeGroup?.members?.length || 0})</span>
    </button>
  </div>
</aside>

{#if showCreateModal}
  <CreateChannelModal
    initialCategoryID={targetCategoryID}
    onClose={() => (showCreateModal = false)}
    onCreated={() => {}}
  />
{/if}

{#if showEventsModal && $activeGroup}
  <GroupEventsModal
    groupID={$activeGroup.id}
    {channels}
    onClose={() => (showEventsModal = false)}
  />
{/if}

<style>
  .channel-sidebar {
    width: 240px;
    height: 100vh;
    background: rgba(10, 14, 26, 0.95);
    border-right: 1px solid var(--border-glass);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    user-select: none;
  }

  /* Header */
  .group-header {
    height: 64px;
    padding: 0 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--border-glass);
    background: rgba(255, 255, 255, 0.02);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 10px;
    overflow: hidden;
  }

  .group-badge {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    object-fit: cover;
    flex-shrink: 0;
  }

  .header-info {
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .group-name {
    font-size: 14px;
    font-weight: 700;
    color: #fff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin: 0;
  }

  .group-members-count {
    font-size: 11px;
    color: var(--text-dim);
  }

  .header-tools {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .events-trigger-btn {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 8px 12px 4px 12px;
    padding: 7px 10px;
    background: rgba(16, 185, 129, 0.08);
    border: 1px solid rgba(16, 185, 129, 0.2);
    border-radius: 8px;
    color: #cbd5e1;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.15s;
  }
  .events-trigger-btn:hover {
    background: rgba(16, 185, 129, 0.15);
    color: #fff;
    border-color: rgba(16, 185, 129, 0.35);
  }
  .events-trigger-left {
    display: flex;
    align-items: center;
    gap: 7px;
  }
  .event-cal-icon {
    color: #34d399;
  }
  .events-count-badge {
    background: #10b981;
    color: #fff;
    font-size: 10px;
    font-weight: 700;
    padding: 1px 6px;
    border-radius: 10px;
  }

  .tool-btn {
    width: 28px;
    height: 28px;
    border-radius: 6px;
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.15s;
  }

  .tool-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .add-btn:hover {
    color: #38bdf8;
    background: rgba(56, 189, 248, 0.15);
  }

  /* Navigation Tree */
  .channel-tree {
    flex: 1;
    overflow-y: auto;
    padding: 14px 8px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .category-block {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .category-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 8px;
    cursor: pointer;
    border-radius: 4px;
    transition: color 0.15s;
  }

  .cat-left {
    display: flex;
    align-items: center;
    gap: 6px;
    overflow: hidden;
  }

  :global(.arrow-icon) {
    color: var(--text-muted);
  }

  .cat-title {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.6px;
    color: var(--text-muted);
    text-transform: uppercase;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .category-header:hover .cat-title {
    color: var(--text-primary);
  }

  .cat-actions {
    display: flex;
    align-items: center;
    gap: 2px;
  }

  .cat-action-btn {
    width: 20px;
    height: 20px;
    border-radius: 4px;
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: 0.15s;
  }

  .category-header:hover .cat-action-btn {
    opacity: 1;
  }

  .cat-action-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.15);
  }

  .cat-action-btn.danger:hover {
    color: #f87171;
    background: rgba(239, 68, 68, 0.2);
  }

  /* Channel Item */
  .channel-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: 2px;
  }

  .empty-channel-hint {
    padding: 4px 14px;
    font-size: 11px;
    font-style: italic;
    color: var(--text-muted);
  }

  .channel-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 7px 10px;
    border-radius: 6px;
    cursor: pointer;
    color: var(--text-dim);
    font-size: 13px;
    font-weight: 500;
    transition: background 0.15s, color 0.15s;
  }

  .channel-item:hover {
    background: rgba(255, 255, 255, 0.05);
    color: #e2e8f0;
  }

  .channel-item.active {
    background: rgba(56, 189, 248, 0.16);
    color: #fff;
    font-weight: 600;
  }

  .chan-left {
    display: flex;
    align-items: center;
    gap: 8px;
    overflow: hidden;
  }

  :global(.chan-icon) {
    color: var(--text-muted);
    flex-shrink: 0;
  }

  .channel-item:hover :global(.chan-icon),
  .channel-item.active :global(.chan-icon) {
    color: #38bdf8;
  }

  :global(.announcement-icon) {
    color: #f59e0b !important;
  }

  .chan-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chan-right {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .unread-pill {
    padding: 1px 6px;
    border-radius: 10px;
    background: #ef4444;
    color: #fff;
    font-size: 11px;
    font-weight: 700;
  }

  .chan-action-btn {
    width: 20px;
    height: 20px;
    border-radius: 4px;
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    display: none;
    align-items: center;
    justify-content: center;
  }

  .channel-item:hover .chan-action-btn {
    display: flex;
  }

  .chan-action-btn:hover {
    color: #f87171;
    background: rgba(239, 68, 68, 0.2);
  }

  /* Footer */
  .sidebar-footer {
    padding: 10px 12px;
    border-top: 1px solid var(--border-glass);
    background: rgba(0, 0, 0, 0.2);
  }

  .footer-btn {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 8px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-glass);
    color: var(--text-dim);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: 0.15s;
  }

  .footer-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }
</style>
