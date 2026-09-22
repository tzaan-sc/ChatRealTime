import { get } from 'svelte/store';
import { notificationSettings } from '../stores/notification';

// Singleton AudioContext tái sử dụng, triệt tiêu memory leak
let sharedAudioCtx = null;

export function getAudioContext() {
  if (typeof window === 'undefined') return null;
  const AudioCtx = window.AudioContext || window.webkitAudioContext;
  if (!AudioCtx) return null;

  if (!sharedAudioCtx || sharedAudioCtx.state === 'closed') {
    sharedAudioCtx = new AudioCtx();
  }

  if (sharedAudioCtx.state === 'suspended') {
    sharedAudioCtx.resume().catch(() => {});
  }

  return sharedAudioCtx;
}

// Tự động unlock AudioContext khi người dùng click lần đầu vào trang
if (typeof window !== 'undefined') {
  const unlockAudio = () => {
    const ctx = getAudioContext();
    if (ctx && ctx.state === 'suspended') {
      ctx.resume().catch(() => {});
    }
    window.removeEventListener('click', unlockAudio);
    window.removeEventListener('keydown', unlockAudio);
  };
  window.addEventListener('click', unlockAudio, { passive: true });
  window.addEventListener('keydown', unlockAudio, { passive: true });
}

// Phát âm thanh thông báo tin nhắn
// subtle = true: âm thanh rất nhẹ, êm dịu khi người dùng đang mở sẵn cuộc trò chuyện
// subtle = false: âm thanh Pop trong trẻo khi có tin nhắn từ cuộc trò chuyện khác / tab ẩn
export function playNotificationSound({ subtle = false, volume = null } = {}) {
  try {
    const settings = get(notificationSettings);
    if (!settings.soundEnabled) return;

    const ctx = getAudioContext();
    if (!ctx) return;

    const now = ctx.currentTime;
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    const masterVolume = volume !== null ? volume : (settings.soundVolume ?? 0.8);

    if (subtle) {
      // Âm click êm dịu: F5 (698Hz) -> C6 (1046Hz), thời lượng 60ms
      osc.type = 'sine';
      osc.frequency.setValueAtTime(698.46, now);
      osc.frequency.exponentialRampToValueAtTime(1046.5, now + 0.05);

      const targetGain = 0.04 * masterVolume;
      gain.gain.setValueAtTime(targetGain, now);
      gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.06);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now);
      osc.stop(now + 0.06);
    } else {
      // Âm Pop vui tai: D5 (587Hz) -> A5 (880Hz), thời lượng 120ms
      osc.type = 'sine';
      osc.frequency.setValueAtTime(587.33, now);
      osc.frequency.exponentialRampToValueAtTime(880, now + 0.08);

      const targetGain = 0.14 * masterVolume;
      gain.gain.setValueAtTime(targetGain, now);
      gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.12);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now);
      osc.stop(now + 0.12);
    }
  } catch (e) {
    console.warn('Không thể phát âm thanh thông báo:', e);
  }
}

// Biến quản trị trạng thái chuông cuộc gọi đến
let ringtoneTimer = null;

// Phát chuông reo cho cuộc gọi đến (Incoming Call Ringtone)
export function startRingtone() {
  stopRingtone();
  try {
    const settings = get(notificationSettings);
    if (!settings.soundEnabled) return;

    const ctx = getAudioContext();
    if (!ctx) return;

    const playChime = () => {
      if (!ctx || ctx.state === 'closed') return;
      const now = ctx.currentTime;
      const masterVolume = settings.soundVolume ?? 0.8;

      // Hợp âm 4 nốt chuông hiện đại: E5 (659Hz) -> G#5 (830Hz) -> B5 (987Hz) -> E6 (1318Hz)
      const notes = [659.25, 830.61, 987.77, 1318.51];
      notes.forEach((freq, idx) => {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, now + idx * 0.12);

        gain.gain.setValueAtTime(0.12 * masterVolume, now + idx * 0.12);
        gain.gain.exponentialRampToValueAtTime(0.001, now + idx * 0.12 + 0.35);

        osc.connect(gain);
        gain.connect(ctx.destination);

        osc.start(now + idx * 0.12);
        osc.stop(now + idx * 0.12 + 0.35);
      });
    };

    playChime();
    ringtoneTimer = setInterval(playChime, 2200);
  } catch (e) {
    console.warn('Không thể khởi tạo chuông cuộc gọi:', e);
  }
}

export function stopRingtone() {
  if (ringtoneTimer) {
    clearInterval(ringtoneTimer);
    ringtoneTimer = null;
  }
}

// Biến quản trị tiếng tút chờ máy khi gọi đi
let callingTimer = null;

// Phát tiếng "tút... tút..." khi đang gọi đi
export function startCallingTone() {
  stopCallingTone();
  try {
    const ctx = getAudioContext();
    if (!ctx) return;

    const playBeep = () => {
      if (!ctx || ctx.state === 'closed') return;
      const now = ctx.currentTime;

      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = 'sine';
      osc.frequency.setValueAtTime(440, now);

      gain.gain.setValueAtTime(0.06, now);
      gain.gain.setValueAtTime(0.06, now + 0.8);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.85);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now);
      osc.stop(now + 0.85);
    };

    playBeep();
    callingTimer = setInterval(playBeep, 2600);
  } catch (e) {
    console.warn('Không thể khởi tạo tiếng chờ máy:', e);
  }
}

export function stopCallingTone() {
  if (callingTimer) {
    clearInterval(callingTimer);
    callingTimer = null;
  }
}
