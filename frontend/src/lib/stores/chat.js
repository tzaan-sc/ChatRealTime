import { writable, get } from 'svelte/store';
import { apiRequest } from '../services/api';
import { token } from './auth';

export const conversations = writable([]);
export const activeConversation = writable(null); // { conversation, other_user }
export const userGroups = writable([]);
export const activeGroup = writable(null); // Group object
export const currentMessages = writable([]);
export const userDirectory = writable([]);

// Set các User ID đang online (Ví dụ: Set(['id1', 'id2']))
export const onlineUsers = writable(new Set());

// Set các User ID đang gõ phím vào khung chat
export const typingUsers = writable(new Set());

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

// Tải danh sách nhóm của user
export async function loadUserGroups() {
  const t = get(token);
  if (!t) return;
  try {
    const res = await apiRequest('/groups', 'GET', null, t);
    userGroups.set(res.data || []);
  } catch (err) {
    console.error('Lỗi tải danh sách nhóm:', err);
  }
}

// Chọn cuộc trò chuyện 1-1 và tải lịch sử
export async function selectConversation(convItem) {
  activeGroup.set(null);
  activeConversation.set(convItem);
  const t = get(token);
  const convID = convItem.conversation.custom_id;
  try {
    const res = await apiRequest(`/chat/messages/${convID}?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
    // Đánh dấu đã xem trên server
    apiRequest(`/chat/messages/${convID}/read`, 'POST', null, t);
  } catch (err) {
    console.error('Lỗi tải tin nhắn:', err);
  }
}

// Chọn nhóm chat và tải lịch sử tin nhắn nhóm
export async function selectGroup(groupItem) {
  activeConversation.set(null);
  activeGroup.set(groupItem);
  const t = get(token);
  try {
    const res = await apiRequest(`/groups/${groupItem.id}/messages?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
  } catch (err) {
    console.error('Lỗi tải tin nhắn nhóm:', err);
  }
}

// Thêm tin nhắn mới vào danh sách hiện tại nếu đang mở đúng cuộc trò chuyện 1-1
export function appendMessage(msg) {
  const active = get(activeConversation);
  if (active && active.conversation?.custom_id === msg.conversation_id) {
    currentMessages.update((msgs) => [...msgs, msg]);
  }
  loadConversations();
}

// Thêm tin nhắn nhóm mới vào danh sách nếu đang mở đúng nhóm
export function appendGroupMessage(msg) {
  const grp = get(activeGroup);
  if (grp && grp.id === msg.group_id) {
    currentMessages.update((msgs) => [...msgs, msg]);
  }
  loadUserGroups();
}


// Cập nhật trạng thái tin nhắn đã đọc khi nhận event chat:read_ack
export function markMessagesAsReadLocally(convID) {
  const active = get(activeConversation);
  if (active && active.conversation?.custom_id === convID) {
    currentMessages.update((msgs) =>
      msgs.map((m) => ({ ...m, is_read: true }))
    );
  }
}

// Cập nhật trạng thái Online / Offline
export function setUserOnlineStatus(userID, isOnline) {
  onlineUsers.update((set) => {
    const next = new Set(set);
    if (isOnline) {
      next.add(userID);
    } else {
      next.delete(userID);
    }
    return next;
  });
}

// Cập nhật trạng thái Đang gõ phím
export function setUserTyping(userID, isTyping) {
  typingUsers.update((set) => {
    const next = new Set(set);
    if (isTyping) {
      next.add(userID);
    } else {
      next.delete(userID);
    }
    return next;
  });
}

// Cập nhật reactions của tin nhắn cục bộ
export function updateMessageReactionsLocally(messageID, reactions) {
  currentMessages.update((msgs) =>
    msgs.map((m) => {
      if (m.id === messageID) {
        return { ...m, reactions: reactions || [] };
      }
      return m;
    })
  );
}

// Đánh dấu tin nhắn đã thu hồi cục bộ
export function markMessageDeletedLocally(messageID) {
  currentMessages.update((msgs) =>
    msgs.map((m) => {
      if (m.id === messageID) {
        return { ...m, is_deleted: true, content: '', file_name: '', file_size: 0 };
      }
      return m;
    })
  );
}

// Cập nhật nội dung tin nhắn đã chỉnh sửa cục bộ
export function updateMessageEditedLocally(messageID, newContent) {
  currentMessages.update((msgs) =>
    msgs.map((m) => {
      if (m.id === messageID) {
        return { ...m, content: newContent, is_edited: true };
      }
      return m;
    })
  );
}

