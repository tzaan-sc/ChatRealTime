<script>
  import {
    callState,
    callType,
    callPartner,
    localStream,
    remoteStream,
    isMuted,
    isCameraOff,
    isScreenSharing,
    callDuration,
    acceptIncomingCall,
    rejectIncomingCall,
    endCall,
    toggleMute,
    toggleCamera,
    toggleScreenShare
  } from '../stores/call';
  import {
    Phone,
    PhoneOff,
    Mic,
    MicOff,
    Video,
    VideoOff,
    Monitor,
    Maximize2
  } from 'lucide-svelte';

  let localVideoEl;
  let remoteVideoEl;
  let remoteAudioEl;

  // Gắn luồng local stream vào thẻ video của bản thân
  $: if (localVideoEl && $localStream) {
    localVideoEl.srcObject = $localStream;
  }

  // Gắn luồng remote stream vào thẻ video và audio của đối phương
  $: if (remoteVideoEl && $remoteStream) {
    remoteVideoEl.srcObject = $remoteStream;
  }
  $: if (remoteAudioEl && $remoteStream) {
    remoteAudioEl.srcObject = $remoteStream;
  }

  function formatTime(totalSeconds) {
    const mins = Math.floor(totalSeconds / 60);
    const secs = totalSeconds % 60;
    return `${mins < 10 ? '0' : ''}${mins}:${secs < 10 ? '0' : ''}${secs}`;
  }
</script>

