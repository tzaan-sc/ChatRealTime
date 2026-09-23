<script>
  import { wrapMarkdown } from '../utils/markdown';
  import {
    Bold,
    Italic,
    Underline,
    Strikethrough,
    Code,
    FileCode,
    EyeOff
  } from 'lucide-svelte';

  export let textareaRef = null;
  export let onContentChange = () => {};

  function applyFormat(type) {
    if (!textareaRef) return;

    let updated = '';
    switch (type) {
      case 'bold':
        updated = wrapMarkdown(textareaRef, '**', '**', 'chữ đậm');
        break;
      case 'italic':
        updated = wrapMarkdown(textareaRef, '*', '*', 'chữ nghiêng');
        break;
      case 'underline':
        updated = wrapMarkdown(textareaRef, '__', '__', 'gạch chân');
        break;
      case 'strike':
        updated = wrapMarkdown(textareaRef, '~~', '~~', 'gạch ngang');
        break;
      case 'code':
        updated = wrapMarkdown(textareaRef, '`', '`', 'mã');
        break;
      case 'block':
        updated = wrapMarkdown(textareaRef, '```javascript\n', '\n```', 'console.log("hello");');
        break;
      case 'spoiler':
        updated = wrapMarkdown(textareaRef, '||', '||', 'nội dung bí mật');
        break;
    }

    if (onContentChange) {
      onContentChange(updated);
    }
  }
</script>

<div class="md-toolbar">
  <button type="button" class="md-btn" on:click={() => applyFormat('bold')} title="In đậm (**text**)">
    <Bold size={14} />
  </button>
  <button type="button" class="md-btn" on:click={() => applyFormat('italic')} title="In nghiêng (*text*)">
    <Italic size={14} />
  </button>
  <button type="button" class="md-btn" on:click={() => applyFormat('underline')} title="Gạch chân (__text__)">
    <Underline size={14} />
  </button>
  <button type="button" class="md-btn" on:click={() => applyFormat('strike')} title="Gạch ngang (~~text~~)">
    <Strikethrough size={14} />
  </button>
  <div class="toolbar-divider"></div>
  <button type="button" class="md-btn" on:click={() => applyFormat('code')} title="Mã đơn dòng (`code`)">
    <Code size={14} />
  </button>
  <button type="button" class="md-btn" on:click={() => applyFormat('block')} title="Khung mã nguồn (```code```)">
    <FileCode size={14} />
  </button>
  <button type="button" class="md-btn spoiler-btn" on:click={() => applyFormat('spoiler')} title="Ẩn nội dung bí mật (||spoiler||)">
    <EyeOff size={14} />
    <span class="btn-label">Spoiler</span>
  </button>
</div>

<style>
  .md-toolbar {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    background: rgba(15, 23, 42, 0.65);
    border-radius: 8px 8px 0 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .md-btn {
    background: transparent;
    border: none;
    color: var(--text-muted, #94a3b8);
    width: 28px;
    height: 28px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: 0.15s;
  }

  .md-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .spoiler-btn {
    width: auto;
    padding: 0 8px;
    gap: 4px;
  }

  .btn-label {
    font-size: 11px;
    font-weight: 500;
  }

  .toolbar-divider {
    width: 1px;
    height: 16px;
    background: rgba(255, 255, 255, 0.1);
    margin: 0 4px;
  }
</style>
