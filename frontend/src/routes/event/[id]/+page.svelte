<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api';
  import dayjs from 'dayjs';
  import 'dayjs/locale/ja';
  import copy from 'copy-to-clipboard';
  import { toast } from '@zerodevx/svelte-toast';

  dayjs.locale('ja');

  interface CandidateAnswer {
    candidate_id: number;
    answer_status: 'ok' | 'maybe' | 'ng';
  }

  interface Response {
    id: number;
    respondent_name: string;
    comment: string;
    created_at: string;
    answers: CandidateAnswer[];
  }

  interface EventCandidate {
    id: number;
    event_date: string;
    start_time: string;
    end_time: string;
  }

  interface Event {
    id: string;
    title: string;
    description: string;
    status: string;
    confirmed_candidate_id?: number;
    confirmed_candidate?: EventCandidate;
    created_at: string;
    candidates: EventCandidate[];
    responses: Response[];
  }

  const eventId = $page.params.id;

  // Svelte 5 Runes for Reactivity
  let event = $state<Event | null>(null);
  let loading = $state(true);
  let errorMsg = $state('');

  // Form states
  let respondentName = $state('');
  let comment = $state('');
  let myAnswers = $state<Record<number, 'ok' | 'maybe' | 'ng'>>({});
  let submitting = $state(false);

  // Load Event Details
  async function loadEvent() {
    try {
      const data = await api.get<Event>(`/events/${eventId}`);
      event = data;
      
      // 回答フォームの初期値（すべて〇にする）
      const initialAnswers: Record<number, 'ok' | 'maybe' | 'ng'> = {};
      data.candidates.forEach(cand => {
        initialAnswers[cand.id] = 'ok';
      });
      myAnswers = initialAnswers;
    } catch (err: any) {
      errorMsg = err.message || 'イベントの読み込みに失敗しました。URLが正しいか確認してください。';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadEvent();
  });

  // Derived state: 候補日ごとの〇△×の集計 (Svelte 5 Runes)
  let candidateStats = $derived.by(() => {
    if (!event) return {};
    
    const stats: Record<number, { ok: number; maybe: number; ng: number; score: number }> = {};
    
    event.candidates.forEach(cand => {
      stats[cand.id] = { ok: 0, maybe: 0, ng: 0, score: 0 };
    });

    event.responses.forEach(resp => {
      resp.answers.forEach(ans => {
        if (stats[ans.candidate_id]) {
          stats[ans.candidate_id][ans.answer_status]++;
          // スコア計算 (ok=2, maybe=1, ng=0)
          if (ans.answer_status === 'ok') stats[ans.candidate_id].score += 2;
          if (ans.answer_status === 'maybe') stats[ans.candidate_id].score += 1;
        }
      });
    });

    return stats;
  });

  // 回答ステータスのヘルパー
  const statusConfig = {
    ok: { label: '〇', class: 'status-ok' },
    maybe: { label: '△', class: 'status-maybe' },
    ng: { label: '×', class: 'status-ng' }
  };

  function selectAnswer(candidateId: number, status: 'ok' | 'maybe' | 'ng') {
    myAnswers = {
      ...myAnswers,
      [candidateId]: status
    };
  }

  async function submitResponse(e: SubmitEvent) {
    e.preventDefault();
    if (!respondentName.trim()) {
      toast.push('お名前を入力してください', {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
      return;
    }

    submitting = true;
    
    // リクエスト用に配列化
    const answersArray = Object.entries(myAnswers).map(([candId, status]) => ({
      candidate_id: Number(candId),
      answer_status: status
    }));

    try {
      await api.post(`/events/${eventId}/responses`, {
        respondent_name: respondentName,
        comment,
        answers: answersArray
      });

      toast.push('回答を登録しました！');
      
      // フォームのクリア
      respondentName = '';
      comment = '';
      
      // 再ロード
      await loadEvent();
    } catch (err: any) {
      toast.push('回答の登録に失敗しました: ' + err.message, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    } finally {
      submitting = false;
    }
  }

  async function deleteResponse(responseId: number, name: string) {
    if (!confirm(`「${name}」さんの回答を削除しますか？`)) {
      return;
    }
    
    try {
      await api.delete(`/events/${eventId}/responses/${responseId}`);
      toast.push('回答を削除しました');
      await loadEvent();
    } catch (err: any) {
      toast.push('回答の削除に失敗しました: ' + err.message, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    }
  }

  function formatDateTime(dateStr: string, startStr: string, endStr: string): string {
    const d = dayjs(dateStr);
    const formattedDate = d.format('M/D(ddd)');
    return `${formattedDate} ${startStr.slice(0, 5)}〜${endStr.slice(0, 5)}`;
  }
</script>

<div class="container event-page-container animate-fade-in">
  {#if loading}
    <div class="glass-panel loading-panel">
      <div class="spinner"></div>
      <p>イベント情報をロード中...</p>
    </div>
  {:else if errorMsg}
    <div class="glass-panel error-panel-large">
      <span class="material-symbols-rounded error-icon-lg">warning</span>
      <h2>エラーが発生しました</h2>
      <p>{errorMsg}</p>
      <a href="/" class="btn btn-primary">トップページに戻る</a>
    </div>
  {:else if event}
    <!-- Event Details Panel -->
    <div class="event-header-card glass-panel">
      <div class="event-status-header">
        <span class="status-badge" class:confirmed={event.status === 'confirmed'}>
          {event.status === 'confirmed' ? '日程確定済み' : '日程調整中'}
        </span>
      </div>
      <h1 class="event-title">{event.title}</h1>
      <p class="event-desc-text">{event.description || '説明はありません。'}</p>
      
      {#if event.status === 'confirmed' && event.confirmed_candidate}
        <div class="confirmed-box glass-panel animate-fade-in">
          <span class="material-symbols-rounded confirmed-icon">celebration</span>
          <div class="confirmed-info">
            <p class="confirmed-label">確定した日時</p>
            <h3 class="confirmed-time">
              {formatDateTime(event.confirmed_candidate.event_date, event.confirmed_candidate.start_time, event.confirmed_candidate.end_time)}
            </h3>
          </div>
        </div>
      {/if}

      <div class="share-box">
        <label for="share-url">回答共有URL</label>
        <div class="copy-input-group">
          <input type="text" id="share-url" readonly value={window.location.href} />
          <button 
            class="btn btn-secondary" 
            onclick={() => {
              copy(window.location.href);
              toast.push('URLをコピーしました！');
            }}
          >
            コピー
          </button>
        </div>
      </div>
    </div>

    <!-- Table of Answers Section -->
    <div class="answers-table-section glass-panel">
      <h3 class="section-subtitle">みんなの回答状況</h3>
      
      <div class="table-wrapper">
        <table class="answers-table">
          <thead>
            <tr>
              <th class="sticky-col">候補日程</th>
              {#each event.responses as resp}
                <th>
                  <div class="respondent-header">
                    <span class="respondent-name">{resp.respondent_name}</span>
                    <button 
                      class="delete-resp-btn" 
                      title="回答を削除"
                      onclick={() => deleteResponse(resp.id, resp.respondent_name)}
                    >
                      <span class="material-symbols-rounded">close</span>
                    </button>
                  </div>
                </th>
              {/each}
              <th class="stats-col-header">〇 △ ×</th>
            </tr>
          </thead>
          <tbody>
            {#each event.candidates as cand}
              <tr class:row-confirmed={event.status === 'confirmed' && event.confirmed_candidate_id === cand.id}>
                <td class="sticky-col datetime-cell">
                  {formatDateTime(cand.event_date, cand.start_time, cand.end_time)}
                  {#if event.status === 'confirmed' && event.confirmed_candidate_id === cand.id}
                    <span class="row-confirmed-badge">確定日</span>
                  {/if}
                </td>
                
                {#each event.responses as resp}
                  {@const userAns = resp.answers.find(a => a.candidate_id === cand.id)}
                  <td class="status-cell">
                    {#if userAns}
                      <span class={`status-indicator ${statusConfig[userAns.answer_status].class}`}>
                        {statusConfig[userAns.answer_status].label}
                      </span>
                    {:else}
                      <span class="status-indicator">-</span>
                    {/if}
                  </td>
                {/each}

                <!-- Stats summary -->
                <td class="stats-cell">
                  <div class="stats-summary-box">
                    <span class="stat-badge ok">〇 {candidateStats[cand.id]?.ok || 0}</span>
                    <span class="stat-badge maybe">△ {candidateStats[cand.id]?.maybe || 0}</span>
                    <span class="stat-badge ng">× {candidateStats[cand.id]?.ng || 0}</span>
                  </div>
                </td>
              </tr>
            {/each}

            <!-- Comment row -->
            <tr class="comment-row">
              <td class="sticky-col comment-label-cell">コメント</td>
              {#each event.responses as resp}
                <td class="comment-cell">
                  {#if resp.comment}
                    <div class="comment-tooltip-trigger" title={resp.comment}>
                      <span class="material-symbols-rounded">chat_bubble</span>
                      <p class="comment-preview">{resp.comment}</p>
                    </div>
                  {:else}
                    <span class="no-comment">-</span>
                  {/if}
                </td>
              {/each}
              <td class="comment-empty"></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Response Input Form -->
    {#if event.status !== 'confirmed'}
      <div class="input-form-card glass-panel">
        <h3 class="section-subtitle">日程を回答する</h3>
        
        <form onsubmit={submitResponse}>
          <div class="form-grid">
            <div class="form-group">
              <label for="name">回答者のお名前</label>
              <input type="text" id="name" placeholder="e.g. 田中" bind:value={respondentName} required />
            </div>
            
            <div class="form-group">
              <label for="comment">コメント（希望条件や一言）</label>
              <input type="text" id="comment" placeholder="e.g. 19:30以降なら参加できます！" bind:value={comment} />
            </div>
          </div>

          <div class="response-dates-picker">
            <label>各日程への都合</label>
            <div class="date-picker-list">
              {#each event.candidates as cand}
                <div class="picker-row">
                  <span class="picker-date">{formatDateTime(cand.event_date, cand.start_time, cand.end_time)}</span>
                  <div class="btn-group-status">
                    <button 
                      type="button" 
                      class="btn btn-secondary btn-status-choice active-ok"
                      class:active={myAnswers[cand.id] === 'ok'}
                      onclick={() => selectAnswer(cand.id, 'ok')}
                    >
                      〇 可
                    </button>
                    <button 
                      type="button" 
                      class="btn btn-secondary btn-status-choice active-maybe"
                      class:active={myAnswers[cand.id] === 'maybe'}
                      onclick={() => selectAnswer(cand.id, 'maybe')}
                    >
                      △ 条件付
                    </button>
                    <button 
                      type="button" 
                      class="btn btn-secondary btn-status-choice active-ng"
                      class:active={myAnswers[cand.id] === 'ng'}
                      onclick={() => selectAnswer(cand.id, 'ng')}
                    >
                      × 不可
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <button type="submit" class="btn btn-primary btn-lg submit-resp-btn" disabled={submitting}>
            <span class="material-symbols-rounded">save</span>
            {submitting ? '送信中...' : 'この内容で回答を登録する'}
          </button>
        </form>
      </div>
    {/if}
  {/if}
</div>

<style>
  .event-page-container {
    display: flex;
    flex-direction: column;
    gap: 2rem;
    padding: 1.5rem 0;
  }

  /* Header Card */
  .event-header-card {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .event-status-header {
    display: flex;
    justify-content: flex-start;
  }

  .event-title {
    font-size: 2.2rem;
  }

  .event-desc-text {
    color: var(--text-secondary);
    font-size: 1.05rem;
    white-space: pre-wrap;
  }

  .confirmed-box {
    display: flex;
    align-items: center;
    gap: 1.25rem;
    background: hsla(150, 84%, 37%, 0.1);
    border-color: var(--color-ok);
    padding: 1.5rem 2rem;
  }

  .confirmed-icon {
    font-size: 3rem;
    color: var(--color-ok);
  }

  .confirmed-label {
    color: var(--text-secondary);
    font-size: 0.85rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  .confirmed-time {
    font-size: 1.6rem;
    color: var(--color-ok);
  }

  .share-box {
    margin-top: 1rem;
    border-top: 1px solid var(--border-glass);
    padding-top: 1.5rem;
  }

  .copy-input-group {
    display: flex;
    gap: 0.75rem;
  }

  .copy-input-group input {
    flex: 1;
    background: hsla(223, 40%, 6%, 0.8);
  }

  /* Table styles */
  .table-wrapper {
    overflow-x: auto;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-glass);
  }

  .answers-table {
    width: 100%;
    border-collapse: collapse;
    text-align: left;
  }

  .answers-table th, .answers-table td {
    padding: 1rem;
    border-bottom: 1px solid var(--border-glass);
  }

  .answers-table th {
    background: hsla(223, 40%, 10%, 0.9);
    font-family: var(--font-display);
    font-weight: 600;
    font-size: 0.95rem;
    white-space: nowrap;
  }

  .sticky-col {
    position: sticky;
    left: 0;
    background: var(--bg-secondary);
    z-index: 10;
    min-width: 200px;
    border-right: 1px solid var(--border-glass);
  }

  .datetime-cell {
    font-family: var(--font-display);
    font-weight: 600;
  }

  .row-confirmed {
    background: hsla(150, 84%, 37%, 0.05);
  }

  .row-confirmed-badge {
    background: var(--color-ok);
    color: #fff;
    font-size: 0.7rem;
    padding: 0.15rem 0.4rem;
    border-radius: var(--radius-sm);
    margin-left: 0.5rem;
    font-weight: 700;
  }

  .respondent-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .respondent-name {
    font-weight: 600;
  }

  .delete-resp-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    display: flex;
    padding: 0.1rem;
    border-radius: var(--radius-full);
    transition: all var(--transition-fast);
  }

  .delete-resp-btn:hover {
    color: var(--color-ng);
    background: hsla(350, 89%, 60%, 0.1);
  }

  .delete-resp-btn .material-symbols-rounded {
    font-size: 0.95rem;
  }

  .status-cell {
    text-align: center;
  }

  .status-indicator {
    font-size: 1.3rem;
    font-weight: 800;
    display: inline-block;
  }

  .status-ok { color: var(--color-ok); }
  .status-maybe { color: var(--color-maybe); }
  .status-ng { color: var(--color-ng); }

  .stats-col-header {
    min-width: 140px;
    text-align: center;
  }

  .stats-cell {
    text-align: center;
  }

  .stats-summary-box {
    display: inline-flex;
    gap: 0.5rem;
  }

  .stat-badge {
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.2rem 0.5rem;
    border-radius: var(--radius-sm);
  }

  .stat-badge.ok { background: hsla(150, 84%, 37%, 0.15); color: var(--color-ok); }
  .stat-badge.maybe { background: hsla(38, 92%, 50%, 0.15); color: var(--color-maybe); }
  .stat-badge.ng { background: hsla(350, 89%, 60%, 0.15); color: var(--color-ng); }

  .comment-row td {
    border-bottom: none;
    padding: 0.75rem 1rem;
  }

  .comment-label-cell {
    color: var(--text-secondary);
    font-size: 0.85rem;
    font-weight: 600;
  }

  .comment-cell {
    font-size: 0.85rem;
    color: var(--text-secondary);
    text-align: center;
  }

  .comment-tooltip-trigger {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.35rem;
    cursor: help;
  }

  .comment-tooltip-trigger .material-symbols-rounded {
    font-size: 1rem;
    color: var(--color-primary);
  }

  .comment-preview {
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: inline-block;
  }

  .no-comment {
    color: var(--text-muted);
  }

  /* Response Picker */
  .response-dates-picker {
    margin: 2rem 0;
  }

  .date-picker-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 0.5rem;
  }

  .picker-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: hsla(223, 40%, 10%, 0.5);
    border: 1px solid var(--border-glass);
    padding: 0.75rem 1.5rem;
    border-radius: var(--radius-sm);
  }

  @media (max-width: 600px) {
    .picker-row {
      flex-direction: column;
      align-items: flex-start;
      gap: 1rem;
    }
  }

  .picker-date {
    font-weight: 600;
  }

  .btn-group-status {
    display: flex;
    gap: 0.5rem;
  }

  .btn-status-choice {
    font-size: 0.85rem;
    padding: 0.5rem 1rem;
  }

  .btn-status-choice.active-ok.active {
    background-color: var(--color-ok);
    color: #fff;
    border-color: var(--color-ok);
  }

  .btn-status-choice.active-maybe.active {
    background-color: var(--color-maybe);
    color: #fff;
    border-color: var(--color-maybe);
  }

  .btn-status-choice.active-ng.active {
    background-color: var(--color-ng);
    color: #fff;
    border-color: var(--color-ng);
  }

  .submit-resp-btn {
    width: 100%;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1.5fr;
    gap: 1.5rem;
  }

  @media (max-width: 600px) {
    .form-grid {
      grid-template-columns: 1fr;
    }
  }

  .error-panel-large {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 5rem 2rem;
    text-align: center;
    gap: 1rem;
  }

  .error-icon-lg {
    font-size: 4rem;
    color: var(--color-ng);
  }
</style>
