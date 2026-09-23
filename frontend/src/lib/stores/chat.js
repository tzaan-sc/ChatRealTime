import { writable, get } from 'svelte/store';
import { apiRequest } from '../services/api';
import { wsService } from '../services/websocket';
import { token, currentUser } from './auth';

export const conversations = writable([]);
export const activeConversation = writable(null); // { conversation, other_user }
export const userGroups = writable([]);
export const activeGroup = writable(null); // Group object
export const currentMessages = writable([]);
export const userDirectory = writable([]);
export const pinnedMessages = writable([]);
export const drafts = writable({}); // map target_id -> content string

// Set các User ID đang online (Ví dụ: Set(['id1', 'id2']))
export const onlineUsers = writable(new Set());

// Set các User ID đang gõ phím vào khung chat
export const typingUsers = writable(new Set());

// Quản lý số lượng tin nhắn chưa đọc cho từng nhóm (groupId -> number)
function loadGroupUnreads() {
  if (typeof window === 'undefined') return {};
  try {
    const raw = localStorage.getItem('chat_group_unreads');
    return raw ? JSON.parse(raw) : {};
  } catch (_) {
    return {};
  }
}

export const groupUnreadCounts = writable(loadGroupUnreads());

export function incrementGroupUnread(groupId) {
  if (!groupId) return;
  groupUnreadCounts.update((counts) => {
    const next = { ...counts, [groupId]: (counts[groupId] || 0) + 1 };
    try {
      localStorage.setItem('chat_group_unreads', JSON.stringify(next));
    } catch (_) {}
    return next;
  });
}

export function clearGroupUnread(groupId) {
  if (!groupId) return;
  groupUnreadCounts.update((counts) => {
    if (!counts[groupId]) return counts;
    const next = { ...counts, [groupId]: 0 };
    try {
      localStorage.setItem('chat_group_unreads', JSON.stringify(next));
    } catch (_) {}
    return next;
  });
}

