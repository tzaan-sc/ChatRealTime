<script>
  import { onMount } from 'svelte';
  import { currentUser } from '../stores/auth';
  import { votePoll, closePoll, getPoll } from '../stores/chat';
  import { BarChart2, Check, Lock, EyeOff, CheckSquare, Square, Circle, CheckCircle2 } from 'lucide-svelte';

  export let message = {};

  let poll = message.poll || null;
  let isVoting = false;
  let isClosing = false;
  let errorMsg = '';

  $: if (message.poll) {
    poll = message.poll;
  }

  $: myId = $currentUser?.id;
  $: isCreator = poll && myId && poll.created_by === myId;
  $: isClosed = poll?.is_closed;
  $: totalVotes = poll?.total_votes || 0;

  // Tổng số lượt vote (tổng các option)
  $: sumOptionVotes = (poll?.options || []).reduce((acc, opt) => acc + (opt.vote_count || 0), 0);

  onMount(async () => {
    if (!poll && message.poll_id && message.group_id) {
      try {
        poll = await getPoll(message.group_id, message.poll_id);
      } catch (err) {
        console.error('Lỗi tải thông tin bình chọn:', err);
      }
    }
  });

  async function handleVote(optionId) {
    if (isClosed || isVoting || !poll?.id) return;
    isVoting = true;
    errorMsg = '';
    try {
      const updated = await votePoll(message.group_id, poll.id, optionId);
      poll = updated;
    } catch (err) {
      errorMsg = err.message || 'Không thể bỏ phiếu';
    } finally {
      isVoting = false;
    }
  }

  async function handleClosePoll() {
    if (isClosed || isClosing || !poll?.id) return;
    if (!confirm('Bạn có chắc muốn kết thúc cuộc bình chọn này?')) return;
    isClosing = true;
    try {
      const updated = await closePoll(message.group_id, poll.id);
      poll = updated;
    } catch (err) {
      alert('Lỗi đóng bình chọn: ' + err.message);
    } finally {
      isClosing = false;
    }
  }

  function getPercentage(voteCount) {
    if (!sumOptionVotes || sumOptionVotes === 0) return 0;
    return Math.round((voteCount / sumOptionVotes) * 100);
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="poll-card glass-card" class:closed={isClosed}>
  <!-- Poll Header -->
  <div class="poll-header">
    <div class="header-left">
      <BarChart2 size={16} class="poll-icon" />
      <span class="poll-title">{poll?.question || message.content || 'Khảo sát'}</span>
    </div>
    <div class="header-tags">
      {#if isClosed}
        <span class="tag tag-closed"><Lock size={10} /> Đã đóng</span>
      {:else}
        <span class="tag tag-open">Đang mở</span>
      {/if}

      {#if poll?.is_anonymous}
        <span class="tag tag-anon" title="Bình chọn ẩn danh"><EyeOff size={10} /> Ẩn danh</span>
      {/if}

      {#if poll?.multiple_choice}
        <span class="tag tag-multi">Nhiều lựa chọn</span>
      {/if}
    </div>
  </div>

  {#if errorMsg}
    <div class="poll-error">{errorMsg}</div>
  {/if}

  <!-- Options List -->
  <div class="options-container">
    {#each poll?.options || [] as opt}
      {@const hasVoted = myId && (opt.voter_ids || []).includes(myId)}
      {@const pct = getPercentage(opt.vote_count || 0)}

      <div
        class="poll-option-row"
        class:voted={hasVoted}
        class:disabled={isClosed}
        on:click={() => handleVote(opt.id)}
      >
        <!-- Background Progress Bar -->
        <div class="progress-fill" style="width: {pct}%;"></div>

        <div class="option-content">
          <div class="option-left">
            <div class="check-box" class:checked={hasVoted}>
              {#if poll?.multiple_choice}
                {#if hasVoted}
                  <CheckSquare size={14} />
                {:else}
                  <Square size={14} />
                {/if}
              {:else}
                {#if hasVoted}
                  <CheckCircle2 size={14} />
                {:else}
                  <Circle size={14} />
                {/if}
              {/if}
            </div>
            <span class="opt-text">{opt.text}</span>
          </div>

          <div class="option-right">
            <span class="vote-count">{opt.vote_count || 0} phiếu</span>
            <span class="vote-pct">{pct}%</span>
          </div>
        </div>
      </div>
    {/each}
  </div>

  <!-- Poll Footer -->
  <div class="poll-footer">
    <span class="footer-stats">
      {totalVotes} người đã tham gia ({sumOptionVotes} lượt bầu)
    </span>

    {#if isCreator && !isClosed}
      <button class="close-poll-btn" on:click={handleClosePoll} disabled={isClosing}>
        {isClosing ? 'Đang đóng...' : 'Đóng bình chọn'}
      </button>
    {/if}
  </div>
</div>

<style>
  .poll-card {
    width: 100%;
    max-width: 480px;
    background: rgba(22, 27, 46, 0.75);
    border: 1px solid rgba(129, 140, 248, 0.25);
    border-radius: 12px;
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin: 4px 0;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.25);
  }
  .poll-card.closed {
    border-color: rgba(255, 255, 255, 0.1);
    opacity: 0.9;
  }

  .poll-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .header-left {
    display: flex;
    align-items: flex-start;
    gap: 8px;
  }
  .poll-icon {
    color: #818cf8;
    flex-shrink: 0;
    margin-top: 3px;
  }
  .poll-title {
    font-size: 14px;
    font-weight: 600;
    color: #fff;
    line-height: 1.4;
  }
  .header-tags {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }
  .tag {
    font-size: 10.5px;
    padding: 1px 7px;
    border-radius: 10px;
    font-weight: 500;
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .tag-open {
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.3);
  }
  .tag-closed {
    background: rgba(148, 163, 184, 0.15);
    color: #94a3b8;
    border: 1px solid rgba(148, 163, 184, 0.3);
  }
  .tag-anon {
    background: rgba(245, 158, 11, 0.15);
    color: #fbbf24;
    border: 1px solid rgba(245, 158, 11, 0.3);
  }
  .tag-multi {
    background: rgba(99, 102, 241, 0.15);
    color: #a5b4fc;
    border: 1px solid rgba(99, 102, 241, 0.3);
  }

  .poll-error {
    font-size: 11.5px;
    color: #f87171;
    background: rgba(239, 68, 68, 0.1);
    padding: 4px 8px;
    border-radius: 4px;
  }

  /* Options */
  .options-container {
    display: flex;
    flex-direction: column;
    gap: 7px;
  }
  .poll-option-row {
    position: relative;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    overflow: hidden;
    cursor: pointer;
    transition: 0.15s ease;
  }
  .poll-option-row:hover:not(.disabled) {
    background: rgba(255, 255, 255, 0.07);
    border-color: rgba(129, 140, 248, 0.4);
  }
  .poll-option-row.voted {
    border-color: #818cf8;
  }
  .poll-option-row.disabled {
    cursor: default;
  }

  .progress-fill {
    position: absolute;
    top: 0;
    left: 0;
    bottom: 0;
    background: linear-gradient(90deg, rgba(99, 102, 241, 0.25) 0%, rgba(139, 92, 246, 0.35) 100%);
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    z-index: 1;
  }
  .poll-option-row.voted .progress-fill {
    background: linear-gradient(90deg, rgba(99, 102, 241, 0.35) 0%, rgba(16, 185, 129, 0.35) 100%);
  }

  .option-content {
    position: relative;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 9px 12px;
    gap: 8px;
  }
  .option-left {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
  }
  .check-box {
    color: var(--text-dim);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .check-box.checked {
    color: #818cf8;
  }
  .opt-text {
    font-size: 13px;
    font-weight: 500;
    color: #fff;
  }
  .option-right {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .vote-count {
    font-size: 11px;
    color: var(--text-dim);
  }
  .vote-pct {
    font-size: 12px;
    font-weight: 600;
    color: #cbd5e1;
    min-width: 32px;
    text-align: right;
  }

  /* Footer */
  .poll-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 11.5px;
    color: var(--text-dim);
    padding-top: 4px;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
  }
  .close-poll-btn {
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.25);
    color: #f87171;
    font-size: 11px;
    padding: 3px 8px;
    border-radius: 5px;
    cursor: pointer;
    transition: 0.15s;
  }
  .close-poll-btn:hover {
    background: rgba(239, 68, 68, 0.25);
  }
</style>