{#if $callState !== 'idle'}
  <div class="call-overlay">
    <!-- Thẻ audio ẩn để phát âm thanh đối phương luôn luôn hoạt động -->
    <audio bind:this={remoteAudioEl} autoplay playsinline style="display: none;"></audio>

    <div class="call-container glass-card">
      <!-- 1. Trạng thái Đang có cuộc gọi đến (Incoming Call) -->
      {#if $callState === 'incoming'}
        <div class="ringing-card">
          <div class="pulsing-avatar-wrap">
            <div class="pulse-ring ring-1"></div>
            <div class="pulse-ring ring-2"></div>
            <img
              src={$callPartner?.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${$callPartner?.name}`}
              alt="avatar"
              class="caller-avatar"
            />
          </div>

          <h3 class="caller-name">{$callPartner?.name || 'Ai đó'}</h3>
          <p class="call-subtext">
            Đang gọi {$callType === 'video' ? 'Video' : 'Thoại'} cho bạn...
          </p>

          <div class="ring-actions">
            <button class="ring-btn decline-btn" on:click={() => rejectIncomingCall('declined')} title="Từ chối">
              <PhoneOff size={26} />
              <span>Từ chối</span>
            </button>
            <button class="ring-btn accept-btn" on:click={acceptIncomingCall} title="Trả lời">
              <Phone size={26} />
              <span>Trả lời</span>
            </button>
          </div>
        </div>

      <!-- 2. Trạng thái Đang gọi đi (Outgoing Call) -->
      {:else if $callState === 'outgoing'}
        <div class="ringing-card">
          <div class="pulsing-avatar-wrap">
            <div class="pulse-ring ring-1"></div>
            <div class="pulse-ring ring-2"></div>
            <img
              src={$callPartner?.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${$callPartner?.name}`}
              alt="avatar"
              class="caller-avatar"
            />
          </div>

          <h3 class="caller-name">{$callPartner?.name || 'Đối phương'}</h3>
          <p class="call-subtext">
            Đang kết nối cuộc gọi {$callType === 'video' ? 'Video' : 'Thoại'}...
          </p>

          <div class="ring-actions single-action">
            <button class="ring-btn decline-btn" on:click={endCall} title="Hủy cuộc gọi">
              <PhoneOff size={26} />
              <span>Hủy</span>
            </button>
          </div>
        </div>

      <!-- 3. Trạng thái Đang trong cuộc đàm thoại (Connected) -->
      {:else if $callState === 'connected'}
        <div class="active-call-layout">
          <!-- Thanh Header cuộc gọi -->
          <div class="call-topbar">
            <div class="top-partner">
              <span class="partner-title">{$callPartner?.name}</span>
              <span class="call-timer">{formatTime($callDuration)}</span>
            </div>
          </div>

          <!-- Khu vực hiển thị Video chính (Remote) -->
          <div class="main-video-area">
            {#if $callType === 'video'}
              <!-- svelte-ignore a11y_media_has_caption -->
              <video
                bind:this={remoteVideoEl}
                autoplay
                playsinline
                class="remote-video"
              ></video>
            {:else}
              <div class="audio-call-visual">
                <div class="pulse-ring ring-1"></div>
                <img
                  src={$callPartner?.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${$callPartner?.name}`}
                  alt="avatar"
                  class="audio-avatar"
                />
                <span class="audio-label">Đang đàm thoại thoại</span>
              </div>
            {/if}

            <!-- Khung video nhỏ của chính mình (Picture-in-Picture) -->
            {#if $callType === 'video'}
              <div class="pip-container" class:cam-off={$isCameraOff}>
                <!-- svelte-ignore a11y_media_has_caption -->
                <video
                  bind:this={localVideoEl}
                  autoplay
                  playsinline
                  muted
                  class="local-video"
                ></video>
                {#if $isCameraOff}
                  <div class="cam-off-overlay">
                    <VideoOff size={24} />
                    <span>Camera tắt</span>
                  </div>
                {/if}
              </div>
            {/if}
          </div>

          <!-- Bảng nút điều khiển đàm thoại phía dưới -->
          <div class="call-controls-bar glass-card">
            <!-- Nút Bật/Tắt Micro -->
            <button
              class="ctrl-btn"
              class:active-off={$isMuted}
              on:click={toggleMute}
              title={$isMuted ? 'Bật Micro' : 'Tắt Micro'}
            >
              {#if $isMuted}
                <MicOff size={20} />
              {:else}
                <Mic size={20} />
              {/if}
            </button>

            <!-- Nút Bật/Tắt Camera (chỉ hiện khi gọi video) -->
            {#if $callType === 'video'}
              <button
                class="ctrl-btn"
                class:active-off={$isCameraOff}
                on:click={toggleCamera}
                title={$isCameraOff ? 'Bật Camera' : 'Tắt Camera'}
              >
                {#if $isCameraOff}
                  <VideoOff size={20} />
                {:else}
                  <Video size={20} />
                {/if}
              </button>

              <!-- Nút Chia sẻ màn hình -->
              <button
                class="ctrl-btn"
                class:active-screen={$isScreenSharing}
                on:click={toggleScreenShare}
                title={$isScreenSharing ? 'Dừng chia sẻ màn hình' : 'Chia sẻ màn hình'}
              >
                <Monitor size={20} />
              </button>
            {/if}

            <!-- Nút Kết thúc cuộc gọi -->
            <button class="ctrl-btn hangup-btn" on:click={endCall} title="Kết thúc cuộc gọi">
              <PhoneOff size={22} />
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .call-overlay {
    position: fixed;
    inset: 0;
    z-index: 9999;
    background: rgba(8, 10, 20, 0.85);
    backdrop-filter: blur(16px);
    display: flex;
    align-items: center;
    justify-content: center;
    animation: fadeIn 0.25s ease;
  }

  .call-container {
    width: 90vw;
    max-width: 960px;
    height: 85vh;
    max-height: 700px;
    background: rgba(18, 22, 40, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 20px;
    overflow: hidden;
    position: relative;
    display: flex;
    box-shadow: 0 25px 60px rgba(0, 0, 0, 0.65);
  }

  /* Màn hình Đổ chuông (Incoming / Outgoing) */
  .ringing-card {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    text-align: center;
  }

  .pulsing-avatar-wrap {
    position: relative;
    width: 130px;
    height: 130px;
    margin-bottom: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .caller-avatar {
    width: 110px;
    height: 110px;
    border-radius: 50%;
    object-fit: cover;
    border: 3px solid #a78bfa;
    position: relative;
    z-index: 2;
    box-shadow: 0 8px 30px rgba(167, 139, 250, 0.4);
  }

  .pulse-ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 2px solid #a78bfa;
    animation: pulseRing 2s cubic-bezier(0.215, 0.61, 0.355, 1) infinite;
  }
  .ring-2 {
    animation-delay: 0.6s;
  }

  @keyframes pulseRing {
    0% { transform: scale(0.85); opacity: 0.8; }
    80% { transform: scale(1.6); opacity: 0; }
    100% { transform: scale(1.7); opacity: 0; }
  }

  .caller-name {
    font-size: 24px;
    font-weight: 700;
    color: #fff;
    margin-bottom: 8px;
  }

  .call-subtext {
    font-size: 15px;
    color: #a78bfa;
    margin-bottom: 40px;
  }

  .ring-actions {
    display: flex;
    align-items: center;
    gap: 36px;
  }
  .single-action {
    justify-content: center;
  }

  .ring-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 13px;
    color: #e2e8f0;
    transition: transform 0.2s;
  }
  .ring-btn:hover {
    transform: scale(1.08);
  }

  .accept-btn :global(svg) {
    width: 58px;
    height: 58px;
    padding: 14px;
    background: linear-gradient(135deg, #10b981, #059669);
    border-radius: 50%;
    color: #fff;
    box-shadow: 0 6px 20px rgba(16, 185, 129, 0.45);
  }

  .decline-btn :global(svg) {
    width: 58px;
    height: 58px;
    padding: 14px;
    background: linear-gradient(135deg, #ef4444, #dc2626);
    border-radius: 50%;
    color: #fff;
    box-shadow: 0 6px 20px rgba(239, 68, 68, 0.45);
  }

  /* Màn hình Active Call */
  .active-call-layout {
    flex: 1;
    display: flex;
    flex-direction: column;
    position: relative;
    background: #0b0f19;
  }

  .call-topbar {
    position: absolute;
    top: 16px;
    left: 20px;
    right: 20px;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: space-between;
    pointer-events: none;
  }

  .top-partner {
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(8px);
    padding: 6px 14px;
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .partner-title {
    font-size: 14px;
    font-weight: 600;
    color: #fff;
  }
  .call-timer {
    font-family: monospace;
    font-size: 13px;
    color: #38bdf8;
    background: rgba(56, 189, 248, 0.15);
    padding: 2px 8px;
    border-radius: 6px;
  }

  .main-video-area {
    flex: 1;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    background: #000;
  }

  .remote-video {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .audio-call-visual {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
  }
  .audio-avatar {
    width: 120px;
    height: 120px;
    border-radius: 50%;
    object-fit: cover;
    border: 3px solid #38bdf8;
    box-shadow: 0 10px 40px rgba(56, 189, 248, 0.35);
  }
  .audio-label {
    color: #94a3b8;
    font-size: 14px;
    letter-spacing: 0.5px;
  }

  /* Picture in picture (Local Camera) */
  .pip-container {
    position: absolute;
    bottom: 84px;
    right: 20px;
    width: 180px;
    height: 120px;
    border-radius: 12px;
    overflow: hidden;
    border: 2px solid rgba(255, 255, 255, 0.25);
    background: #1e293b;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
    z-index: 15;
  }
  .local-video {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transform: scaleX(-1); /* Gương camera trước */
  }

  .cam-off-overlay {
    position: absolute;
    inset: 0;
    background: #1e2238;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    color: #94a3b8;
    font-size: 11px;
  }

  /* Toolbar điều khiển */
  .call-controls-bar {
    position: absolute;
    bottom: 18px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 20;
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 8px 18px;
    border-radius: 40px;
    background: rgba(15, 23, 42, 0.85);
    border: 1px solid rgba(255, 255, 255, 0.12);
  }

  .ctrl-btn {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    border: none;
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: 0.2s transform, 0.2s background;
  }
  .ctrl-btn:hover {
    transform: scale(1.08);
    background: rgba(255, 255, 255, 0.2);
  }
  .ctrl-btn.active-off {
    background: rgba(239, 68, 68, 0.25);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.4);
  }
  .ctrl-btn.active-screen {
    background: rgba(56, 189, 248, 0.3);
    color: #38bdf8;
    border: 1px solid #38bdf8;
  }
  .hangup-btn {
    background: #ef4444;
    color: #fff;
  }
  .hangup-btn:hover {
    background: #dc2626;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
</style>
