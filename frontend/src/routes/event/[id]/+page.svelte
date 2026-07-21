<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api';
  import { Accordion, AccordionItem } from '$lib';
  import dayjs from 'dayjs';
  import 'dayjs/locale/ja';
  import copy from 'copy-to-clipboard';
  import { toast } from '@zerodevx/svelte-toast';

  import { reveal } from '$lib/reveal';

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
    edit_token?: string; // 新規登録時に返却される
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

  // Edit Mode states (Passwordless response updates)
  let isEditing = $state(false);
  let editingResponseId = $state<number | null>(null);
  let responseTokens = $state<Record<number, string>>({}); // { responseId: editToken }

  let shareUrl = $state('');

  // Load Stored Tokens on Mount
  onMount(() => {
    loadEvent();
    shareUrl = window.location.href;
    const stored = localStorage.getItem('kanji_chan_response_tokens');
    if (stored) {
      try {
        responseTokens = JSON.parse(stored);
      } catch {
        responseTokens = {};
      }
    }
  });

  function isMyResponse(responseId: number): boolean {
    return !!responseTokens[responseId];
  }

  // Load Event Details
  async function loadEvent() {
    try {
      const data = await api.get<Event>(`/events/${eventId}`);
      event = data;
      
      // 回答フォームの初期値（すべて〇にする）
      const initialAnswers: Record<number, 'ok' | 'maybe' | 'ng'> = {};
      const candidates = data.candidates || [];
      candidates.forEach(cand => {
        initialAnswers[cand.id] = 'ok';
      });
      myAnswers = initialAnswers;
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'イベントの読み込みに失敗しました。';
      errorMsg = msg + ' URLが正しいか確認してください。';
    } finally {
      loading = false;
    }
  }

  // Derived state: 候補日ごとの〇△×の集計 (Svelte 5 Runes)
  let candidateStats = $derived.by(() => {
    try {
      if (!event) return {};
      
      const stats: Record<number, { ok: number; maybe: number; ng: number; score: number }> = {};
      
      const candidates = event.candidates || [];
      candidates.forEach(cand => {
        stats[cand.id] = { ok: 0, maybe: 0, ng: 0, score: 0 };
      });

      const responses = event.responses || [];
      responses.forEach(resp => {
        const answers = resp.answers || [];
        answers.forEach(ans => {
          if (ans && stats[ans.candidate_id]) {
            stats[ans.candidate_id][ans.answer_status]++;
            // スコア計算 (ok=2, maybe=1, ng=0)
            if (ans.answer_status === 'ok') stats[ans.candidate_id].score += 2;
            if (ans.answer_status === 'maybe') stats[ans.candidate_id].score += 1;
          }
        });
      });

      return stats;
    } catch (err) {
      console.error("Error in candidateStats derived state:", err);
      return {};
    }
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

  function startEdit(resp: Response) {
    isEditing = true;
    editingResponseId = resp.id;
    respondentName = resp.respondent_name;
    comment = resp.comment;

    // 既存回答の読み込み
    const initialAnswers: Record<number, 'ok' | 'maybe' | 'ng'> = {};
    event?.candidates.forEach(cand => {
      initialAnswers[cand.id] = 'ok'; // fallback
    });
    resp.answers.forEach(ans => {
      initialAnswers[ans.candidate_id] = ans.answer_status;
    });
    myAnswers = initialAnswers;

    // フォームへスクロール
    const formEl = document.querySelector('.input-form-card');
    if (formEl) {
      formEl.scrollIntoView({ behavior: 'smooth' });
    }
  }

  function cancelEdit() {
    isEditing = false;
    editingResponseId = null;
    respondentName = '';
    comment = '';
    const initialAnswers: Record<number, 'ok' | 'maybe' | 'ng'> = {};
    event?.candidates.forEach(cand => {
      initialAnswers[cand.id] = 'ok';
    });
    myAnswers = initialAnswers;
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
    const answersArray = Object.entries(myAnswers).map(([candId, status]) => ({
      candidate_id: Number(candId),
      answer_status: status
    }));

    try {
      if (isEditing && editingResponseId !== null) {
        // 編集・更新リクエスト
        const token = responseTokens[editingResponseId];
        const headers = token ? { 'X-Response-Token': token } : undefined;
        await api.put<Response>(
          `/events/${eventId}/responses/${editingResponseId}`,
          {
            respondent_name: respondentName,
            comment,
            answers: answersArray
          },
          { headers }
        );

        toast.push('回答を更新しました！');
        cancelEdit();
      } else {
        // 新規回答登録
        const created = await api.post<Response>(`/events/${eventId}/responses`, {
          respondent_name: respondentName,
          comment,
          answers: answersArray
        });

        toast.push('回答を登録しました！');

        // レスポンスから返却された edit_token を LocalStorage に保存
        if (created.id && created.edit_token) {
          responseTokens = {
            ...responseTokens,
            [created.id]: created.edit_token
          };
          localStorage.setItem('kanji_chan_response_tokens', JSON.stringify(responseTokens));
        }

        respondentName = '';
        comment = '';
      }

      await loadEvent();
    } catch (err) {
      const msg = err instanceof Error ? err.message : '回答の送信に失敗しました。';
      toast.push('送信に失敗しました: ' + msg, {
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
      const token = responseTokens[responseId];
      const headers = token ? { 'X-Response-Token': token } : undefined;
      
      await api.delete(`/events/${eventId}/responses/${responseId}`, { headers });
      toast.push('回答を削除しました');
      
      // 保存トークンをクリーンアップ
      if (responseTokens[responseId]) {
        const nextTokens = { ...responseTokens };
        delete nextTokens[responseId];
        responseTokens = nextTokens;
        localStorage.setItem('kanji_chan_response_tokens', JSON.stringify(responseTokens));
      }

      await loadEvent();
    } catch (err) {
      const msg = err instanceof Error ? err.message : '削除に失敗しました。';
      toast.push('回答の削除に失敗しました: ' + msg, {
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

<svelte:head>
  <title>{event ? `「${event.title}」の出欠入力・日程調整 | 幹事ちゃん` : '日程調整 | 幹事ちゃん'}</title>
  {#if event}
    <meta property="og:type" content="website" />
    <meta property="og:title" content={`「${event.title}」の日程調整 | 幹事ちゃん`} />
    <meta property="og:description" content={event.description || '幹事ちゃんで出欠回答をお願いします。'} />
    <meta property="og:image" content={`${shareUrl.split('/event/')[0]}/api/ogp/${event.id}.png`} />
    <meta property="og:image:type" content="image/svg+xml" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content={`「${event.title}」の日程調整 | 幹事ちゃん`} />
    <meta name="twitter:description" content={event.description || '幹事ちゃんで出欠回答をお願いします。'} />
    <meta name="twitter:image" content={`${shareUrl.split('/event/')[0]}/api/ogp/${event.id}.png`} />
  {/if}
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="container event-page-container" use:reveal>
  {#if loading}
    <div class="glass-panel loading-panel" role="status" aria-label="読み込み中">
      <div class="spinner"></div>
      <p>イベント情報をロード中...</p>
    </div>
  {:else if errorMsg}
    <div class="glass-panel error-panel-large" role="alert">
      <span class="material-symbols-rounded error-icon-lg" aria-hidden="true">warning</span>
      <h2>エラーが発生しました</h2>
      <p>{errorMsg}</p>
      <a href="/" class="btn btn-primary">トップページに戻る</a>
    </div>
  {:else if event}
    <!-- Event Details Panel -->
    <div class="event-header-card glass-panel" use:reveal>
      <div class="event-status-header">
        <span class="status-badge" class:confirmed={event.status === 'confirmed'}>
          {event.status === 'confirmed' ? '日程確定済み' : '日程調整中'}
        </span>
      </div>
      <h1 class="event-title">{event.title}</h1>
      <p class="event-desc-text">{event.description || '説明はありません。'}</p>
      
      {#if event.status === 'confirmed' && event.confirmed_candidate}
        <div class="confirmed-box glass-panel" use:reveal>
          <span class="material-symbols-rounded confirmed-icon" aria-hidden="true">celebration</span>
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
          <input type="text" id="share-url" readonly value={shareUrl} />
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
    <div class="answers-table-section glass-panel" use:reveal>
      <h3 class="section-subtitle">みんなの回答状況</h3>
      
      <div class="table-wrapper">
        <table class="answers-table">
          <thead>
            <tr>
              <th scope="col" class="sticky-col">候補日程</th>
              {#each event.responses || [] as resp}
                <th scope="col" class:th-my-response={isMyResponse(resp.id)}>
                  <div class="respondent-header">
                    <div class="respondent-name-wrapper">
                      {#if isMyResponse(resp.id)}
                        <span class="my-response-badge">あなた</span>
                      {/if}
                      <span class="respondent-name">{resp.respondent_name}</span>
                    </div>
                    <div class="header-action-group">
                      <button 
                        class="edit-resp-btn" 
                        class:my-btn={isMyResponse(resp.id)}
                        title="回答を編集"
                        aria-label={`「${resp.respondent_name}」の回答を編集`}
                        onclick={() => startEdit(resp)}
                      >
                        <span class="material-symbols-rounded" aria-hidden="true">edit</span>
                      </button>
                      <button 
                        class="delete-resp-btn" 
                        class:my-btn={isMyResponse(resp.id)}
                        title="回答を削除"
                        aria-label={`「${resp.respondent_name}」の回答を削除`}
                        onclick={() => deleteResponse(resp.id, resp.respondent_name)}
                      >
                        <span class="material-symbols-rounded" aria-hidden="true">close</span>
                      </button>
                    </div>
                  </div>
                </th>
              {/each}
              <th scope="col" class="stats-col-header">〇 △ ×</th>
            </tr>
          </thead>
          <tbody>
            {#each event.candidates || [] as cand}
              <tr class:row-confirmed={event.status === 'confirmed' && event.confirmed_candidate_id === cand.id}>
                <th scope="row" class="sticky-col datetime-cell">
                  {formatDateTime(cand.event_date, cand.start_time, cand.end_time)}
                  {#if event.status === 'confirmed' && event.confirmed_candidate_id === cand.id}
                    <span class="row-confirmed-badge">確定日</span>
                  {/if}
                </th>
                
                {#each event.responses || [] as resp}
                  {@const userAns = resp.answers.find(a => a.candidate_id === cand.id)}
                  <td class="status-cell" class:cell-my-response={isMyResponse(resp.id)}>
                    {#if userAns}
                      <span 
                        class={`status-indicator ${statusConfig[userAns.answer_status].class}`}
                        aria-label={userAns.answer_status === 'ok' ? '可' : userAns.answer_status === 'maybe' ? '条件付き可' : '不可'}
                      >
                        {statusConfig[userAns.answer_status].label}
                      </span>
                    {:else}
                      <span class="status-indicator" aria-label="未回答">-</span>
                    {/if}
                  </td>
                {/each}

                <!-- Stats summary -->
                <td class="stats-cell">
                  <div class="stats-summary-box" aria-label={`〇 ${candidateStats[cand.id]?.ok || 0}件、△ ${candidateStats[cand.id]?.maybe || 0}件、× ${candidateStats[cand.id]?.ng || 0}件`}>
                    <span class="stat-badge ok">〇 {candidateStats[cand.id]?.ok || 0}</span>
                    <span class="stat-badge maybe">△ {candidateStats[cand.id]?.maybe || 0}</span>
                    <span class="stat-badge ng">× {candidateStats[cand.id]?.ng || 0}</span>
                  </div>
                </td>
              </tr>
            {/each}

            <!-- Comment row -->
            <tr class="comment-row">
              <th scope="row" class="sticky-col comment-label-cell">コメント</th>
              {#each event.responses || [] as resp}
                <td class="comment-cell" class:cell-my-response={isMyResponse(resp.id)}>
                  {#if resp.comment}
                    <div class="comment-tooltip-trigger" title={resp.comment}>
                      <span class="material-symbols-rounded" aria-hidden="true">chat_bubble</span>
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
      <div class="input-form-card glass-panel" use:reveal>
        <h3 class="section-subtitle">
          {isEditing ? 'あなたの回答を編集する' : '日程を回答する'}
        </h3>
        
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
            <span class="form-label" id="picker-label">各日程への都合</span>
            <div class="date-picker-list" role="group" aria-labelledby="picker-label">
              {#each event.candidates || [] as cand}
                <div class="picker-row">
                  <span class="picker-date">{formatDateTime(cand.event_date, cand.start_time, cand.end_time)}</span>
                  <div class="btn-group-status" role="group" aria-label={`${formatDateTime(cand.event_date, cand.start_time, cand.end_time)} の回答`}>
                    <button 
                      type="button" 
                      class="btn btn-secondary btn-status-choice active-ok"
                      class:active={myAnswers[cand.id] === 'ok'}
                      aria-pressed={myAnswers[cand.id] === 'ok'}
                      onclick={() => selectAnswer(cand.id, 'ok')}
                    >
                      〇 可
                    </button>
                    <button 
                      type="button" 
                      class="btn btn-secondary btn-status-choice active-maybe"
                      class:active={myAnswers[cand.id] === 'maybe'}
                      aria-pressed={myAnswers[cand.id] === 'maybe'}
                      onclick={() => selectAnswer(cand.id, 'maybe')}
                    >
                      △ 条件付
                    </button>
                    <button 
                      type="button" 
                      class="btn btn-secondary btn-status-choice active-ng"
                      class:active={myAnswers[cand.id] === 'ng'}
                      aria-pressed={myAnswers[cand.id] === 'ng'}
                      onclick={() => selectAnswer(cand.id, 'ng')}
                    >
                      × 不可
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <div class="form-actions-row">
            <button type="submit" class="btn btn-primary btn-lg submit-resp-btn" disabled={submitting}>
              <span class="material-symbols-rounded" aria-hidden="true">save</span>
              {submitting ? '送信中...' : isEditing ? '回答を更新する' : 'この内容で回答を登録する'}
            </button>
            {#if isEditing}
              <button type="button" class="btn btn-secondary btn-lg cancel-edit-btn" onclick={cancelEdit}>
                キャンセル
              </button>
            {/if}
          </div>
        </form>
      </div>
    {/if}

    <!-- Mobile Accordion Helper & Comments -->
    {#if event.responses && event.responses.length > 0}
      <div class="event-summary-accordion-section glass-panel" use:reveal>
        <h3 class="section-subtitle">サマリー ＆ コメント一覧</h3>
        <Accordion>
          <AccordionItem title="💬 参加者コメント一覧" icon="chat" badge={`${event.responses.filter(r => r.comment).length}件`} open={true}>
            <div class="comments-accordion-list">
              {#each event.responses.filter(r => r.comment) as resp}
                <div class="comment-item-card">
                  <div class="comment-item-header">
                    <strong>{resp.respondent_name}</strong>
                    <span class="comment-time">{new Date(resp.created_at).toLocaleDateString()}</span>
                  </div>
                  <p class="comment-body">{resp.comment}</p>
                </div>
              {:else}
                <p class="no-comment-text">まだコメントはありません</p>
              {/each}
            </div>
          </AccordionItem>

          <AccordionItem title="📊 候補日スコア集計ランキング" icon="bar_chart" open={false}>
            <div class="ranking-accordion-list">
              {#each (event.candidates || []).slice().sort((a, b) => (candidateStats[b.id]?.score || 0) - (candidateStats[a.id]?.score || 0)) as cand, idx}
                <div class="ranking-item-row" class:top-rank={idx === 0}>
                  <div class="rank-badge">{idx + 1}位</div>
                  <div class="rank-datetime">
                    {formatDateTime(cand.event_date, cand.start_time, cand.end_time)}
                  </div>
                  <div class="rank-stats">
                    <span class="stat-badge ok">〇 {candidateStats[cand.id]?.ok || 0}</span>
                    <span class="stat-badge maybe">△ {candidateStats[cand.id]?.maybe || 0}</span>
                    <span class="stat-badge ng">× {candidateStats[cand.id]?.ng || 0}</span>
                  </div>
                </div>
              {/each}
            </div>
          </AccordionItem>
        </Accordion>
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

  @media (max-width: 640px) {
    .event-page-container {
      gap: 1.25rem;
      padding: 1rem 0;
    }
  }

  /* ===========================
     Header Card
     =========================== */
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

  @media (max-width: 640px) {
    .event-title {
      font-size: 1.5rem;
    }
  }

  .event-desc-text {
    color: var(--text-secondary);
    font-size: 1.05rem;
    white-space: pre-wrap;
  }

  @media (max-width: 640px) {
    .event-desc-text {
      font-size: 0.9rem;
    }
  }

  /* ===========================
     Confirmed Box
     =========================== */
  .confirmed-box {
    display: flex;
    align-items: center;
    gap: 1.25rem;
    background: hsla(150, 84%, 37%, 0.1);
    border-color: var(--color-ok);
    padding: 1.5rem 2rem;
  }

  @media (max-width: 640px) {
    .confirmed-box {
      padding: 1rem 1.25rem;
      gap: 0.75rem;
    }
  }

  .confirmed-icon {
    font-size: 3rem;
    color: var(--color-ok);
  }

  @media (max-width: 640px) {
    .confirmed-icon {
      font-size: 2rem;
    }
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

  @media (max-width: 640px) {
    .confirmed-time {
      font-size: 1.2rem;
    }
  }

  /* ===========================
     Share Box
     =========================== */
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
    background: #FAF8F5;
    /* Prevent URL truncation on mobile */
    min-width: 0;
  }

  @media (max-width: 480px) {
    .copy-input-group {
      flex-direction: column;
    }

    .copy-input-group input {
      font-size: 0.8rem;
    }
  }

  /* ===========================
     Table
     =========================== */
  .table-wrapper {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-glass);
    /* Hint to user that table is scrollable */
  }

  @media (max-width: 640px) {
    .table-wrapper {
      margin: 0 -0.25rem; /* bleed slightly to show scroll hint */
    }
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

  @media (max-width: 640px) {
    .answers-table th, .answers-table td {
      padding: 0.65rem 0.75rem;
      font-size: 0.85rem;
    }
  }

  .answers-table th {
    background: var(--bg-secondary);
    color: var(--text-primary);
    font-family: var(--font-sans);
    font-weight: 600;
    font-size: 0.85rem;
    white-space: nowrap;
    letter-spacing: 0.02em;
  }

  @media (max-width: 640px) {
    .answers-table th {
      font-size: 0.75rem;
    }
  }

  .sticky-col {
    position: sticky;
    left: 0;
    background: var(--bg-secondary);
    z-index: 10;
    min-width: 120px;
    border-right: 1px solid var(--border-glass);
  }

  @media (min-width: 641px) {
    .sticky-col {
      min-width: 200px;
    }
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

  /* My Response Highlights */
  .th-my-response {
    background: rgba(138, 116, 99, 0.08) !important;
    border-left: 1px dashed var(--color-primary);
    border-right: 1px dashed var(--color-primary);
  }

  .cell-my-response {
    background: rgba(138, 116, 99, 0.03) !important;
    border-left: 1px dashed rgba(138, 116, 99, 0.2);
    border-right: 1px dashed rgba(138, 116, 99, 0.2);
  }

  .respondent-name-wrapper {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.2rem;
  }

  .my-response-badge {
    font-size: 0.65rem;
    background: var(--color-primary);
    color: #FAF8F5;
    padding: 0.15rem 0.5rem;
    border-radius: var(--radius-full);
    font-weight: 600;
  }

  /* ===========================
     Edit & Delete Actions
     =========================== */
  .header-action-group {
    display: flex;
    gap: 0.25rem;
  }

  .edit-resp-btn, .delete-resp-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    display: flex;
    padding: 0.4rem;
    border-radius: var(--radius-full);
    transition: opacity var(--transition-fast), color var(--transition-fast), background-color var(--transition-fast);
    opacity: 0.5;
    min-width: 32px;
    min-height: 32px;
    align-items: center;
    justify-content: center;
    -webkit-tap-highlight-color: transparent;
  }

  .edit-resp-btn.my-btn, .delete-resp-btn.my-btn {
    opacity: 1;
    color: var(--text-primary);
  }

  .edit-resp-btn:hover {
    opacity: 1;
    color: var(--color-accent);
    background: rgba(94, 111, 98, 0.08);
  }

  .delete-resp-btn:hover {
    opacity: 1;
    color: var(--color-ng);
    background: rgba(194, 134, 127, 0.08);
  }

  .edit-resp-btn .material-symbols-rounded,
  .delete-resp-btn .material-symbols-rounded {
    font-size: 0.95rem;
  }

  /* ===========================
     Status Cells
     =========================== */
  .status-cell {
    text-align: center;
  }

  .status-indicator {
    font-size: 1.3rem;
    font-weight: 800;
    display: inline-block;
  }

  @media (max-width: 640px) {
    .status-indicator {
      font-size: 1.1rem;
    }
  }

  .status-ok { color: var(--color-ok); }
  .status-maybe { color: var(--color-maybe); }
  .status-ng { color: var(--color-ng); }

  .stats-col-header {
    min-width: 100px;
    text-align: center;
  }

  @media (min-width: 641px) {
    .stats-col-header {
      min-width: 140px;
    }
  }

  .stats-cell {
    text-align: center;
  }

  .stats-summary-box {
    display: inline-flex;
    gap: 0.4rem;
    flex-wrap: wrap;
    justify-content: center;
  }

  .stat-badge {
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.2rem 0.5rem;
    border-radius: var(--radius-sm);
  }

  .stat-badge.ok { background: rgba(94, 111, 98, 0.08); color: var(--color-ok); }
  .stat-badge.maybe { background: rgba(208, 169, 126, 0.08); color: var(--color-maybe); }
  .stat-badge.ng { background: rgba(194, 134, 127, 0.08); color: var(--color-ng); }

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
    max-width: 80px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: inline-block;
  }

  @media (min-width: 641px) {
    .comment-preview {
      max-width: 120px;
    }
  }

  .no-comment {
    color: var(--text-muted);
  }

  /* ===========================
     Response Picker (date voting)
     =========================== */
  .response-dates-picker {
    margin: 2rem 0;
  }

  @media (max-width: 640px) {
    .response-dates-picker {
      margin: 1.25rem 0;
    }
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
    background: #FAF8F5;
    border: 1px solid var(--border-glass);
    padding: 1rem 1.5rem;
    border-radius: var(--radius-sm);
  }

  @media (max-width: 600px) {
    .picker-row {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.75rem;
      padding: 1rem;
    }
  }

  .picker-date {
    font-weight: 600;
  }

  @media (max-width: 640px) {
    .picker-date {
      font-size: 0.9rem;
    }
  }

  .btn-group-status {
    display: flex;
    gap: 0.5rem;
  }

  .btn-status-choice {
    font-size: 0.85rem;
    padding: 0.5rem 1rem;
    min-height: 40px;
  }

  @media (max-width: 380px) {
    /* Very small screens: smaller status buttons */
    .btn-status-choice {
      padding: 0.5rem 0.75rem;
      font-size: 0.8rem;
    }
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

  /* ===========================
     Form Actions
     =========================== */
  .form-actions-row {
    display: flex;
    gap: 1rem;
    width: 100%;
  }

  .submit-resp-btn {
    flex: 1;
  }

  .cancel-edit-btn {
    flex: 0.3;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1.5fr;
    gap: 1.5rem;
  }

  @media (max-width: 600px) {
    .form-grid {
      grid-template-columns: 1fr;
      gap: 1rem;
    }
    .form-actions-row {
      flex-direction: column;
    }
    .cancel-edit-btn {
      flex: none;
    }
  }

  /* ===========================
     Error State
     =========================== */
  .error-panel-large {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 5rem 2rem;
    text-align: center;
    gap: 1rem;
  }

  @media (max-width: 640px) {
    .error-panel-large {
      padding: 3rem 1rem;
    }
  }

  .error-icon-lg {
    font-size: 4rem;
    color: var(--color-ng);
  }

  /* Comments and Ranking Accordion Styles */
  .comments-accordion-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 0.5rem 0;
  }

  .comment-item-card {
    background: rgba(255, 255, 255, 0.4);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 0.75rem 1rem;
  }

  .comment-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.88rem;
    margin-bottom: 0.25rem;
  }

  .comment-time {
    font-size: 0.78rem;
    color: var(--text-muted);
  }

  .comment-body {
    font-size: 0.88rem;
    color: var(--text-secondary);
    white-space: pre-wrap;
  }

  .no-comment-text {
    font-size: 0.85rem;
    color: var(--text-muted);
    font-style: italic;
  }

  .ranking-accordion-list {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 0.5rem 0;
  }

  .ranking-item-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.6rem 0.85rem;
    background: rgba(255, 255, 255, 0.35);
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-glass);
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .ranking-item-row.top-rank {
    border-color: rgba(212, 140, 56, 0.4);
    background: rgba(212, 140, 56, 0.08);
  }

  .rank-badge {
    font-weight: 700;
    font-size: 0.82rem;
    padding: 0.15rem 0.5rem;
    border-radius: var(--radius-full);
    background: var(--bg-secondary);
    color: var(--text-primary);
  }

  .top-rank .rank-badge {
    background: var(--color-accent);
    color: #fff;
  }

  .rank-datetime {
    font-weight: 600;
    font-size: 0.88rem;
    flex: 1;
  }

  .rank-stats {
    display: flex;
    gap: 0.4rem;
  }
</style>
