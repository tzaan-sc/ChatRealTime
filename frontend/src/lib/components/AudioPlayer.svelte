<script>
  import { onMount, onDestroy } from 'svelte';
  import { Play, Pause, Volume2 } from 'lucide-svelte';

  export let src = '';

  let audio;
  let isPlaying = false;
  let currentTime = 0;
  let duration = 0;

  function togglePlay() {
    if (!audio) return;
    if (isPlaying) {
      audio.pause();
    } else {
      audio.play();
    }
  }

  function handleTimeUpdate() {
    if (audio) {
      currentTime = audio.currentTime;
    }
  }

  function handleLoadedMetadata() {
    if (audio && audio.duration && !isNaN(audio.duration) && isFinite(audio.duration)) {
      duration = audio.duration;
    }
  }

  function handleEnded() {
    isPlaying = false;
    currentTime = 0;
  }

  function handleSeek(e) {
    const seekTime = parseFloat(e.target.value);
    if (audio) {
      audio.currentTime = seekTime;
      currentTime = seekTime;
    }
  }

  function formatTime(seconds) {
    if (isNaN(seconds) || !isFinite(seconds)) return '0:00';
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs < 10 ? '0' : ''}${secs}`;
  }

  onDestroy(() => {
    if (audio) {
      audio.pause();
    }
  });
</script>

<div class="voice-player">
  <audio
    bind:this={audio}
    {src}
    on:timeupdate={handleTimeUpdate}
    on:loadedmetadata={handleLoadedMetadata}
    on:ended={handleEnded}
    on:play={() => (isPlaying = true)}
    on:pause={() => (isPlaying = false)}
    preload="metadata"
  ></audio>

  <button class="play-btn" on:click={togglePlay} title={isPlaying ? 'Tạm dừng' : 'Phát'}>
    {#if isPlaying}
      <Pause size={15} />
    {:else}
      <Play size={15} style="margin-left: 2px;" />
    {/if}
  </button>

  <div class="player-body">
    <!-- Thanh tiến trình / waveform giả lập -->
    <div class="track-container">
      <input
        type="range"
        min="0"
        max={duration || 1}
        step="0.1"
        value={currentTime}
        on:input={handleSeek}
        class="progress-slider"
      />
    </div>

    <div class="time-meta">
      <span class="time-text">{formatTime(currentTime)}</span>
      <span class="time-text duration">{formatTime(duration)}</span>
    </div>
  </div>
</div>

<style>
  .voice-player {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 10px;
    border-radius: var(--radius-md);
    background: rgba(0, 0, 0, 0.22);
    min-width: 220px;
    max-width: 280px;
  }

  .play-btn {
    width: 34px;
    height: 34px;
    border-radius: 50%;
    border: none;
    background: var(--accent-gradient);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    flex-shrink: 0;
    box-shadow: 0 4px 10px rgba(99, 102, 241, 0.4);
    transition: 0.2s transform, 0.2s filter;
  }

  .play-btn:hover {
    transform: scale(1.08);
    filter: brightness(1.15);
  }

  .player-body {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .track-container {
    display: flex;
    align-items: center;
    width: 100%;
  }

  .progress-slider {
    width: 100%;
    height: 4px;
    -webkit-appearance: none;
    appearance: none;
    background: rgba(255, 255, 255, 0.15);
    border-radius: 4px;
    outline: none;
    cursor: pointer;
    transition: background 0.2s;
  }

  .progress-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #a78bfa;
    cursor: pointer;
    box-shadow: 0 0 6px rgba(167, 139, 250, 0.8);
    transition: transform 0.15s;
  }

  .progress-slider:hover::-webkit-slider-thumb {
    transform: scale(1.3);
    background: #fff;
  }

  .time-meta {
    display: flex;
    justify-content: space-between;
    font-size: 10px;
    color: var(--text-dim);
    font-family: monospace;
  }

  .duration {
    opacity: 0.8;
  }
</style>
