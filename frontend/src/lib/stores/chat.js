import { writable, get } from 'svelte/store';
import { apiRequest } from '../services/api';
import { token } from './auth';

export const conversations = writable([]);
export const activeConversation = writable(null); // { conversation, other_user }
export const currentMessages = writable([]);
export const userDirectory = writable([]);

// Tải danh sách hội thoại
export async function loadConversations() {
  const t = get(token);
  if (!t) return;
  try {
    const res = await apiRequest('/chat/conversations', 'GET', null, t);
    conversations.set(res.data || []);
  } catch (err) {
    console.error('Lỗi tải danh sách hội thoại:', err);
  }
}

// Tải lịch sử tin nhắn của cuộc trò chuyện được chọn
export async function selectConversation(convItem) {
  activeConversation.set(convItem);
  const t = get(token);
  const convID = convItem.conversation.custom_id;
  try {
    const res = await apiRequest(`/chat/messages/${convID}?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
    // Đánh dấu đã xem
    apiRequest(`/chat/messages/${convID}/read`, 'POST', null, t);
  } catch (err) {
    console.error('Lỗi tải tin nhắn:', err);
  }
}

// Thêm tin nhắn mới vào danh sách hiện tại nếu đang mở đúng cuộc trò chuyện
export function appendMessage(msg) {
  const active = get(activeConversation);
  if (active && active.conversation?.custom_id === msg.conversation_id) {
    currentMessages.update((msgs) => [...msgs, msg]);
  }
  loadConversations(); // Cập nhật lại danh sách hội thoại và tin nhắn cuối
}