// Tải danh sách hội thoại
export async function loadConversations() {
  const t = get(token);
  if (!t) return;
  try {
    const res = await apiRequest('/chat/conversations', 'GET', null, t);
    const serverList = res.data || [];

    // Giữ nguyên optimistic unread_count = 0 nếu cuộc trò chuyện đang được mở xem
    const active = get(activeConversation);
    const activeConvId = active?.conversation?.custom_id;

    const merged = serverList.map((item) => {
      if (activeConvId && item.conversation?.custom_id === activeConvId) {
        return { ...item, unread_count: 0 };
      }
      return item;
    });

    conversations.set(merged);
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
  const partnerID = convItem.other_user?.id;

  // 1. Optimistic update: xóa badge chưa đọc ngay tức thì trên UI
  conversations.update((list) =>
    list.map((c) =>
      c.conversation?.custom_id === convID ? { ...c, unread_count: 0 } : c
    )
  );

  // 2. Gửi WebSocket báo đã đọc cho đối phương
  if (partnerID) {
    wsService.send('chat:read', {
      conversation_id: convID,
      partner_id: partnerID
    });
  }

  // 3. Tải tin nhắn & gọi REST API đánh dấu đã xem trên DB
  try {
    const res = await apiRequest(`/chat/messages/${convID}?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
    loadPinnedMessages(convID, false);
    apiRequest(`/chat/messages/${convID}/read`, 'POST', null, t).catch(() => {});
  } catch (err) {
    console.error('Lỗi tải tin nhắn:', err);
  }
}

// Chọn nhóm chat và tải lịch sử tin nhắn nhóm
export async function selectGroup(groupItem) {
  activeConversation.set(null);
  activeGroup.set(groupItem);
  clearGroupUnread(groupItem.id);

  const t = get(token);
  try {
    const res = await apiRequest(`/groups/${groupItem.id}/messages?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
    loadPinnedMessages(groupItem.id, true);
  } catch (err) {
    console.error('Lỗi tải tin nhắn nhóm:', err);
  }
}

// Thêm tin nhắn mới vào danh sách hiện tại nếu đang mở đúng cuộc trò chuyện 1-1
export function appendMessage(msg) {
  const active = get(activeConversation);
  const activeConvId = active?.conversation?.custom_id;
  const isViewingThisChat = active && activeConvId === msg.conversation_id;
  const isTabVisible = typeof document !== 'undefined' && document.visibilityState === 'visible';

  if (isViewingThisChat) {
    if (isTabVisible) {
      // Đang nhìn màn hình chat: tin nhắn tự động chuyển thành đã xem
      msg.is_read = true;
      currentMessages.update((msgs) => [...msgs, msg]);

      // Báo ngay lập tức qua WebSocket cho người gửi
      if (msg.sender_id) {
        wsService.send('chat:read', {
          conversation_id: msg.conversation_id,
          partner_id: msg.sender_id
        });
      }

      // Cập nhật dòng snippet trong danh sách hội thoại nhưng giữ unread_count = 0
      conversations.update((list) =>
        list.map((c) =>
          c.conversation?.custom_id === msg.conversation_id
            ? {
                ...c,
                conversation: {
                  ...c.conversation,
                  last_message: msg.content,
                  last_message_at: msg.created_at
                },
                unread_count: 0
              }
            : c
        )
      );
    } else {
      // Đang chọn hội thoại nhưng thu nhỏ tab: vẫn append nhưng chưa xem
      currentMessages.update((msgs) => [...msgs, msg]);

      conversations.update((list) =>
        list.map((c) =>
          c.conversation?.custom_id === msg.conversation_id
            ? {
                ...c,
                conversation: {
                  ...c.conversation,
                  last_message: msg.content,
                  last_message_at: msg.created_at
                },
                unread_count: (c.unread_count || 0) + 1
              }
            : c
        )
      );
    }
  } else {
    // Nhận tin nhắn từ cuộc trò chuyện khác
    let found = false;
    conversations.update((list) =>
      list.map((c) => {
        if (c.conversation?.custom_id === msg.conversation_id) {
          found = true;
          return {
            ...c,
            conversation: {
              ...c.conversation,
              last_message: msg.content,
              last_message_at: msg.created_at
            },
            unread_count: (c.unread_count || 0) + 1
          };
        }
        return c;
      })
    );

    // Nếu hội thoại mới tinh chưa có trong danh sách thì nạp lại
    if (!found) {
      loadConversations();
    }
  }
}

// Thêm tin nhắn nhóm mới vào danh sách nếu đang mở đúng nhóm
export function appendGroupMessage(msg) {
  const grp = get(activeGroup);
  const isViewingThisGroup = grp && grp.id === msg.group_id;
  const isTabVisible = typeof document !== 'undefined' && document.visibilityState === 'visible';
  const me = get(currentUser);

  if (isViewingThisGroup) {
    currentMessages.update((msgs) => [...msgs, msg]);
    if (!isTabVisible && msg.sender_id !== me?.id) {
      incrementGroupUnread(msg.group_id);
    }
  } else {
    if (msg.sender_id !== me?.id) {
      incrementGroupUnread(msg.group_id);
    }
  }
}

// Đánh dấu tin nhắn đã đọc cục bộ khi nhận event chat:read_ack
export function markMessagesAsReadLocally(convID) {
  const active = get(activeConversation);
  if (active && active.conversation?.custom_id === convID) {
    currentMessages.update((msgs) =>
      msgs.map((m) => ({ ...m, is_read: true }))
    );
  }
}

// Đánh dấu cuộc trò chuyện là chưa đọc (Mark as Unread)
export async function markConversationAsUnread(convID) {
  if (!convID) return;
  const t = get(token);

  // 1. Cập nhật giao diện lập tức: tăng unread_count lên 1
  conversations.update((list) =>
    list.map((c) =>
      c.conversation?.custom_id === convID
        ? { ...c, unread_count: Math.max(1, (c.unread_count || 0) + 1) }
        : c
    )
  );

  // Nếu đang mở đúng cuộc trò chuyện này, đóng khung chat để user thấy rõ trạng thái chưa đọc
  const active = get(activeConversation);
  if (active && active.conversation?.custom_id === convID) {
    activeConversation.set(null);
  }

  // 2. Gửi tín hiệu WebSocket & REST API lên máy chủ
  wsService.send('chat:unread', { conversation_id: convID });
  if (t) {
    try {
      await apiRequest(`/chat/messages/${convID}/unread`, 'POST', null, t);
    } catch (e) {
      console.warn('Lỗi gọi API mark unread:', e);
    }
  }
}

// Đánh dấu cuộc trò chuyện là đã đọc (Mark as Read)
export async function markConversationAsRead(convID, partnerID) {
  if (!convID) return;
  const t = get(token);

  // 1. Cập nhật giao diện lập tức: unread_count = 0
  conversations.update((list) =>
    list.map((c) =>
      c.conversation?.custom_id === convID ? { ...c, unread_count: 0 } : c
    )
  );

  // 2. Gửi tín hiệu WebSocket & REST API lên máy chủ
  if (partnerID) {
    wsService.send('chat:read', {
      conversation_id: convID,
      partner_id: partnerID
    });
  }

  if (t) {
    try {
      await apiRequest(`/chat/messages/${convID}/read`, 'POST', null, t);
    } catch (e) {
      console.warn('Lỗi gọi API mark read:', e);
    }
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

// Tải danh sách tin nhắn ghim
export async function loadPinnedMessages(targetId, isGroup) {
  const t = get(token);
  if (!t || !targetId) return;
  try {
    const res = await apiRequest(`/chat/pinned/${targetId}?is_group=${isGroup}`, 'GET', null, t);
    pinnedMessages.set(res.data || []);
  } catch (e) {
    console.error('Lỗi tải tin nhắn ghim:', e);
  }
}

// Cập nhật trạng thái ghim tin nhắn cục bộ
export function updatePinnedLocally(msgId, isPinned) {
  currentMessages.update((msgs) =>
    msgs.map((m) => {
      if (m.id === msgId) {
        return { ...m, is_pinned: isPinned };
      }
      return m;
    })
  );

  pinnedMessages.update((list) => {
    if (isPinned) {
      const allMsgs = get(currentMessages);
      const found = allMsgs.find((m) => m.id === msgId);
      if (found && !list.some((p) => p.id === msgId)) {
        return [found, ...list];
      }
      return list;
    } else {
      return list.filter((p) => p.id !== msgId);
    }
  });
}

// Cập nhật số lượng phản hồi luồng thảo luận
export function updateThreadCountLocally(rootId, newCount) {
  currentMessages.update((msgs) =>
    msgs.map((m) => {
      if (m.id === rootId) {
        return { ...m, thread_count: newCount };
      }
      return m;
    })
  );
}

// Tải các bản nháp từ server
export async function loadDrafts() {
  const t = get(token);
  if (!t) return;
  try {
    const res = await apiRequest('/chat/drafts', 'GET', null, t);
    const map = {};
    for (const d of res.data || []) {
      if (d.content && d.content.trim()) {
        map[d.target_id] = d.content;
      }
    }
    drafts.set(map);
  } catch (e) {
    console.error('Lỗi tải tin nháp:', e);
  }
}

// Lưu tin nhắn nháp (cục bộ + server)
export function saveDraftLocallyAndRemote(targetId, content, targetType) {
  if (!targetId) return;
  const t = get(token);

  drafts.update((d) => {
    const next = { ...d };
    if (content && content.trim()) {
      next[targetId] = content;
    } else {
      delete next[targetId];
    }
    return next;
  });

  // Gửi WebSocket đồng bộ đa thiết bị
  wsService.send('draft:sync', {
    target_id: targetId,
    target_type: targetType,
    content: content || ''
  });

  if (t) {
    if (content && content.trim()) {
      apiRequest('/chat/drafts', 'POST', { target_id: targetId, target_type: targetType, content }, t).catch(() => {});
    } else {
      apiRequest(`/chat/drafts/${targetId}`, 'DELETE', null, t).catch(() => {});
    }
  }
}

// Mở Hộp thư lưu trữ cá nhân (Saved Messages / Cloud Notebook)
export async function openSavedMessages() {
  const me = get(currentUser);
  if (!me) return;

  activeGroup.set(null);
  const savedConv = {
    conversation: {
      custom_id: `saved_${me.id}`,
      members: [me.id],
      last_message: 'Hộp thư lưu trữ cá nhân'
    },
    other_user: {
      id: me.id,
      username: me.username,
      display_name: 'Tin nhắn đã lưu (Cloud Notebook)',
      avatar_url: me.avatar_url || ''
    },
    is_saved: true,
    unread_count: 0
  };

  activeConversation.set(savedConv);
  const t = get(token);
  try {
    const res = await apiRequest(`/chat/messages/saved_${me.id}?limit=50`, 'GET', null, t);
    currentMessages.set(res.data || []);
    loadPinnedMessages(`saved_${me.id}`, false);
  } catch (err) {
    console.error('Lỗi tải Saved Messages:', err);
  }
}

