<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api';
  import dayjs from 'dayjs';
  import 'dayjs/locale/ja';
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

  interface AISuggestion {
    candidate_id: number;
    rank: number;
    score: number;
    reason: string;
  }

  interface AISuggestionsResponse {
    suggestions: AISuggestion[];
    overall_analysis: string;
  }

  const eventId = $page.params.id;

  // Svelte 5 Runes for Reactivity
  let event = $state<Event | null>(null);
  let loading = $state(true);
  let errorMsg = $state('');

  // AI analysis states
  let preferencesInput = $state('');
  let isAnalyzing = $state(false);
  let aiSuggestions = $state<AISuggestionsResponse | null>(null);

  // Confirmation states
  let selectedCandidateId = $state<number | null>(null);
  let submitting = $state(false);

  // Load Event Details (Admin only path, Auth checked on API side)
  async function loadEvent() {
    try {
      const data = await api.get<Event>(`/events/${eventId}`);
      event = data;
      if (data.confirmed_candidate_id) {
        selectedCandidateId = data.confirmed_candidate_id;
      }
    } catch (err: any) {
      errorMsg = err.message || 'イベントのロードに失敗しました。';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadEvent();
  });

  // Derived: 候補日ごとの〇△×の集計 (Svelte 5 Runes)
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
          if (ans.answer_status === 'ok') stats[ans.candidate_id].score += 2;
          if (ans.answer_status === 'maybe') stats[ans.candidate_id].score += 1;
        }
      });
    });

    return stats;
  });

  // AI に日程絞り込みを依頼
  async function runAIAnalysis() {
    isAnalyzing = true;
    aiSuggestions = null;
    
    try {
      const data = await api.post<AISuggestionsResponse>('/ai/suggest-schedule', {
        event_id: eventId,
        preferences: preferencesInput
      });
      aiSuggestions = data;
      
      // AI推薦の1番（Rank 1）を自動的にデフォルト選択にする
      if (data.suggestions && data.suggestions.length > 0) {
        selectedCandidateId = data.suggestions[0].candidate_id;
      }
      toast.push('AIの予定推薦が完了しました！');
    } catch (err: any) {
      toast.push('AI推薦に失敗しました: ' + err.message, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    } finally {
      isAnalyzing = false;
    }
  }

  // 日程の確定
  async function confirmSchedule() {
    if (!selectedCandidateId) {
      toast.push('確定する日程を選択してください', {
        theme: {
          '--toastBackground': 'var(--color-maybe)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
      return;
    }

    submitting = true;
    try {
      const updated = await api.put<Event>(`/events/${eventId}`, {
        status: 'confirmed',
        confirmed_candidate_id: selectedCandidateId
      });
      event = updated;
      toast.push('日程を確定しました！参加者に共有されます。');
    } catch (err: any) {
      toast.push('日程の確定に失敗しました: ' + err.message, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    } finally {
      submitting = false;
    }
  }

  // 確定取り消し
  async function cancelConfirmation() {
    if (!confirm('日程確定を解除し、再度調整中に戻しますか？')) {
      return;
    }

    submitting = true;
    try {
      const updated = await api.put<Event>(`/events/${eventId}`, {
        status: 'scheduling',
        confirmed_candidate_id: null
      });
      event = updated;
      selectedCandidateId = null;
      toast.push('調整中に戻しました。');
    } catch (err: any) {
      toast.push('解除に失敗しました: ' + err.message, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    } finally {
      submitting = false;
    }
  }

  // 回答の削除
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

<!-- noindex: 管理画面は検索エンジンに表示しない -->
<svelte:head>
  <title>イベント管理 | 幹事ちゃん</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="container admin-event-container animate-fade-in">
  <div class="back-link-area">
    <a href="/admin" class="btn btn-secondary btn-sm">
      <span class="material-symbols-rounded" aria-hidden="true">arrow_back</span>
      ダッシュボードに戻る
    </a>
  </div>

  {#if loading}
    <div class="glass-panel loading-panel" role="status" aria-label="読み込み中">
      <div class="spinner"></div>
      <p>イベント管理データをロード中...</p>
    </div>
  {:else if errorMsg}
    <div class="glass-panel error-panel-large" role="alert">
      <span class="material-symbols-rounded error-icon-lg" aria-hidden="true">warning</span>
      <h2>エラーが発生しました</h2>
      <p>{errorMsg}</p>
      <a href="/admin" class="btn btn-primary">管理画面に戻る</a>
    </div>
  {:else if event}
    <div class="event-details-layout">
      <!-- Left side: Information and Answers Table -->
      <div class="main-info-column">
        <!-- Event Card -->
        <div class="glass-panel info-card">
          <div class="event-meta-top">
            <span class="status-badge" class:confirmed={event.status === 'confirmed'}>
              {event.status === 'confirmed' ? '日程確定済み' : '調整中'}
            </span>
          </div>
          <h2>{event.title}</h2>
          <p class="desc-text">{event.description || '説明はありません。'}</p>
        </div>

        <!-- Answers Status Table -->
        <div class="glass-panel table-card">
          <h3>みんなの回答状況</h3>
          <div class="table-wrapper-admin">
            <table class="admin-table">
              <thead>
                <tr>
                  <th scope="col">日程</th>
                  {#each event.responses as resp}
                    <th scope="col">
                      <div class="th-resp">
                        <span>{resp.respondent_name}</span>
                        <button 
                          onclick={() => deleteResponse(resp.id, resp.respondent_name)} 
                          class="btn-del-resp" 
                          title="回答削除"
                          aria-label={`「${resp.respondent_name}」の回答を削除`}
                        >
                          <span class="material-symbols-rounded" aria-hidden="true">close</span>
                        </button>
                      </div>
                    </th>
                  {/each}
                  <th scope="col">〇 △ ×</th>
                </tr>
              </thead>
              <tbody>
                {#each event.candidates as cand}
                  <tr class:confirmed-row={event.status === 'confirmed' && event.confirmed_candidate_id === cand.id}>
                    <th scope="row" class="td-datetime font-mono">
                      {formatDateTime(cand.event_date, cand.start_time, cand.end_time)}
                      {#if event.status === 'confirmed' && event.confirmed_candidate_id === cand.id}
                        <span class="conf-label">確定</span>
                      {/if}
                    </th>
                    {#each event.responses as resp}
                      {@const ans = resp.answers.find(a => a.candidate_id === cand.id)}
                      <td class="td-status">
                        {#if ans}
                          <span 
                            class={`status-sym ${ans.answer_status}`}
                            aria-label={ans.answer_status === 'ok' ? '可' : ans.answer_status === 'maybe' ? '条件付き可' : '不可'}
                          >
                            {ans.answer_status === 'ok' ? '〇' : ans.answer_status === 'maybe' ? '△' : '×'}
                          </span>
                        {:else}
                          <span aria-label="未回答">-</span>
                        {/if}
                      </td>
                    {/each}
                    <td>
                      <div class="stats-row" aria-label={`〇 ${candidateStats[cand.id]?.ok || 0}件、△ ${candidateStats[cand.id]?.maybe || 0}件、× ${candidateStats[cand.id]?.ng || 0}件`}>
                        <span class="stat ok">〇 {candidateStats[cand.id]?.ok || 0}</span>
                        <span class="stat maybe">△ {candidateStats[cand.id]?.maybe || 0}</span>
                        <span class="stat ng">× {candidateStats[cand.id]?.ng || 0}</span>
                      </div>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Right side: AI helper and final confirmation -->
      <div class="ai-helper-column">
        <!-- AI suggest request form -->
        <div class="ai-aurora-card ai-panel animate-fade-in">
          <div class="ai-panel-header">
            <span class="material-symbols-rounded ai-icon" aria-hidden="true">psychology</span>
            <h3>AI日程選定アシスタント</h3>
          </div>
          <p class="ai-intro">回答データをもとに、AIが最適な開催日程を決定します。優先したい希望があれば入力してください。</p>

          {#if event.responses.length === 0}
            <div class="ai-empty" role="status">
              <span class="material-symbols-rounded" aria-hidden="true">info</span>
              <p>回答者がまだ集まっていません。日程提案には回答データが必要です。</p>
            </div>
          {:else}
            <div class="form-group">
              <label for="prefs">幹事のこだわり・追加条件</label>
              <textarea 
                id="prefs" 
                rows="3" 
                placeholder="e.g. Aさんは必須参加でお願いします。なるべく週末が良いです。多少人数が減ってもいいので全員が〇の日を優先して。"
                bind:value={preferencesInput}
                disabled={event.status === 'confirmed'}
              ></textarea>
            </div>

            {#if event.status !== 'confirmed'}
              <button 
                type="button" 
                class="btn btn-primary w-full" 
                onclick={runAIAnalysis}
                disabled={isAnalyzing}
              >
                <span class="material-symbols-rounded" aria-hidden="true">auto_awesome</span>
                {isAnalyzing ? 'AIが提案作成中...' : 'AIに最適な日程を絞り込ませる'}
              </button>
            {/if}
          {/if}

          <!-- AI suggestions results -->
          {#if aiSuggestions}
            <div class="ai-results animate-fade-in">
              <hr class="divider-small" />
              <h4>AI推薦候補ランキング</h4>
              
              <div class="suggestions-cards" role="group" aria-label="AI推薦候補リスト">
                {#each aiSuggestions.suggestions as sug}
                  {@const candidate = event.candidates.find(c => c.id === sug.candidate_id)}
                  {#if candidate}
                    <button 
                      type="button"
                      class="sug-card glass-panel" 
                      class:active={selectedCandidateId === sug.candidate_id}
                      class:rank-1={sug.rank === 1}
                      aria-pressed={selectedCandidateId === sug.candidate_id}
                      onclick={() => {
                        if (event?.status !== 'confirmed') {
                          selectedCandidateId = sug.candidate_id;
                        }
                      }}
                    >
                      <div class="sug-card-header">
                        <span class="rank-badge">第 {sug.rank} 候補</span>
                        <span class="score-badge">スコア: {sug.score}</span>
                      </div>
                      <h5 class="font-mono">{formatDateTime(candidate.event_date, candidate.start_time, candidate.end_time)}</h5>
                      <p class="sug-reason">{sug.reason}</p>
                    </button>
                  {/if}
                {/each}
              </div>

              <div class="overall-box glass-panel">
                <h5>全体分析レビュー</h5>
                <p>{aiSuggestions.overall_analysis}</p>
              </div>
            </div>
          {/if}
        </div>

        <!-- Schedule final confirmation -->
        <div class="glass-panel confirmation-panel">
          <h3>開催日程の確定</h3>
          
          {#if event.status === 'confirmed' && event.confirmed_candidate}
            <div class="finalized-box">
              <span class="material-symbols-rounded ok-check" aria-hidden="true">check_circle</span>
              <div>
                <p>確定済み日程</p>
                <h4 class="font-mono">{formatDateTime(event.confirmed_candidate.event_date, event.confirmed_candidate.start_time, event.confirmed_candidate.end_time)}</h4>
              </div>
            </div>
            <button 
              type="button" 
              class="btn btn-secondary w-full" 
              onclick={cancelConfirmation}
              disabled={submitting}
            >
              確定をキャンセルして再調整
            </button>
          {:else}
            <p class="confirm-intro">
              確定させたい候補日を選択し、ボタンを押してください。確定すると日程が参加者に公開されます。
            </p>

            <div class="form-group select-confirm-group">
              <label for="final-select">日程を選択する</label>
              <select id="final-select" bind:value={selectedCandidateId}>
                <option value={null} disabled>-- 日程を選択してください --</option>
                {#each event.candidates as cand}
                  <option value={cand.id}>
                    {formatDateTime(cand.event_date, cand.start_time, cand.end_time)} 
                    (〇:{candidateStats[cand.id]?.ok || 0} △:{candidateStats[cand.id]?.maybe || 0} ×:{candidateStats[cand.id]?.ng || 0})
                  </option>
                {/each}
              </select>
            </div>

            <button 
              type="button" 
              class="btn btn-primary w-full btn-lg" 
              onclick={confirmSchedule}
              disabled={submitting || !selectedCandidateId}
            >
              <span class="material-symbols-rounded" aria-hidden="true">celebration</span>
              この日程で決定・確定する
            </button>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .admin-event-container {
    padding: 1.5rem 0;
  }

  .back-link-area {
    margin-bottom: 1.5rem;
  }

  .event-details-layout {
    display: grid;
    grid-template-columns: 1.1fr 0.9fr;
    gap: 2rem;
  }

  @media (max-width: 900px) {
    .event-details-layout {
      grid-template-columns: 1fr;
    }
  }

  .main-info-column, .ai-helper-column {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  /* Info Card */
  .info-card {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .desc-text {
    color: var(--text-secondary);
    font-size: 1rem;
    white-space: pre-wrap;
  }

  /* Table styles */
  .table-wrapper-admin {
    overflow-x: auto;
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
  }

  .admin-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  .admin-table th, .admin-table td {
    padding: 0.85rem;
    border-bottom: 1px solid var(--border-glass);
  }

  .admin-table th {
    background: hsla(223, 40%, 10%, 0.8);
    font-weight: 600;
  }

  .th-resp {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    white-space: nowrap;
  }

  .btn-del-resp {
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    border-radius: 50%;
    padding: 0.1rem;
    display: flex;
  }

  .btn-del-resp:hover {
    color: var(--color-ng);
    background: hsla(350, 89%, 60%, 0.1);
  }

  .btn-del-resp .material-symbols-rounded {
    font-size: 0.85rem;
  }

  .td-datetime {
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .conf-label {
    background: var(--color-ok);
    color: #fff;
    font-size: 0.65rem;
    font-weight: 800;
    padding: 0.15rem 0.35rem;
    border-radius: var(--radius-sm);
  }

  .confirmed-row {
    background: hsla(150, 84%, 37%, 0.04);
  }

  .td-status {
    text-align: center;
  }

  .status-sym {
    font-weight: 700;
    font-size: 1.1rem;
  }

  .status-sym.ok { color: var(--color-ok); }
  .status-sym.maybe { color: var(--color-maybe); }
  .status-sym.ng { color: var(--color-ng); }

  .stats-row {
    display: flex;
    gap: 0.4rem;
    font-size: 0.75rem;
  }

  .stat {
    padding: 0.15rem 0.4rem;
    border-radius: var(--radius-sm);
    font-weight: 600;
  }

  .stat.ok { background: hsla(150, 84%, 37%, 0.12); color: var(--color-ok); }
  .stat.maybe { background: hsla(38, 92%, 50%, 0.12); color: var(--color-maybe); }
  .stat.ng { background: hsla(350, 89%, 60%, 0.12); color: var(--color-ng); }

  /* AI Panel */
  .ai-panel-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .ai-icon {
    font-size: 2.2rem;
    color: var(--color-primary);
  }

  .ai-intro {
    color: var(--text-secondary);
    font-size: 0.85rem;
    margin-bottom: 1.5rem;
  }

  .ai-empty {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    background: hsla(223, 40%, 16%, 0.5);
    border: 1px solid var(--border-glass);
    padding: 1rem;
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-size: 0.85rem;
  }

  .w-full {
    width: 100%;
  }

  .divider-small {
    border: 0;
    border-top: 1px solid var(--border-glass);
    margin: 1.5rem 0;
  }

  .ai-results h4 {
    font-size: 1.05rem;
    margin-bottom: 1rem;
  }

  .suggestions-cards {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .sug-card {
    background: hsla(223, 40%, 10%, 0.6);
    padding: 1rem;
    cursor: pointer;
    transition: all var(--transition-fast);
    border: 1px solid var(--border-glass);
    text-align: left;
    width: 100%;
    font-family: inherit;
  }

  .sug-card:hover {
    border-color: var(--color-primary);
    background: hsla(223, 40%, 14%, 0.8);
  }

  .sug-card.active {
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px var(--border-focus);
    background: hsla(263, 90%, 65%, 0.05);
  }

  .sug-card.rank-1 {
    border-color: hsla(322, 85%, 60%, 0.3);
    background: hsla(322, 85%, 60%, 0.02);
  }

  .sug-card.rank-1:hover {
    border-color: var(--color-secondary);
    background: hsla(322, 85%, 60%, 0.06);
  }

  .sug-card.rank-1.active {
    border-color: var(--color-secondary);
    box-shadow: 0 0 0 3px hsla(322, 85%, 60%, 0.3);
    background: hsla(322, 85%, 60%, 0.08);
  }

  .sug-card-header {
    display: flex;
    justify-content: space-between;
    font-size: 0.75rem;
    font-weight: 700;
    margin-bottom: 0.5rem;
  }

  .rank-badge {
    color: var(--color-secondary);
  }

  .sug-card.rank-1 .rank-badge {
    color: var(--color-secondary);
    background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .score-badge {
    color: var(--text-muted);
  }

  .sug-card h5 {
    font-size: 1.05rem;
    margin-bottom: 0.5rem;
    color: var(--text-primary);
  }

  .sug-reason {
    font-size: 0.8rem;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .overall-box {
    background: hsla(223, 40%, 8%, 0.7);
    padding: 1.25rem;
    font-size: 0.85rem;
  }

  .overall-box h5 {
    margin-bottom: 0.5rem;
    color: var(--color-primary);
  }

  .overall-box p {
    color: var(--text-secondary);
    line-height: 1.5;
  }

  /* Confirmation Panel */
  .confirm-intro {
    color: var(--text-secondary);
    font-size: 0.85rem;
    margin-bottom: 1.5rem;
  }

  .finalized-box {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: hsla(150, 84%, 37%, 0.1);
    border: 1px solid var(--color-ok);
    padding: 1.25rem;
    border-radius: var(--radius-sm);
    margin-bottom: 1.5rem;
  }

  .ok-check {
    color: var(--color-ok);
    font-size: 2.2rem;
  }

  .finalized-box p {
    font-size: 0.75rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    font-weight: 600;
  }

  .finalized-box h4 {
    font-size: 1.2rem;
    color: var(--color-ok);
  }

  .select-confirm-group {
    margin-bottom: 1.5rem;
  }
</style>
