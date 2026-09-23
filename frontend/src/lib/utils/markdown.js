/**
 * Trình biên dịch Markdown nhẹ, an toàn và tối ưu cho Chat Realtime
 * Hỗ trợ: Bold, Italic, Underline, Strikethrough, Inline Code, Code Block, Spoiler, Links
 */

// Hàm escape HTML chống XSS
export function escapeHtml(text) {
  if (!text) return '';
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

export function parseMarkdown(text) {
  if (!text) return '';

  // 1. Thoát các ký tự HTML nguy hiểm trước
  let parsed = escapeHtml(text);

  // 2. Khung mã nguồn đa dòng (Code block: ```lang\ncode\n```)
  parsed = parsed.replace(/```([a-zA-Z0-9_-]*)\n([\s\S]*?)```/g, (match, lang, code) => {
    return `<div class="md-code-block"><div class="code-header"><span>${lang || 'code'}</span><button class="copy-code-btn" onclick="navigator.clipboard.writeText(this.closest('.md-code-block').querySelector('pre').innerText); this.innerText='Đã sao chép!'; setTimeout(() => this.innerText='Sao chép', 2000);">Sao chép</button></div><pre><code>${code.trim()}</code></pre></div>`;
  });

  // 3. Khung mã nguồn đơn dòng (Inline code: `code`)
  parsed = parsed.replace(/`([^`\n]+)`/g, '<code class="md-inline-code">$1</code>');

  // 4. Nội dung ẩn Spoiler (||nội dung bí mật||)
  parsed = parsed.replace(/\|\|([\s\S]+?)\|\|/g, '<span class="md-spoiler" onclick="this.classList.toggle(\'revealed\')" title="Bấm để xem nội dung bí mật">$1</span>');

  // 5. In đậm (**chữ đậm**)
  parsed = parsed.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

  // 6. In nghiêng (*chữ nghiêng*)
  parsed = parsed.replace(/(?<!\*)\*([^*]+)\*(?!\*)/g, '<em>$1</em>');

  // 7. Gạch chân (<u>gạch chân</u> hoặc __gạch chân__)
  parsed = parsed.replace(/__([^_]+)__/g, '<u>$1</u>');

  // 8. Gạch ngang (~~gạch ngang~~)
  parsed = parsed.replace(/~~([^~]+)~~/g, '<del>$1</del>');

  // 9. Tự động nhận diện liên kết URL an toàn
  parsed = parsed.replace(
    /(https?:\/\/[^\s<]+)/g,
    '<a href="$1" target="_blank" rel="noopener noreferrer" class="md-link">$1</a>'
  );

  // 10. Chuyển đổi dòng mới thành <br>
  parsed = parsed.replace(/\n/g, '<br>');

  return parsed;
}

/**
 * Tiện ích chèn thẻ Markdown vào textarea tại vị trí con trỏ
 */
export function wrapMarkdown(textarea, prefix, suffix = prefix, placeholder = '') {
  if (!textarea) return '';
  const start = textarea.selectionStart;
  const end = textarea.selectionEnd;
  const val = textarea.value;

  const selectedText = val.substring(start, end) || placeholder;
  const replacement = prefix + selectedText + suffix;

  const newVal = val.substring(0, start) + replacement + val.substring(end);
  textarea.value = newVal;

  // Đặt lại con trỏ ở trong phần vừa bao bọc
  const newCursorPos = start + prefix.length + selectedText.length;
  textarea.focus();
  textarea.setSelectionRange(newCursorPos, newCursorPos);

  return newVal;
}
