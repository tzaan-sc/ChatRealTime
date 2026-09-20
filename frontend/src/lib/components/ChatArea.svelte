<script>
  import { activeConversation, activeGroup, currentMessages, onlineUsers, typingUsers } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import { wsService } from '../services/websocket';
  import { afterUpdate } from 'svelte';
  import {
    Send,
    Paperclip,
    Mic,
    Trash2,
    Check,
    FileText,
    Download,
    Loader2,
    Reply,
    Edit2,
    X,
    Smile,
    Users
  } from 'lucide-svelte';
  import ImageModal from './ImageModal.svelte';
  import AudioPlayer from './AudioPlayer.svelte';
  import GroupMembersModal from './GroupMembersModal.svelte';

  const QUICK_EMOJIS = ['❤️', '😂', '👍', '😢', '🔥', '🚀'];

  let inputContent = '';
  let messagesContainer;
  let typingTimeout;
  let isTyping = false;
  let fileInput;

  // Trạng thái Upload & Lightbox
  let isUploading = false;
  let uploadProgressText = '';
  let selectedImage = null;

  // Trạng thái Ghi âm Voice Note
  let isRecording = false;
  let recordingTime = 0;
  let recordingTimer = null;
  let mediaRecorder = null;
  let audioChunks = [];
  let audioStream = null;

  // Trạng thái Trả lời (Reply) & Chỉnh sửa (Edit)
  let replyingTo = null; // { message_id, sender_name, content }
  let editingMessage = null; // message object

  // Trạng thái Modal Thành viên nhóm
  let showGroupMembersModal = false;

  $: isGroup = !!$activeGroup;
  $: partnerID = $activeConversation?.other_user?.id;
  $: isPartnerOnline = partnerID ? $onlineUsers.has(partnerID) : false;
  $: isPartnerTyping = partnerID ? $typingUsers.has(partnerID) : false;

  // Tự động cuộn xuống đáy khi có tin nhắn mới (chỉ cuộn nếu không đang xem tin cũ)
  afterUpdate(() => {
    if (messagesContainer && !editingMessage && !replyingTo) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  });

  // Bắt sự kiện gõ phím để bắn Typing Indicator
  function handleInput() {
    if (isGroup || !$activeConversation || !partnerID || editingMessage) return;

    if (!isTyping) {
      isTyping = true;
      wsService.send('typing:start', { receiver_id: partnerID });
    }

    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => {
      isTyping = false;
      wsService.send('typing:stop', { receiver_id: partnerID });
    }, 2000);
  }

  // Gửi tin nhắn chữ hoặc lưu sửa tin nhắn
  function handleSend() {
    if (!inputContent.trim()) return;
    if (!isGroup && (!$activeConversation || !partnerID)) return;
    if (isGroup && !$activeGroup) return;

    // Nếu đang trong chế độ chỉnh sửa tin nhắn
    if (editingMessage) {
      if (isGroup) {
        wsService.send('chat:edit', {
          message_id: editingMessage.id,
          group_id: $activeGroup.id,
          content: inputContent.trim()
        });
      } else {
        wsService.send('chat:edit', {
          message_id: editingMessage.id,
          conversation_id: $activeConversation.conversation.custom_id,
          receiver_id: partnerID,
          content: inputContent.trim()
        });
      }
      cancelEdit();
      return;
    }

    // Dừng typing khi gửi tin (với 1-1)
    if (!isGroup) {
      clearTimeout(typingTimeout);
      if (isTyping) {
        isTyping = false;
        wsService.send('typing:stop', { receiver_id: partnerID });
      }
    }

    if (isGroup) {
      const payload = {
        group_id: $activeGroup.id,
        content: inputContent.trim(),
        type: 'text'
      };

      if (replyingTo) {
        payload.reply_to = {
          message_id: replyingTo.message_id,
          sender_name: replyingTo.sender_name,
          content: replyingTo.content
        };
        replyingTo = null;
      }

      wsService.send('group:send', payload);
    } else {
      const payload = {
        receiver_id: partnerID,
        content: inputContent.trim(),
        type: 'text'
      };

      if (replyingTo) {
        payload.reply_to = {
          message_id: replyingTo.message_id,
          sender_name: replyingTo.sender_name,
          content: replyingTo.content
        };
        replyingTo = null;
      }

      wsService.send('chat:send', payload);
    }

    inputContent = '';
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  // Upload và gửi tin nhắn tệp/ảnh/voice
  async function uploadAndSend(file) {
    if (!file) return;
    if (!isGroup && (!$activeConversation || !partnerID)) return;
    if (isGroup && !$activeGroup) return;

    if (file.size > 25 * 1024 * 1024) {
      alert('Dung lượng tệp vượt quá giới hạn 25MB');
      return;
    }

    isUploading = true;
    uploadProgressText = `Đang tải lên ${file.name || 'tệp'}...`;

    try {
      const formData = new FormData();
      formData.append('file', file);

      const token = localStorage.getItem('token');
      const res = await fetch('http://localhost:8080/api/upload', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`
        },
        body: formData
      });

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Lỗi tải tệp lên máy chủ');
      }

      const data = await res.json();

      if (isGroup) {
        const payload = {
          group_id: $activeGroup.id,
          content: data.file_url,
          type: data.type,
          file_name: data.file_name,
          file_size: data.file_size
        };

        if (replyingTo) {
          payload.reply_to = {
            message_id: replyingTo.message_id,
            sender_name: replyingTo.sender_name,
            content: replyingTo.content
          };
          replyingTo = null;
        }

        wsService.send('group:send', payload);
      } else {
        const payload = {
          receiver_id: partnerID,
          content: data.file_url,
          type: data.type,
          file_name: data.file_name,
          file_size: data.file_size
        };

        if (replyingTo) {
          payload.reply_to = {
            message_id: replyingTo.message_id,
            sender_name: replyingTo.sender_name,
            content: replyingTo.content
          };
          replyingTo = null;
        }

        wsService.send('chat:send', payload);
      }
    } catch (err) {
      console.error('Lỗi upload:', err);
      alert('Không thể gửi tệp: ' + err.message);
    } finally {
      isUploading = false;
      uploadProgressText = '';
      if (fileInput) fileInput.value = '';
    }
  }

  function handleFileInputChange(e) {
    const files = e.target.files;
    if (files && files.length > 0) {
      uploadAndSend(files[0]);
    }
  }

  // Dán ảnh trực tiếp từ clipboard (Ctrl + V)
  function handlePaste(e) {
    if (!e.clipboardData || !e.clipboardData.items) return;
    const items = e.clipboardData.items;
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf('image') !== -1) {
        e.preventDefault();
        const file = items[i].getAsFile();
        if (file) {
          uploadAndSend(file);
        }
        return;
      }
    }
  }

  // Ghi âm Voice Note
  async function startRecording() {
    try {
      audioChunks = [];
      audioStream = await navigator.mediaDevices.getUserMedia({ audio: true });
      mediaRecorder = new MediaRecorder(audioStream);

      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          audioChunks.push(e.data);
        }
      };

      mediaRecorder.start();
      isRecording = true;
      recordingTime = 0;
      recordingTimer = setInterval(() => {
        recordingTime += 1;
      }, 1000);
    } catch (err) {
      console.error('Lỗi mở microphone:', err);
      alert('Không thể kích hoạt microphone. Hãy đảm bảo bạn đã cho phép quyền Micro trên trình duyệt.');
    }
  }

  function cancelRecording() {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.stop();
    }
    cleanupRecording();
  }

  function finishRecording() {
    if (!mediaRecorder || mediaRecorder.state === 'inactive') return;

    mediaRecorder.onstop = () => {
      const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
      const voiceFile = new File([audioBlob], `voice_${Date.now()}.webm`, { type: 'audio/webm' });
      cleanupRecording();
      uploadAndSend(voiceFile);
    };

    mediaRecorder.stop();
  }

  function cleanupRecording() {
    if (audioStream) {
      audioStream.getTracks().forEach((t) => t.stop());
      audioStream = null;
    }
    clearInterval(recordingTimer);
    recordingTimer = null;
    isRecording = false;
    recordingTime = 0;
    mediaRecorder = null;
  }

  // Chức năng Trả lời (Reply)
  function startReply(msg) {
    let snippet = msg.content;
    if (msg.type === 'image') snippet = '🖼️ [Hình ảnh]';
    else if (msg.type === 'voice') snippet = '🎙️ [Tin nhắn thoại]';
    else if (msg.type === 'file') snippet = `📎 [Tệp] ${msg.file_name || ''}`;

    let senderName = 'Bạn';
    if (msg.sender_id !== $currentUser?.id) {
      if (isGroup) {
        senderName = msg.sender_name || 'Thành viên';
      } else {
        senderName =
          $activeConversation?.other_user?.display_name ||
          $activeConversation?.other_user?.username ||
          'Đối phương';
      }
    }

    replyingTo = {
      message_id: msg.id,
      sender_name: senderName,
      content: snippet
    };

    // Đóng chế độ edit nếu đang mở
    editingMessage = null;
  }

  function cancelReply() {
    replyingTo = null;
  }

  // Chức năng Chỉnh sửa (Edit)
  function startEdit(msg) {
    if (msg.type !== 'text' || msg.is_deleted) return;
    editingMessage = msg;
    inputContent = msg.content;
    replyingTo = null;
  }

  function cancelEdit() {
    editingMessage = null;
    inputContent = '';
  }

  // Chức năng Thu hồi (Delete / Unsend)
  function deleteMessage(msg) {
    if (!confirm('Bạn có chắc chắn muốn thu hồi tin nhắn này với mọi người?')) return;
    if (isGroup) {
      wsService.send('chat:delete', {
        message_id: msg.id,
        group_id: $activeGroup.id
      });
    } else {
      wsService.send('chat:delete', {
        message_id: msg.id,
        conversation_id: $activeConversation.conversation.custom_id,
        receiver_id: partnerID
      });
    }
  }

  // Thả reaction
  function toggleReaction(msg, emoji) {
    if (!msg || !msg.id) return;
    if (isGroup) {
      if (!$activeGroup) return;
      wsService.send('chat:react', {
        message_id: msg.id,
        group_id: $activeGroup.id,
        emoji: emoji
      });
    } else {
      if (!$activeConversation || !partnerID) return;
      wsService.send('chat:react', {
        message_id: msg.id,
        conversation_id: $activeConversation.conversation.custom_id,
        receiver_id: partnerID,
        emoji: emoji
      });
    }
  }

  // Tính toán nhóm reaction
  function getGroupedReactions(reactions) {
    if (!reactions || reactions.length === 0) return [];
    const counts = {};
    for (const r of reactions) {
      if (!counts[r.emoji]) {
        counts[r.emoji] = { emoji: r.emoji, count: 0, hasMe: false };
      }
      counts[r.emoji].count++;
      if (r.user_id === $currentUser?.id) {
        counts[r.emoji].hasMe = true;
      }
    }
    return Object.values(counts);
  }

  // Cuộn tới tin nhắn gốc khi click vào trích dẫn
  function scrollToMessage(messageId) {
    if (!messageId) return;
    const el = document.getElementById(`msg-${messageId}`);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      el.classList.add('highlight-target');
      setTimeout(() => el.classList.remove('highlight-target'), 1800);
    }
  }

  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins < 10 ? '0' : ''}${mins}:${secs < 10 ? '0' : ''}${secs}`;
  }

  function formatFileSize(bytes) {
    if (!bytes || bytes === 0) return '';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  // Gửi event xác nhận đã đọc khi đang mở xem khung chat (1-1)
  $: if (!isGroup && $activeConversation && partnerID) {
    wsService.send('chat:read', {
      conversation_id: $activeConversation.conversation.custom_id,
      partner_id: partnerID
    });
  }
</script>

{#if !$activeConversation && !$activeGroup}
  <div class="empty-chat glass-card">
    <div class="empty-content">
      <div class="icon-circle">💬</div>
      <h2>Chào mừng đến với Realtime Chat!</h2>
      <p>Chọn một người bạn hoặc một nhóm trò chuyện ở danh sách bên trái để bắt đầu.</p>
    </div>
  </div>
{:else}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <main class="chat-area" on:paste={handlePaste}>
    <!-- Header -->
    <div class="chat-header glass-card">
      {#if isGroup}
        <div class="partner-info">
          <div class="avatar-wrap">
            <img
              src={$activeGroup.avatar || `https://api.dicebear.com/7.x/identicon/svg?seed=${$activeGroup.id}`}
              alt="avatar"
              class="header-avatar"
            />
          </div>
          <div>
            <h3>{$activeGroup.name}</h3>
            <span class="status-sub">
              <span class="online-text">{$activeGroup.member_count || $activeGroup.members?.length || 0} thành viên</span>
            </span>
          </div>
        </div>
        <div class="header-actions">
          <button class="header-action-btn" on:click={() => (showGroupMembersModal = true)} title="Xem thành viên & Quản lý nhóm">
            <Users size={19} />
          </button>
        </div>
      {:else}
        <div class="partner-info">
          <div class="avatar-wrap">
            <img src={$activeConversation.other_user.avatar_url} alt="avatar" class="header-avatar" />
            <span class="status-dot" class:online={isPartnerOnline}></span>
          </div>
          <div>
            <h3>{$activeConversation.other_user.display_name || $activeConversation.other_user.username}</h3>
            <span class="status-sub">
              {#if isPartnerTyping}
                <span class="typing-text">đang soạn tin...</span>
              {:else if isPartnerOnline}
                <span class="online-text">Đang hoạt động</span>
              {:else}
                <span class="offline-text">Ngoại tuyến</span>
              {/if}
            </span>
          </div>
        </div>
      {/if}
    </div>

    <!-- Messages Body -->
    <div class="messages-viewport" bind:this={messagesContainer}>
      {#each $currentMessages as msg (msg.id)}
        {@const isMe = msg.sender_id === $currentUser.id}
        {@const groupedReactions = getGroupedReactions(msg.reactions)}
        <div class="message-row" class:me={isMe} id={`msg-${msg.id}`}>
          <!-- Thanh Công Cụ Tương Tác Nổi (Reaction Bar & Actions Toolbar) -->
          {#if !msg.is_deleted}
            <div class="msg-toolbar glass-card" class:toolbar-me={isMe}>
              <!-- 6 Emoji Phổ Biến -->
              <div class="quick-emojis">
                {#each QUICK_EMOJIS as emo}
                  <button
                    class="emoji-btn"
                    on:click={() => toggleReaction(msg, emo)}
                    title={`Thả ${emo}`}
                  >
                    {emo}
                  </button>
                {/each}
              </div>

              <div class="toolbar-divider"></div>

              <!-- Nút Trả lời -->
              <button class="tool-btn" on:click={() => startReply(msg)} title="Trả lời tin nhắn">
                <Reply size={15} />
              </button>

              <!-- Nút Sửa & Thu hồi (Chỉ hiện với tin của mình) -->
              {#if isMe}
                {#if msg.type === 'text'}
                  <button class="tool-btn" on:click={() => startEdit(msg)} title="Chỉnh sửa">
                    <Edit2 size={14} />
                  </button>
                {/if}
                <button class="tool-btn danger" on:click={() => deleteMessage(msg)} title="Thu hồi tin nhắn">
                  <Trash2 size={14} />
                </button>
              {/if}
            </div>
          {/if}

          <!-- Bong Bóng Tin Nhắn Chính -->
          <div class="bubble" class:bubble-me={isMe} class:bubble-other={!isMe}>
            <!-- Tên & Avatar thành viên gửi trong nhóm chat -->
            {#if isGroup && !isMe}
              <div class="group-sender-header">
                {#if msg.sender_avatar}
                  <img src={msg.sender_avatar} alt="avatar" class="bubble-sender-avatar" />
                {/if}
                <span class="bubble-sender-name">{msg.sender_name || 'Thành viên'}</span>
              </div>
            {/if}

            <!-- Khung Trích Dẫn Tin Nhắn Cũ (Reply Card) -->
            {#if msg.reply_to}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <div
                class="reply-quote"
                role="button"
                tabindex="0"
                on:click={() => scrollToMessage(msg.reply_to.message_id)}
                title="Bấm để cuộn tới tin gốc"
              >
                <span class="quote-sender">{msg.reply_to.sender_name}</span>
                <span class="quote-text">{msg.reply_to.content}</span>
              </div>
            {/if}

            <!-- Nội dung tin nhắn theo trạng thái / loại -->
            {#if msg.is_deleted}
              <div class="deleted-msg">
                <span class="deleted-icon">🚫</span>
                <em>Tin nhắn đã được thu hồi</em>
              </div>
            {:else if msg.type === 'image'}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <div
                class="image-bubble"
                role="button"
                tabindex="0"
                on:click={() => (selectedImage = { url: msg.content, name: msg.file_name || 'image' })}
              >
                <img src={msg.content} alt={msg.file_name || 'ảnh gửi'} loading="lazy" class="chat-image" />
              </div>
            {:else if msg.type === 'file'}
              <div class="file-card">
                <div class="file-icon-box">
                  <FileText size={22} />
                </div>
                <div class="file-info">
                  <span class="file-name" title={msg.file_name}>{msg.file_name || 'Tệp đính kèm'}</span>
                  {#if msg.file_size}
                    <span class="file-size">{formatFileSize(msg.file_size)}</span>
                  {/if}
                </div>
                <a
                  href={msg.content}
                  download={msg.file_name || 'file'}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="file-dl-btn"
                  title="Tải tệp"
                >
                  <Download size={16} />
                </a>
              </div>
            {:else if msg.type === 'voice'}
              <AudioPlayer src={msg.content} />
            {:else}
              <p class="text">{msg.content}</p>
            {/if}

            <!-- Thông Tin Phụ: Giờ, Đã sửa, Tích đọc -->
            <div class="msg-meta">
              {#if msg.is_edited && !msg.is_deleted}
                <span class="edited-tag" title="Tin nhắn đã qua chỉnh sửa">(đã sửa)</span>
              {/if}
              <span class="timestamp">
                {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
              {#if isMe && !msg.is_deleted}
                <span class="receipt-icon" class:seen={msg.is_read} title={msg.is_read ? 'Đã xem' : 'Đã gửi'}>
                  {msg.is_read ? '✓✓' : '✓'}
                </span>
              {/if}
            </div>

            <!-- Gom Nhóm Reaction Dưới Bong Bóng -->
            {#if groupedReactions.length > 0 && !msg.is_deleted}
              <div class="reactions-badge-container">
                {#each groupedReactions as r}
                  <button
                    class="reaction-pill"
                    class:reacted-by-me={r.hasMe}
                    on:click={() => toggleReaction(msg, r.emoji)}
                    title={r.hasMe ? 'Bấm để hủy reaction' : 'Bấm để thả cùng reaction'}
                  >
                    <span class="pill-emoji">{r.emoji}</span>
                    <span class="pill-count">{r.count}</span>
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      {/each}

      <!-- Animation Đang Gõ Phím (Typing Dots) -->
      {#if isPartnerTyping}
        <div class="message-row">
          <div class="typing-bubble glass-card">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
          </div>
        </div>
      {/if}
    </div>

    <!-- Thông báo đang tải file lên -->
    {#if isUploading}
      <div class="upload-indicator">
        <Loader2 size={16} class="spinner" />
        <span>{uploadProgressText}</span>
      </div>
    {/if}

    <!-- Input Footer & Action Preview Bar -->
    <div class="chat-footer glass-card">
      <input
        type="file"
        bind:this={fileInput}
        on:change={handleFileInputChange}
        style="display: none;"
      />

      <!-- Thanh Banner Xem Trước Khi Đang Trả Lời (Reply Bar) -->
      {#if replyingTo}
        <div class="context-banner reply-banner">
          <div class="banner-content">
            <Reply size={15} class="banner-icon" />
            <div class="banner-text">
              <span class="banner-title">Trả lời <strong>{replyingTo.sender_name}</strong></span>
              <span class="banner-sub">{replyingTo.content}</span>
            </div>
          </div>
          <button class="banner-close" on:click={cancelReply} title="Hủy trả lời">
            <X size={16} />
          </button>
        </div>
      {/if}

      <!-- Thanh Banner Xem Trước Khi Đang Sửa Tin Nhắn (Edit Bar) -->
      {#if editingMessage}
        <div class="context-banner edit-banner">
          <div class="banner-content">
            <Edit2 size={15} class="banner-icon" />
            <div class="banner-text">
              <span class="banner-title">Đang chỉnh sửa tin nhắn</span>
              <span class="banner-sub">Nhấn Enter hoặc bấm nút Lưu để hoàn tất</span>
            </div>
          </div>
          <button class="banner-close" on:click={cancelEdit} title="Hủy chỉnh sửa">
            <X size={16} />
          </button>
        </div>
      {/if}

      {#if isRecording}
        <!-- Thanh Ghi Âm Voice Note -->
        <div class="recording-bar">
          <div class="recording-status">
            <span class="rec-dot"></span>
            <span class="rec-text">Đang ghi âm...</span>
            <span class="rec-timer">{formatTime(recordingTime)}</span>
          </div>
          <div class="rec-actions">
            <button class="rec-btn cancel-btn" on:click={cancelRecording} title="Hủy ghi âm">
              <Trash2 size={18} />
            </button>
            <button class="rec-btn send-rec-btn" on:click={finishRecording} title="Gửi ghi âm">
              <Check size={18} />
            </button>
          </div>
        </div>
      {:else}
        <!-- Khung Soạn Thảo Tin Nhắn Bình Thường -->
        <div class="input-wrapper">
          {#if !editingMessage}
            <button
              class="action-icon-btn"
              on:click={() => fileInput.click()}
              title="Đính kèm ảnh hoặc tài liệu"
              disabled={isUploading}
            >
              <Paperclip size={19} />
            </button>
          {/if}

          <textarea
            rows="1"
            placeholder={editingMessage ? 'Sửa tin nhắn... (Enter để Lưu)' : 'Nhập tin nhắn... hoặc dán ảnh (Ctrl + V)'}
            bind:value={inputContent}
            on:input={handleInput}
            on:keydown={handleKeyDown}
            disabled={isUploading}
          ></textarea>

          {#if !editingMessage}
            <button
              class="action-icon-btn mic-btn"
              on:click={startRecording}
              title="Ghi âm giọng nói (Voice note)"
              disabled={isUploading}
            >
              <Mic size={19} />
            </button>
          {/if}

          <button
            class="send-btn"
            class:edit-save-btn={!!editingMessage}
            on:click={handleSend}
            disabled={!inputContent.trim() || isUploading}
            title={editingMessage ? 'Lưu chỉnh sửa' : 'Gửi'}
          >
            {#if editingMessage}
              <Check size={18} />
            {:else}
              <Send size={18} />
            {/if}
          </button>
        </div>
      {/if}
    </div>
  </main>

  <!-- Image Lightbox Modal Phóng To Toàn Màn Hình -->
  {#if selectedImage}
    <ImageModal
      imageUrl={selectedImage.url}
      fileName={selectedImage.name}
      onClose={() => (selectedImage = null)}
    />
  {/if}

  <!-- Modal Quản Lý Thành Viên Nhóm -->
  {#if showGroupMembersModal && $activeGroup}
    <GroupMembersModal
      groupID={$activeGroup.id}
      onClose={() => (showGroupMembersModal = false)}
    />
  {/if}
{/if}

<style>
  .chat-area { flex: 1; height: 100vh; display: flex; flex-direction: column; background: var(--bg-primary); }
  .empty-chat { flex: 1; display: flex; align-items: center; justify-content: center; text-align: center; }
  .icon-circle { font-size: 50px; margin-bottom: 16px; }
  .empty-content h2 { font-size: 22px; margin-bottom: 8px; }
  .empty-content p { color: var(--text-muted); font-size: 14px; }
  .chat-header {
    padding: 16px 24px;
    border-bottom: 1px solid var(--border-glass);
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .header-actions { display: flex; align-items: center; gap: 8px; }
  .header-action-btn {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid var(--border-glass);
    color: var(--text-dim);
    width: 36px;
    height: 36px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
  }
  .header-action-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.16);
    transform: scale(1.05);
  }
  .group-sender-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 4px;
    padding-bottom: 3px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }
  .bubble-sender-avatar {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    object-fit: cover;
  }
  .bubble-sender-name {
    font-size: 11px;
    font-weight: 600;
    color: #a78bfa;
  }
  .partner-info { display: flex; align-items: center; gap: 14px; }
  .avatar-wrap { position: relative; }
  .header-avatar { width: 44px; height: 44px; border-radius: 50%; }
  .status-dot {
    position: absolute;
    bottom: 2px;
    right: 2px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--offline-color);
    border: 2px solid var(--bg-primary);
  }
  .status-dot.online { background: var(--online-color); }
  .partner-info h3 { font-size: 16px; font-weight: 600; color: #fff; }
  .status-sub { font-size: 12px; }
  .online-text { color: var(--online-color); }
  .offline-text { color: var(--text-dim); }
  .typing-text { color: #a78bfa; font-style: italic; animation: pulse 1.5s infinite; }
  @keyframes pulse { 0%, 100% { opacity: 0.6; } 50% { opacity: 1; } }

  .messages-viewport {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  /* Message Row & Action Toolbar */
  .message-row {
    position: relative;
    display: flex;
    width: 100%;
    transition: background 0.3s;
  }
  .message-row.me { justify-content: flex-end; }

  :global(.highlight-target) {
    animation: flashBackground 1.8s ease;
  }
  @keyframes flashBackground {
    0%, 100% { background: transparent; }
    30% { background: rgba(167, 139, 250, 0.22); border-radius: 8px; }
  }

  .msg-toolbar {
    position: absolute;
    top: -34px;
    display: none;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    border-radius: 20px;
    background: rgba(18, 22, 38, 0.92);
    border: 1px solid var(--border-glass);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.45);
    z-index: 10;
  }
  .message-row:hover .msg-toolbar {
    display: flex;
  }
  .message-row.me .msg-toolbar {
    right: 0;
  }
  .message-row:not(.me) .msg-toolbar {
    left: 0;
  }

  .quick-emojis {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .emoji-btn {
    background: transparent;
    border: none;
    font-size: 15px;
    padding: 2px 4px;
    cursor: pointer;
    border-radius: 4px;
    transition: transform 0.15s;
  }
  .emoji-btn:hover {
    transform: scale(1.3);
  }

  .toolbar-divider {
    width: 1px;
    height: 16px;
    background: var(--border-glass);
    margin: 0 4px;
  }

  .tool-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    width: 26px;
    height: 26px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: 0.15s;
  }
  .tool-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.15);
  }
  .tool-btn.danger:hover {
    color: #f87171;
    background: rgba(239, 68, 68, 0.25);
  }

  /* Bubble Styling */
  .bubble {
    max-width: 65%;
    padding: 10px 14px;
    border-radius: var(--radius-md);
    position: relative;
    word-break: break-word;
  }
  .bubble-me {
    background: var(--bubble-me);
    color: #fff;
    border-bottom-right-radius: 4px;
  }
  .bubble-other {
    background: var(--bubble-other);
    color: #f3f4f6;
    border-bottom-left-radius: 4px;
    border: 1px solid var(--border-glass);
  }

  /* Reply Quote Card */
  .reply-quote {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 6px 10px;
    margin-bottom: 8px;
    border-left: 3px solid #a78bfa;
    background: rgba(0, 0, 0, 0.25);
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;
  }
  .reply-quote:hover {
    background: rgba(0, 0, 0, 0.4);
  }
  .quote-sender {
    font-size: 11px;
    font-weight: 600;
    color: #c4b5fd;
  }
  .quote-text {
    font-size: 12px;
    color: var(--text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 260px;
  }

  /* Deleted Message */
  .deleted-msg {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--text-dim);
    font-size: 13px;
    padding: 2px 0;
  }
  .deleted-icon { font-size: 14px; opacity: 0.8; }

  .text { font-size: 14px; line-height: 1.5; }

  .msg-meta {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    margin-top: 4px;
  }
  .edited-tag {
    font-size: 10px;
    color: var(--text-dim);
    font-style: italic;
  }
  .timestamp { font-size: 10px; opacity: 0.75; }
  .receipt-icon { font-size: 11px; font-weight: 700; opacity: 0.6; }
  .receipt-icon.seen { color: #38bdf8; opacity: 1; }

  /* Reaction Badges Container */
  .reactions-badge-container {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-top: 6px;
  }
  .reaction-pill {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
    border-radius: 12px;
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: #fff;
    font-size: 11px;
    cursor: pointer;
    transition: 0.15s transform, 0.15s background;
  }
  .reaction-pill:hover {
    transform: scale(1.08);
    background: rgba(255, 255, 255, 0.18);
  }
  .reaction-pill.reacted-by-me {
    background: rgba(167, 139, 250, 0.35);
    border-color: #a78bfa;
  }
  .pill-emoji { font-size: 12px; }
  .pill-count { font-weight: 600; font-family: monospace; }

  /* Media Bubbles */
  .image-bubble {
    cursor: pointer;
    overflow: hidden;
    border-radius: var(--radius-sm);
    transition: transform 0.2s;
  }
  .image-bubble:hover {
    transform: scale(1.01);
  }
  .chat-image {
    max-width: 320px;
    max-height: 260px;
    border-radius: var(--radius-sm);
    display: block;
    object-fit: cover;
  }

  .file-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 6px 4px;
    min-width: 200px;
    max-width: 300px;
  }
  .file-icon-box {
    width: 38px;
    height: 38px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #a78bfa;
    flex-shrink: 0;
  }
  .file-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .file-name {
    font-size: 13px;
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: #fff;
  }
  .file-size {
    font-size: 11px;
    color: var(--text-dim);
    margin-top: 2px;
  }
  .file-dl-btn {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.12);
    border: 1px solid var(--border-glass);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    text-decoration: none;
    transition: 0.2s;
    flex-shrink: 0;
  }
  .file-dl-btn:hover {
    background: var(--accent-primary);
    transform: scale(1.08);
  }

  /* Upload Indicator */
  .upload-indicator {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 24px;
    background: rgba(99, 102, 241, 0.15);
    border-top: 1px solid rgba(99, 102, 241, 0.3);
    font-size: 12px;
    color: #c7d2fe;
  }
  :global(.spinner) {
    animation: spin 1s linear infinite;
  }
  @keyframes spin { 100% { transform: rotate(360deg); } }

  /* Typing Indicator Animation */
  .typing-bubble {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 10px 16px;
    border-radius: var(--radius-md);
    border-bottom-left-radius: 4px;
  }
  .typing-dot {
    width: 6px;
    height: 6px;
    background: #a78bfa;
    border-radius: 50%;
    animation: typingBounce 1.4s infinite ease-in-out both;
  }
  .typing-dot:nth-child(1) { animation-delay: -0.32s; }
  .typing-dot:nth-child(2) { animation-delay: -0.16s; }
  @keyframes typingBounce {
    0%, 80%, 100% { transform: scale(0); }
    40% { transform: scale(1); }
  }

  /* Footer & Input */
  .chat-footer {
    padding: 12px 24px 16px 24px;
    border-top: 1px solid var(--border-glass);
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  /* Context Banner (Reply / Edit Preview) */
  .context-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 14px;
    border-radius: var(--radius-md);
    animation: slideDown 0.2s ease;
  }
  .reply-banner {
    background: rgba(167, 139, 250, 0.15);
    border: 1px solid rgba(167, 139, 250, 0.3);
  }
  .edit-banner {
    background: rgba(56, 189, 248, 0.15);
    border: 1px solid rgba(56, 189, 248, 0.3);
  }
  .banner-content {
    display: flex;
    align-items: center;
    gap: 10px;
    overflow: hidden;
  }
  :global(.banner-icon) {
    color: #a78bfa;
    flex-shrink: 0;
  }
  .banner-text {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .banner-title {
    font-size: 12px;
    color: #e0e7ff;
  }
  .banner-sub {
    font-size: 11px;
    color: var(--text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .banner-close {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4px;
    border-radius: 50%;
  }
  .banner-close:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.15);
  }
  @keyframes slideDown {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .input-wrapper {
    display: flex;
    align-items: center;
    gap: 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-lg);
    padding: 6px 14px;
  }
  .input-wrapper textarea {
    flex: 1;
    background: transparent;
    border: none;
    color: #fff;
    font-size: 14px;
    outline: none;
    resize: none;
    font-family: inherit;
    max-height: 100px;
  }

  .action-icon-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    width: 32px;
    height: 32px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
    flex-shrink: 0;
  }
  .action-icon-btn:hover:not(:disabled) {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }
  .mic-btn:hover:not(:disabled) {
    color: #38bdf8;
  }

  .send-btn {
    background: var(--accent-gradient);
    border: none;
    color: #fff;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: 0.2s;
    flex-shrink: 0;
  }
  .send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .send-btn:not(:disabled):hover { transform: scale(1.05); }

  .edit-save-btn {
    background: linear-gradient(135deg, #10b981, #059669);
  }

  /* Recording Bar */
  .recording-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: var(--radius-lg);
    padding: 8px 16px;
    animation: fadeIn 0.2s ease;
  }
  .recording-status {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .rec-dot {
    width: 10px;
    height: 10px;
    background: #ef4444;
    border-radius: 50%;
    animation: blink 1s infinite alternate;
  }
  @keyframes blink {
    0% { opacity: 0.3; transform: scale(0.9); }
    100% { opacity: 1; transform: scale(1.1); }
  }
  .rec-text {
    font-size: 13px;
    color: #fca5a5;
    font-weight: 500;
  }
  .rec-timer {
    font-family: monospace;
    font-size: 14px;
    color: #fff;
    background: rgba(0, 0, 0, 0.3);
    padding: 2px 8px;
    border-radius: 4px;
  }
  .rec-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .rec-btn {
    width: 34px;
    height: 34px;
    border-radius: 50%;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: 0.2s;
  }
  .cancel-btn {
    background: rgba(255, 255, 255, 0.1);
    color: #f87171;
  }
  .cancel-btn:hover {
    background: rgba(239, 68, 68, 0.3);
  }
  .send-rec-btn {
    background: #10b981;
    color: #fff;
  }
  .send-rec-btn:hover {
    background: #059669;
    transform: scale(1.08);
  }
</style>
