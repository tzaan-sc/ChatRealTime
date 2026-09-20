<script>
  import { onDestroy } from 'svelte';
  import { token, currentUser } from './lib/stores/auth';
  import { wsService, incomingMessages } from './lib/services/websocket';
  import {
    loadConversations,
    loadUserGroups,
    appendMessage,
    appendGroupMessage,
    onlineUsers,
    setUserOnlineStatus,
    setUserTyping,
    markMessagesAsReadLocally,
    updateMessageReactionsLocally,
    markMessageDeletedLocally,
    updateMessageEditedLocally
  } from './lib/stores/chat';
  import { playNotificationSound } from './lib/utils/sound';
  import AuthModal from './lib/components/AuthModal.svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import ChatArea from './lib/components/ChatArea.svelte';

  let heartbeatInterval;

  // 1. Tự động kết nối WebSocket và thiết lập Heartbeat định kỳ
  $: if ($token) {
    wsService.connect($token);
    loadConversations();
    loadUserGroups();

    clearInterval(heartbeatInterval);
    heartbeatInterval = setInterval(() => {
      wsService.send('heartbeat', {});
    }, 15000); // Mỗi 15 giây gửi 1 lần
  }

  onDestroy(() => {
    clearInterval(heartbeatInterval);
  });

  // 2. Lắng nghe và điều phối các sự kiện Real-time
  $: if ($incomingMessages) {
    const { event, payload } = $incomingMessages;

    switch (event) {
      case 'chat:receive':
        appendMessage(payload);
        // Phát âm thanh thông báo nếu tin nhắn đến từ người khác
        if (payload.sender_id !== $currentUser?.id) {
          playNotificationSound();
        }
        break;

      case 'chat:ack':
        appendMessage(payload.message);
        break;

      case 'group:receive':
        appendGroupMessage(payload);
        if (payload.sender_id !== $currentUser?.id) {
          playNotificationSound();
        }
        break;

      case 'group:ack':
        appendGroupMessage(payload.message);
        break;


      case 'chat:reaction_updated':
        if (payload?.message_id) {
          updateMessageReactionsLocally(payload.message_id, payload.reactions);
        }
        break;

      case 'chat:message_deleted':
        if (payload?.message_id) {
          markMessageDeletedLocally(payload.message_id);
        }
        break;

      case 'chat:message_edited':
        if (payload?.message_id) {
          updateMessageEditedLocally(payload.message_id, payload.content);
        }
        break;

      case 'user:online_list':
        // Danh sách ID đang online khi vừa vào app
        if (Array.isArray(payload)) {
          onlineUsers.set(new Set(payload));
        }
        break;

      case 'user:status':
        // Cập nhật người vừa online / offline
        if (payload?.user_id) {
          setUserOnlineStatus(payload.user_id, payload.is_online);
        }
        break;

      case 'typing:start':
        if (payload?.sender_id) {
          setUserTyping(payload.sender_id, true);
        }
        break;

      case 'typing:stop':
        if (payload?.sender_id) {
          setUserTyping(payload.sender_id, false);
        }
        break;

      case 'chat:read_ack':
        // Đánh dấu đã đọc thành 2 tích xanh
        if (payload?.conversation_id) {
          markMessagesAsReadLocally(payload.conversation_id);
        }
        break;
    }
  }
</script>

<div class="app-layout">
  {#if !$currentUser}
    <AuthModal />
  {:else}
    <Sidebar />
    <ChatArea />
  {/if}
</div>

<style>
  .app-layout {
    display: flex;
    width: 100vw;
    height: 100vh;
    background: var(--bg-primary);
  }
</style>
