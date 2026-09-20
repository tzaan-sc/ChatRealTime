<script>
  import { onMount } from 'svelte';
  import { token, currentUser } from './lib/stores/auth';
  import { wsService, incomingMessages } from './lib/services/websocket';
  import { loadConversations, appendMessage } from './lib/stores/chat';
  import AuthModal from './lib/components/AuthModal.svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import ChatArea from './lib/components/ChatArea.svelte';

  // Tự động kết nối WebSocket và nạp dữ liệu khi mở trang
  $: if ($token) {
    wsService.connect($token);
    loadConversations();
  }

  // Lắng nghe sự kiện từ WebSocket
  $: if ($incomingMessages) {
    const { event, payload } = $incomingMessages;
    if (event === 'chat:receive' || event === 'chat:ack') {
      const msg = event === 'chat:receive' ? payload : payload.message;
      appendMessage(msg);
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
