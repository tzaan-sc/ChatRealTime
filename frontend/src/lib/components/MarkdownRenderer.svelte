<script>
  import { parseMarkdown } from '../utils/markdown';

  export let content = '';

  $: renderedHtml = parseMarkdown(content);
</script>

<div class="markdown-body">
  {@html renderedHtml}
</div>

<style>
  .markdown-body {
    word-break: break-word;
    line-height: 1.5;
    font-size: 14px;
  }

  :global(.md-inline-code) {
    background: rgba(0, 0, 0, 0.35);
    color: #f43f5e;
    padding: 2px 6px;
    border-radius: 4px;
    font-family: 'Consolas', 'Courier New', monospace;
    font-size: 0.9em;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  :global(.md-code-block) {
    margin: 8px 0;
    border-radius: 8px;
    overflow: hidden;
    background: #0f172a;
    border: 1px solid rgba(255, 255, 255, 0.12);
  }

  :global(.code-header) {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4px 10px;
    background: rgba(255, 255, 255, 0.05);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    font-size: 11px;
    color: #94a3b8;
    text-transform: uppercase;
  }

  :global(.copy-code-btn) {
    background: transparent;
    border: none;
    color: #38bdf8;
    cursor: pointer;
    font-size: 11px;
    padding: 2px 6px;
    border-radius: 4px;
    transition: 0.15s;
  }

  :global(.copy-code-btn:hover) {
    background: rgba(56, 189, 248, 0.15);
  }

  :global(.md-code-block pre) {
    margin: 0;
    padding: 10px 12px;
    overflow-x: auto;
    font-family: 'Consolas', 'Courier New', monospace;
    font-size: 12.5px;
    color: #e2e8f0;
  }

  :global(.md-spoiler) {
    background: rgba(255, 255, 255, 0.18);
    color: transparent !important;
    user-select: none;
    cursor: pointer;
    border-radius: 4px;
    padding: 1px 6px;
    filter: blur(5px);
    transition: filter 0.25s, color 0.25s, background 0.25s;
    display: inline-block;
  }

  :global(.md-spoiler:hover) {
    background: rgba(255, 255, 255, 0.25);
  }

  :global(.md-spoiler.revealed) {
    color: inherit !important;
    filter: blur(0);
    background: rgba(244, 63, 94, 0.15);
    user-select: text;
    border: 1px dashed rgba(244, 63, 94, 0.4);
  }

  :global(.md-link) {
    color: #38bdf8;
    text-decoration: underline;
    word-break: break-all;
  }

  :global(.md-link:hover) {
    color: #7dd3fc;
  }
</style>
