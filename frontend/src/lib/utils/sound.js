// Phát âm thanh thông báo "Pop" nhẹ nhàng khi có tin nhắn đến
export function playNotificationSound() {
  try {
    const AudioCtx = window.AudioContext || window.webkitAudioContext;
    if (!AudioCtx) return;

    const ctx = new AudioCtx();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = 'sine';
    // Lướt từ nốt D5 (587.33Hz) lên A5 (880Hz) tạo cảm giác tin nhắn đến êm dịu
    osc.frequency.setValueAtTime(587.33, ctx.currentTime);
    osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.08);

    // Âm lượng fade out mềm mại trong 120ms
    gain.gain.setValueAtTime(0.12, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.12);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start();
    osc.stop(ctx.currentTime + 0.12);
  } catch (e) {
    console.warn('Âm thanh thông báo chưa sẵn sàng:', e);
  }
}

// Biến quản trị trạng thái chuông Web Audio
let ringtoneTimer = null;
let ringtoneAudioCtx = null;

// Phát chuông reo cho cuộc gọi đến (Incoming Call Ringtone)
export function startRingtone() {
  stopRingtone();
  try {
    const AudioCtx = window.AudioContext || window.webkitAudioContext;
    if (!AudioCtx) return;
    ringtoneAudioCtx = new AudioCtx();

    const playChime = () => {
      if (!ringtoneAudioCtx || ringtoneAudioCtx.state === 'closed') return;
      const now = ringtoneAudioCtx.currentTime;

      // Hợp âm 3 nốt chuông hiện đại: E5 (659Hz) -> G#5 (830Hz) -> B5 (987Hz)
      const notes = [659.25, 830.61, 987.77, 1318.51];
      notes.forEach((freq, idx) => {
        const osc = ringtoneAudioCtx.createOscillator();
        const gain = ringtoneAudioCtx.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, now + idx * 0.12);

        gain.gain.setValueAtTime(0.1, now + idx * 0.12);
        gain.gain.exponentialRampToValueAtTime(0.001, now + idx * 0.12 + 0.35);

        osc.connect(gain);
        gain.connect(ringtoneAudioCtx.destination);

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
  if (ringtoneAudioCtx) {
    try {
      ringtoneAudioCtx.close();
    } catch (_) {}
    ringtoneAudioCtx = null;
  }
}

// Biến quản trị tiếng tút chờ máy (Outgoing Calling Tone)
let callingTimer = null;
let callingAudioCtx = null;

// Phát tiếng "tút... tút..." khi đang gọi đi
export function startCallingTone() {
  stopCallingTone();
  try {
    const AudioCtx = window.AudioContext || window.webkitAudioContext;
    if (!AudioCtx) return;
    callingAudioCtx = new AudioCtx();

    const playBeep = () => {
      if (!callingAudioCtx || callingAudioCtx.state === 'closed') return;
      const now = callingAudioCtx.currentTime;

      const osc = callingAudioCtx.createOscillator();
      const gain = callingAudioCtx.createGain();

      osc.type = 'sine';
      osc.frequency.setValueAtTime(440, now); // Chuẩn tần số 440Hz

      gain.gain.setValueAtTime(0.06, now);
      gain.gain.setValueAtTime(0.06, now + 0.8);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.85);

      osc.connect(gain);
      gain.connect(callingAudioCtx.destination);

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
  if (callingAudioCtx) {
    try {
      callingAudioCtx.close();
    } catch (_) {}
    callingAudioCtx = null;
  }
}

