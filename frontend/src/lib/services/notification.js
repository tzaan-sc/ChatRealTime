import { get } from 'svelte/store';
import { notificationSettings } from '../stores/notification';

let originalTitle = typeof document !== 'undefined' ? document.title || 'Chat RealTime' : 'Chat RealTime';
let titleTimer = null;
let isFlashing = false;

// Yêu cầu cấp quyền Desktop Notification từ người dùng
export async function requestNotificationPermission() {
  if (typeof window === 'undefined' || !('Notification' in window)) {
    return 'unsupported';
  }
  try {
    const permission = await Notification.requestPermission();
    return permission;
  } catch (err) {
    console.warn('Lỗi xin quyền thông báo:', err);
    return 'default';
  }
}

// Kiểm tra quyền thông báo hiện tại
export function getNotificationPermission() {
  if (typeof window === 'undefined' || !('Notification' in window)) {
    return 'unsupported';
  }
  return Notification.permission;
}

// Hiển thị Desktop Notification của trình duyệt
export function showDesktopNotification({
  title,
  body,
  icon = '/favicon.png',
  tag = 'chat_message',
  onClick = null
}) {
  if (typeof window === 'undefined' || !('Notification' in window)) return;
  if (Notification.permission !== 'granted') return;

  const settings = get(notificationSettings);
  if (!settings.desktopEnabled) return;

  try {
    const notification = new Notification(title, {
      body: body || 'Bạn có một tin nhắn mới',
      icon: icon || undefined,
      tag,
      renotify: true
    });

    notification.onclick = () => {
      window.focus();
      notification.close();
      if (typeof onClick === 'function') {
        onClick();
      }
    };

    // Tự động đóng sau 6 giây
    setTimeout(() => {
      try {
        notification.close();
      } catch (_) {}
    }, 6000);
  } catch (err) {
    console.warn('Không thể hiển thị thông báo trình duyệt:', err);
  }
}

// Bắt đầu nhấp nháy tiêu đề Tab khi có tin nhắn mới lúc tab đang ẩn
export function startFlashingTitle(senderName, messageText) {
  if (typeof document === 'undefined') return;
  if (document.visibilityState === 'visible') return;

  stopFlashingTitle();
  isFlashing = true;

  const snippet = messageText ? (messageText.length > 25 ? messageText.substring(0, 25) + '...' : messageText) : 'Tin nhắn mới';
  const alertTitle = `(1) 💬 ${senderName || 'Tin mới'}: ${snippet}`;
  let toggle = false;

  titleTimer = setInterval(() => {
    if (document.visibilityState === 'visible') {
      stopFlashingTitle();
      return;
    }
    document.title = toggle ? alertTitle : `(1) Tin nhắn mới - Chat RealTime`;
    toggle = !toggle;
  }, 1200);

  document.title = alertTitle;
}

// Dừng nhấp nháy và trả lại tiêu đề ban đầu
export function stopFlashingTitle() {
  if (titleTimer) {
    clearInterval(titleTimer);
    titleTimer = null;
  }
  if (isFlashing && typeof document !== 'undefined') {
    document.title = 'Chat RealTime';
    isFlashing = false;
  }
}

// Lắng nghe sự kiện quay lại tab để dừng nhấp nháy tiêu đề
if (typeof window !== 'undefined') {
  window.addEventListener('focus', stopFlashingTitle);
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
      stopFlashingTitle();
    }
  });
}
