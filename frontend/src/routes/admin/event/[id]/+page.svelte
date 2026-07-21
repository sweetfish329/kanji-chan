<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api';
  import { Accordion, AccordionItem, AIPromptInput, AlertDialog, Popover, Select, Switch, toast, type AttachedImage, type SelectOptionItem } from '$lib';
  import dayjs from 'dayjs';
  import 'dayjs/locale/ja';

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
  let aiImages = $state<AttachedImage[]>([]);
  let isAnalyzing = $state(false);
  let aiSuggestions = $state<AISuggestionsResponse | null>(null);

  // Confirmation states
  let selectedCandidateId = $state<number | null>(null);
  let selectedCandidateIdStr = $state<string>('');
  let submitting = $state(false);
  let confirmDialogOpen = $state(false);
  let cancelConfirmDialogOpen = $state(false);
  let deleteResponseDialogOpen = $state(false);
  let responseToDelete = $state<{ id: number; name: string } | null>(null);

  $effect(() => {
    if (selectedCandidateId !== null) {
      selectedCandidateIdStr = String(selectedCandidateId);
    } else {
      selectedCandidateIdStr = '';
    }
  });

  function handleCandidateSelectChange(val: string) {
    if (val) {
      selectedCandidateId = Number(val);
    } else {
      selectedCandidateId = null;
    }
  }

  let candidateSelectOptions = $derived.by<SelectOptionItem[]>(() => {
    if (!event?.candidates) return [];
    return event.candidates.map(cand => ({
      value: String(cand.id),
      label: `${formatDateTime(cand.event_date, cand.start_time, cand.end_time)} (〇:${candidateStats[cand.id]?.ok || 0} △:${candidateStats[cand.id]?.maybe || 0} ×:${candidateStats[cand.id]?.ng || 0})`
    }));
  });

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

  let aiStep = $state(0);
  let aiTimer: ReturnType<typeof setInterval> | null = null;

  // AI に日程絞り込みを依頼
  async function runAIAnalysis() {
    isAnalyzing = true;
    aiSuggestions = null;
    aiStep = 0;

    if (aiTimer) clearInterval(aiTimer);
    aiTimer = setInterval(() => {
      aiStep = (aiStep + 1) % 3;
    }, 1000);
    
    try {
      const payload = {
        event_id: eventId,
        preferences: preferencesInput,
        images: aiImages.map(img => ({ data: img.data, mime_type: img.mime_type }))
      };
      const data = await api.post<AISuggestionsResponse>('/ai/suggest-schedule', payload);
      aiSuggestions = data;
      
      // AI推薦の1番（Rank 1）を自動的にデフォルト選択にする
      if (data.suggestions && data.suggestions.length > 0) {
        selectedCandidateId = data.suggestions[0].candidate_id;
      }
      toast.push('✨ AIによる最適日程の推薦が完了しました！');
    } catch (err: any) {
      toast.push('AI推薦に失敗しました: ' + err.message, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    } finally {
      if (aiTimer) clearInterval(aiTimer);
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
  async function handleConfirmCancelConfirmation() {
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
      toast.push('解除に失敗しました: ' + err.message);
    } finally {
      submitting = false;
    }
  }

  function openDeleteResponseDialog(id: number, name: string) {
    responseToDelete = { id, name };
    deleteResponseDialogOpen = true;
  }

  // 回答の削除
  async function handleConfirmDeleteResponse() {
    if (!responseToDelete) return;
    try {
      await api.delete(`/events/${eventId}/responses/${responseToDelete.id}`);
      toast.push('回答を削除しました');
      await loadEvent();
    } catch (err: any) {
      toast.push('回答の削除に失敗しました: ' + err.message);
    } finally {
      responseToDelete = null;
    }
  }

  function formatDateTime(dateStr: string, startStr: string, endStr: string): string {
    const d = dayjs(dateStr);
    const formattedDate = d.format('M/D(ddd)');
    return `${formattedDate} ${startStr.slice(0, 5)}〜${endStr.slice(0, 5)}`;
  }
</script>

<svelte:head>
  <title>{event ? `「${event.title}」の管理・AI提案 | 幹事ちゃん` : 'イベント管理 | 幹事ちゃん'}</title>
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
                          onclick={() => openDeleteResponseDialog(resp.id, resp.respondent_name)} 
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
            <AIPromptInput
              bind:prompt={preferencesInput}
              bind:images={aiImages}
              placeholder="e.g. Aさんは必須で参加。なるべく金曜日を優先（条件メモ・希望カレンダー画像のドラッグ＆ドロップ添付も可能）"
              submitLabel="AIに最適な日程を絞り込ませる"
              isSubmitting={isAnalyzing}
              disabled={event.status === 'confirmed'}
              onSubmit={runAIAnalysis}
            />
          {/if}

          {#if isAnalyzing}
            <div class="ai-loading-card glass-panel animate-fade-in" role="status">
              <div class="shimmer-bar"></div>
              <div class="loading-status-content">
                <div class="ai-pulse-orb"></div>
                <div class="loading-text-wrapper">
                  {#if aiStep === 0}
                    <p class="step-text animate-slide-up">📊 全参加者の〇△×回答スコアを集計中...</p>
                  {:else if aiStep === 1}
                    <p class="step-text animate-slide-up">🔍 幹事のこだわり希望条件を検証中...</p>
                  {:else}
                    <p class="step-text animate-slide-up">⭐ 最適開催日時のランキングを作成中...</p>
                  {/if}
                </div>
              </div>
            </div>
          {/if}

          <!-- AI suggestions results -->
          {#if aiSuggestions && !isAnalyzing}
            <div class="ai-results animate-fade-in">
              <hr class="divider-small" />
              <div class="results-title-row">
                <span class="material-symbols-rounded spark-glow" aria-hidden="true">auto_awesome</span>
                <h4>AI推薦候補ランキング</h4>
              </div>
              
              <div class="suggestions-cards" role="group" aria-label="AI推薦候補リスト">
                {#each aiSuggestions.suggestions as sug, index}
                  {@const candidate = event.candidates.find(c => c.id === sug.candidate_id)}
                  {#if candidate}
                    <button 
                      type="button"
                      class="sug-card glass-panel animated-sug-card" 
                      class:active={selectedCandidateId === sug.candidate_id}
                      class:rank-1={sug.rank === 1}
                      style="animation-delay: {index * 0.1}s"
                      aria-pressed={selectedCandidateId === sug.candidate_id}
                      onclick={() => {
                        if (event?.status !== 'confirmed') {
                          selectedCandidateId = sug.candidate_id;
                        }
                      }}
                    >
                      <div class="sug-card-header">
                        <span class="rank-badge" class:gold={sug.rank === 1}>
                          {sug.rank === 1 ? '👑 第 1 推薦' : `第 ${sug.rank} 候補`}
                        </span>
                        <span class="score-badge">スコア: {sug.score}pt</span>
                      </div>
                      <h5 class="font-mono">{formatDateTime(candidate.event_date, candidate.start_time, candidate.end_time)}</h5>
                      <p class="sug-reason">{sug.reason}</p>
                      {#if selectedCandidateId === sug.candidate_id}
                        <div class="selected-check-badge bounce-in">
                          <span class="material-symbols-rounded" aria-hidden="true">check_circle</span>
                          選択中
                        </div>
                      {/if}
                    </button>
                  {/if}
                {/each}
              </div>

              <Accordion class="ai-overall-accordion">
                <AccordionItem title="AIによる全体分析レビュー" icon="analytics" open={true}>
                  <p class="overall-text">{aiSuggestions.overall_analysis}</p>
                </AccordionItem>
              </Accordion>
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
              onclick={() => cancelConfirmDialogOpen = true}
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
              <Select
                bind:value={selectedCandidateIdStr}
                options={candidateSelectOptions}
                placeholder="-- 確定する日程を選択してください --"
                onValueChange={handleCandidateSelectChange}
              />
            </div>

            <button 
              type="button" 
              class="btn btn-primary w-full btn-lg" 
              onclick={() => confirmDialogOpen = true}
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

  <!-- Bits UI AlertDialog for Schedule Confirmation -->
  <AlertDialog
    bind:open={confirmDialogOpen}
    title="開催日程の確定確認"
    description="選択した開催日時で確定し、回答者に結果を公開します。確定してよろしいですか？"
    confirmText="確定を保存する"
    cancelText="キャンセル"
    onConfirm={confirmSchedule}
  >
    {#if selectedCandidateId && event}
      {@const selectedCand = event.candidates.find(c => c.id === selectedCandidateId)}
      {#if selectedCand}
        <div class="confirm-cand-preview font-mono">
          📅 <strong>{formatDateTime(selectedCand.event_date, selectedCand.start_time, selectedCand.end_time)}</strong>
        </div>
      {/if}
    {/if}
  </AlertDialog>

  <!-- Bits UI AlertDialog for Cancel Confirmation -->
  <AlertDialog
    bind:open={cancelConfirmDialogOpen}
    title="日程確定の解除確認"
    description="日程確定を解除し、再度調整中に戻しますか？"
    confirmText="解除する"
    cancelText="キャンセル"
    danger={true}
    onConfirm={handleConfirmCancelConfirmation}
  />

  <!-- Bits UI AlertDialog for Response Deletion -->
  <AlertDialog
    bind:open={deleteResponseDialogOpen}
    title="回答の削除確認"
    description={responseToDelete ? `「${responseToDelete.name}」さんの回答を削除しますか？` : ''}
    confirmText="削除する"
    cancelText="キャンセル"
    danger={true}
    onConfirm={handleConfirmDeleteResponse}
  />
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
    background: var(--bg-secondary);
    color: var(--text-primary);
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
    background: var(--bg-secondary);
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
    background: #FAF8F5;
    padding: 1rem;
    cursor: pointer;
    transition: border-color var(--transition-fast), background-color var(--transition-fast), box-shadow var(--transition-fast);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    text-align: left;
    width: 100%;
    font-family: inherit;
  }

  .sug-card:hover {
    border-color: var(--color-primary);
    background: #F4EFEA;
  }

  .sug-card.active {
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px var(--border-focus);
    background: rgba(42, 64, 50, 0.05);
  }

  .sug-card.rank-1 {
    border-color: rgba(212, 140, 56, 0.5);
    background: linear-gradient(135deg, #FAF8F5 0%, rgba(212, 140, 56, 0.06) 100%);
  }

  .sug-card.rank-1:hover {
    border-color: var(--color-accent);
    background: linear-gradient(135deg, #F4EFEA 0%, rgba(212, 140, 56, 0.12) 100%);
  }

  .sug-card.rank-1.active {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 3px rgba(212, 140, 56, 0.3);
    background: linear-gradient(135deg, #FAF8F5 0%, rgba(212, 140, 56, 0.14) 100%);
  }

  .sug-card-header {
    display: flex;
    justify-content: space-between;
    font-size: 0.75rem;
    font-weight: 700;
    margin-bottom: 0.5rem;
  }

  .rank-badge {
    color: var(--color-accent);
  }

  .sug-card.rank-1 .rank-badge {
    color: var(--color-accent);
    font-weight: 700;
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
    background: var(--bg-secondary);
    border-radius: var(--radius-sm);
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

  /* AI Animation Styles */
  .ai-sparkle-icon.spin {
    animation: rotateSparkle 1.5s infinite linear;
  }

  @keyframes rotateSparkle {
    0% { transform: rotate(0deg) scale(1); }
    50% { transform: rotate(180deg) scale(1.2); }
    100% { transform: rotate(360deg) scale(1); }
  }

  .results-title-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
    color: var(--color-primary);
  }

  .spark-glow {
    color: var(--color-accent);
    animation: pulseGlow 2s infinite ease-in-out;
  }

  @keyframes pulseGlow {
    0%, 100% { transform: scale(1); filter: drop-shadow(0 0 2px var(--color-accent)); }
    50% { transform: scale(1.2); filter: drop-shadow(0 0 8px var(--color-accent)); }
  }

  .animated-sug-card {
    animation: cardSlideIn 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
    opacity: 0;
    transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.2s ease, border-color 0.2s ease;
  }

  @keyframes cardSlideIn {
    from { opacity: 0; transform: translateY(12px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .animated-sug-card:hover {
    transform: translateY(-3px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  }

  .rank-badge.gold {
    background: linear-gradient(135deg, #d4af37, #f3e5ab);
    color: #3a2e05;
    font-weight: 700;
    box-shadow: 0 2px 8px rgba(212, 175, 55, 0.3);
  }

  .sug-card.rank-1 {
    border: 2px solid #d4af37;
    background: linear-gradient(135deg, #FAF8F5, rgba(212, 175, 55, 0.08));
  }

  .selected-check-badge {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    margin-top: 0.5rem;
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--color-ok);
  }

  .bounce-in {
    animation: bounceIn 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards;
  }

  @keyframes bounceIn {
    0% { transform: scale(0.7); opacity: 0; }
    100% { transform: scale(1); opacity: 1; }
  }

  .overall-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    color: var(--color-primary);
  }

  .ai-loading-card {
    position: relative;
    margin-top: 1.5rem;
    padding: 1.5rem;
    border-radius: var(--radius-md);
    background: linear-gradient(135deg, rgba(94, 111, 98, 0.08), rgba(208, 169, 126, 0.08));
    border: 1px solid var(--color-accent);
    overflow: hidden;
  }

  .shimmer-bar {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 4px;
    background: linear-gradient(90deg, transparent, var(--color-accent), transparent);
    animation: shimmerSlide 2s infinite;
  }

  @keyframes shimmerSlide {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(100%); }
  }

  .loading-status-content {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .ai-pulse-orb {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--color-accent);
    box-shadow: 0 0 0 0 rgba(208, 169, 126, 0.7);
    animation: pulseOrb 1.5s infinite cubic-bezier(0.66, 0, 0, 1);
  }

  @keyframes pulseOrb {
    to {
      box-shadow: 0 0 0 16px rgba(208, 169, 126, 0);
    }
  }

  .loading-text-wrapper {
    flex: 1;
    overflow: hidden;
  }

  .step-text {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .animate-slide-up {
    animation: slideUpFade 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }

  @keyframes slideUpFade {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
