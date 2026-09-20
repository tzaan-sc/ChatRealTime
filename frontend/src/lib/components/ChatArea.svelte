<script>
  import { activeConversation, currentMessages, onlineUsers, typingUsers } from '../stores/chat';
  import { currentUser } from '../stores/auth';
  import { wsService } from '../services/websocket';
  import { afterUpdate } from 'svelte';
  import { Send, Paperclip, Mic, Trash2, Check, FileText, Download, Loader2 } from 'lucide-svelte';
  import ImageModal from './ImageModal.svelte';
  import AudioPlayer from './AudioPlayer.svelte';

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

  // Kiểm tra đối phương có online không
  $: partnerID = $activeConversation?.other_user?.id;
  $: isPartnerOnline = partnerID ? $onlineUsers.has(partnerID) : false;
  $: isPartnerTyping = partnerID ? $typingUsers.has(partnerID) : false;

  // Tự động cuộn xuống đáy khi có tin nhắn mới
  afterUpdate(() => {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  });

  // Bắt sự kiện gõ phím để bắn Typing Indicator
  function handleInput() {
    if (!$activeConversation || !partnerID) return;

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

  // Gửi tin nhắn chữ
  function handleSend() {
    if (!inputContent.trim() || !$activeConversation || !partnerID) return;

    clearTimeout(typingTimeout);
    if (isTyping) {
      isTyping = false;
      wsService.send('typing:stop', { receiver_id: partnerID });
    }

    wsService.send('chat:send', {
      receiver_id: partnerID,
      content: inputContent.trim(),
      type: 'text'
    });

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
    if (!file || !$activeConversation || !partnerID) return;

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

      wsService.send('chat:send', {
        receiver_id: partnerID,
        content: data.file_url,
        type: data.type,
        file_name: data.file_name,
        file_size: data.file_size
      });
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

  // Gửi event xác nhận đã đọc khi đang mở xem khung chat
  $: if ($activeConversation && partnerID) {
    wsService.send('chat:read', {
      conversation_id: $activeConversation.conversation.custom_id,
      partner_id: partnerID
    });
  }
</script>

{#if !$activeConversation}
  <div class="empty-chat glass-card">
    <div class="empty-content">
      <div class="icon-circle">💬</div>
      <h2>Chào mừng đến với Realtime Chat!</h2>
      <p>Chọn một người bạn ở danh sách bên trái để bắt đầu cuộc trò chuyện.</p>
    </div>
  </div>
{:else}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <main class="chat-area" on:paste={handlePaste}>
    <!-- Header -->
    <div class="chat-header glass-card">
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
    </div>

    <!-- Messages Body -->
    <div class="messages-viewport" bind:this={messagesContainer}>
      {#each $currentMessages as msg}
        {@const isMe = msg.sender_id === $currentUser.id}
        <div class="message-row" class:me={isMe}>
          <div class="bubble" class:bubble-me={isMe} class:bubble-other={!isMe}>
            <!-- Render tuỳ loại tin nhắn -->
            {#if msg.type === 'image'}
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

            <div class="msg-meta">
              <span class="timestamp">
                {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
              {#if isMe}
                <span class="receipt-icon" class:seen={msg.is_read} title={msg.is_read ? 'Đã xem' : 'Đã gửi'}>
                  {msg.is_read ? '✓✓' : '✓'}
                </span>
              {/if}
            </div>
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

    <!-- Input Footer -->
    <div class="chat-footer glass-card">
      <input
        type="file"
        bind:this={fileInput}
        on:change={handleFileInputChange}
        style="display: none;"
      />

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
          <button
            class="action-icon-btn"
            on:click={() => fileInput.click()}
            title="Đính kèm ảnh hoặc tài liệu"
            disabled={isUploading}
          >
            <Paperclip size={19} />
          </button>

          <textarea
            rows="1"
            placeholder="Nhập tin nhắn... hoặc dán ảnh (Ctrl + V)"
            bind:value={inputContent}
            on:input={handleInput}
            on:keydown={handleKeyDown}
            disabled={isUploading}
          ></textarea>

          <button
            class="action-icon-btn mic-btn"
            on:click={startRecording}
            title="Ghi âm giọng nói (Voice note)"
            disabled={isUploading}
          >
            <Mic size={19} />
          </button>

          <button class="send-btn" on:click={handleSend} disabled={!inputContent.trim() || isUploading}>
            <Send size={18} />
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
    gap: 12px;
  }
  .message-row { display: flex; width: 100%; }
  .message-row.me { justify-content: flex-end; }
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
  .text { font-size: 14px; line-height: 1.5; }
  .msg-meta {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    margin-top: 4px;
  }
  .timestamp { font-size: 10px; opacity: 0.75; }
  .receipt-icon { font-size: 11px; font-weight: 700; opacity: 0.6; }
  .receipt-icon.seen { color: #38bdf8; opacity: 1; }

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

  .chat-footer { padding: 14px 24px; border-top: 1px solid var(--border-glass); }
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
