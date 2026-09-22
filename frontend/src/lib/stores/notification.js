import { writable, get } from 'svelte/store';

const DEFAULT_SETTINGS = {
  soundEnabled: true,
  desktopEnabled: true,
  inAppToastEnabled: true,
  soundVolume: 0.8
};

function loadSettings() {
  if (typeof window === 'undefined') return DEFAULT_SETTINGS;
  try {
    const raw = localStorage.getItem('chat_notification_settings');
    if (raw) {
      return { ...DEFAULT_SETTINGS, ...JSON.parse(raw) };
    }
  } catch (e) {
    console.warn('Lỗi đọc cài đặt thông báo:', e);
  }
  return DEFAULT_SETTINGS;
}

export const notificationSettings = writable(loadSettings());

export function updateNotificationSettings(newSettings) {
  notificationSettings.update((curr) => {
    const updated = { ...curr, ...newSettings };
    try {
      localStorage.setItem('chat_notification_settings', JSON.stringify(updated));
    } catch (_) {}
    return updated;
  });
}

// Store danh sách Banner thông báo nổi trong ứng dụng
export const activeToasts = writable([]);

export function addToast({
  title = 'Tin nhắn mới',
  senderName = '',
  senderAvatar = '',
  content = '',
  conversationId = null,
  groupId = null,
  duration = 4500
}) {
  const settings = get(notificationSettings);
  if (!settings.inAppToastEnabled) return;

  const id = 'toast_' + Date.now() + '_' + Math.random().toString(36).substr(2, 5);
  const toastItem = {
    id,
    title,
    senderName,
    senderAvatar,
    content,
    conversationId,
    groupId,
    createdAt: Date.now()
  };

  activeToasts.update((items) => [toastItem, ...items.slice(0, 3)]); // Giữ tối đa 4 thông báo nổi cùng lúc

  if (duration > 0) {
    setTimeout(() => {
      removeToast(id);
    }, duration);
  }
}

export function removeToast(id) {
  activeToasts.update((items) => items.filter((t) => t.id !== id));
}
