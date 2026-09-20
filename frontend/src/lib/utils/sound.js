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
    // Trình duyệt có thể chưa cho phép autoplay nếu user chưa có tương tác đầu tiên
    console.warn('Âm thanh thông báo chưa sẵn sàng:', e);
  }
}
