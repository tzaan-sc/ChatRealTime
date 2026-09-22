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
    updateMessageEditedLocally,
    conversations,
    userGroups,
    selectConversation,
    selectGroup,
    activeConversation,
    activeGroup
  } from './lib/stores/chat';
  import { playNotificationSound } from './lib/utils/sound';
  import { addToast } from './lib/stores/notification';
  import { showDesktopNotification, startFlashingTitle } from './lib/services/notification';
  import AuthModal from './lib/components/AuthModal.svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import ChatArea from './lib/components/ChatArea.svelte';
  import CallModal from './lib/components/CallModal.svelte';
  import NotificationToast from './lib/components/NotificationToast.svelte';
  import {
    handleIncomingCall,
    handleCallAccepted,
    handleReceiveOffer,
    handleReceiveAnswer,
    handleReceiveIceCandidate,
    handleRemoteHangup
  } from './lib/stores/call';

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
      case 'chat:receive': {
        const isFromMe = payload.sender_id === $currentUser?.id;
        const active = $activeConversation;
        const isViewingThisChat = active && active.conversation?.custom_id === payload.conversation_id;
        const isTabVisible = typeof document !== 'undefined' && document.visibilityState === 'visible';

        appendMessage(payload);

        if (!isFromMe) {
          if (isViewingThisChat && isTabVisible) {
            playNotificationSound({ subtle: true });
          } else {
            playNotificationSound({ subtle: false });

            const conv = $conversations.find((c) => c.conversation?.custom_id === payload.conversation_id);
            const senderName = conv?.other_user?.display_name || conv?.other_user?.username || 'Tin nhắn mới';
            const senderAvatar = conv?.other_user?.avatar_url || '';

            if (isTabVisible && !isViewingThisChat) {
              addToast({
                title: 'Tin nhắn cá nhân',
                senderName,
                senderAvatar,
                content: payload.content || (payload.type === 'image' ? '[Hình ảnh]' : '[Tệp tin]'),
                conversationId: payload.conversation_id
              });
            }

            if (!isTabVisible) {
              startFlashingTitle(senderName, payload.content);
              showDesktopNotification({
                title: `💬 ${senderName}`,
                body: payload.content || (payload.type === 'image' ? '[Hình ảnh]' : '[Tệp tin]'),
                icon: senderAvatar,
                tag: payload.conversation_id,
                onClick: () => {
                  if (conv) selectConversation(conv);
                }
              });
            }
          }
        }
        break;
      }

      case 'chat:ack':
        appendMessage(payload.message);
        break;

      case 'group:receive': {
        const isFromMe = payload.sender_id === $currentUser?.id;
        const grp = $activeGroup;
        const isViewingThisGroup = grp && grp.id === payload.group_id;
        const isTabVisible = typeof document !== 'undefined' && document.visibilityState === 'visible';

        appendGroupMessage(payload);

        if (!isFromMe) {
          if (isViewingThisGroup && isTabVisible) {
            playNotificationSound({ subtle: true });
          } else {
            playNotificationSound({ subtle: false });

            const targetGroup = $userGroups.find((g) => g.id === payload.group_id);
            const groupName = targetGroup?.name || 'Nhóm';
            const groupAvatar = targetGroup?.avatar || '';
            const senderTitle = `${payload.sender_name || 'Thành viên'} trong ${groupName}`;

            if (isTabVisible && !isViewingThisGroup) {
              addToast({
                title: groupName,
                senderName: payload.sender_name || 'Thành viên',
                senderAvatar: payload.sender_avatar || groupAvatar,
                content: payload.content || (payload.type === 'image' ? '[Hình ảnh]' : '[Tệp tin]'),
                groupId: payload.group_id
              });
            }

            if (!isTabVisible) {
              startFlashingTitle(senderTitle, payload.content);
              showDesktopNotification({
                title: `👥 ${groupName}`,
                body: `${payload.sender_name || 'Thành viên'}: ${payload.content || '[Nội dung tin nhắn]'}`,
                icon: payload.sender_avatar || groupAvatar,
                tag: payload.group_id,
                onClick: () => {
                  if (targetGroup) selectGroup(targetGroup);
                }
              });
            }
          }
        }
        break;
      }

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

      // 3. WebRTC Signaling Events
      case 'call:request':
        handleIncomingCall(payload);
        break;

      case 'call:accept':
        handleCallAccepted();
        break;

      case 'call:reject':
        handleRemoteHangup(payload);
        break;

      case 'call:offer':
        handleReceiveOffer(payload);
        break;

      case 'call:answer':
        handleReceiveAnswer(payload);
        break;

      case 'call:ice_candidate':
        handleReceiveIceCandidate(payload);
        break;

      case 'call:hangup':
        handleRemoteHangup(payload);
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
    <CallModal />
    <NotificationToast />
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
