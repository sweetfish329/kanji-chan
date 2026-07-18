<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api';
  import { toast } from '@zerodevx/svelte-toast';

  interface User {
    id: number;
    name: string;
    email: string;
    gemini_api_key?: string;
  }

  interface Event {
    id: string;
    title: string;
    description: string;
    status: string;
    created_at: string;
  }

  interface CandidateInput {
    event_date: string;
    start_time: string;
    end_time: string;
  }

  // Svelte 5 Runes for Reactivity
  let user = $state<User | null>(null);
  let events = $state<Event[]>([]);
  let activeTab = $state<'list' | 'create-ai' | 'create-manual' | 'settings'>('list');
  let loading = $state(true);

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '';

  // API Key state
  let apiKeyInput = $state('');
  let apiKeyUpdateSuccess = $state('');

  // AI Creation state
  let aiTextInput = $state('');
  let isParsing = $state(false);
  let parseError = $state('');
  let parsedTitle = $state('');
  let parsedDescription = $state('');
  let parsedCandidates = $state<CandidateInput[]>([]);

  // Manual Creation state
  let manualTitle = $state('');
  let manualDescription = $state('');
  let manualCandidates = $state<CandidateInput[]>([{ event_date: '', start_time: '19:00', end_time: '21:00' }]);

  // Load Admin Dashboard Data
  onMount(async () => {
    try {
      const userData = await api.get<User>('/auth/me');
      user = userData;
      if (userData.gemini_api_key) {
        apiKeyInput = userData.gemini_api_key;
      }
      
      const eventsData = await api.get<Event[]>('/events');
      events = eventsData;
    } catch (err) {
      console.error('Failed to load dashboard data:', err);
      // 未ログインの場合はログインへ誘導
      window.location.href = `${apiBaseUrl}/api/auth/login`;
    } finally {
      loading = false;
    }
  });

  async function updateApiKey(e: SubmitEvent) {
    e.preventDefault();
    try {
      await api.post('/auth/apikey', { gemini_api_key: apiKeyInput });
      apiKeyUpdateSuccess = 'APIキーを正常に更新しました！';
      toast.push('APIキーを設定しました！');
      setTimeout(() => apiKeyUpdateSuccess = '', 3000);
    } catch (err) {
      const msg = err instanceof Error ? err.message : '予期せぬエラーが発生しました';
      toast.push('APIキーの更新に失敗しました: ' + msg, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    }
  }

  // AIで自然文をパース
  async function parseNaturalLanguage(e: SubmitEvent) {
    e.preventDefault();
    if (!aiTextInput.trim()) return;
    
    isParsing = true;
    parseError = '';
    
    try {
      const result = await api.post<{ title: string; description: string; candidates: CandidateInput[] }>(
        '/ai/parse-event', 
        { text: aiTextInput }
      );
      
      parsedTitle = result.title;
      parsedDescription = result.description;
      parsedCandidates = result.candidates;
      toast.push('自然文の解析が完了しました！');
    } catch (err) {
      const msg = err instanceof Error ? err.message : '解析に失敗しました。';
      parseError = msg + ' APIキーが設定されているか確認してください。';
      toast.push('解析に失敗しました', {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    } finally {
      isParsing = false;
    }
  }

  // 候補日時行の操作
  function addCandidate(type: 'ai' | 'manual') {
    const newCand = { event_date: '', start_time: '19:00', end_time: '21:00' };
    if (type === 'ai') {
      parsedCandidates = [...parsedCandidates, newCand];
    } else {
      manualCandidates = [...manualCandidates, newCand];
    }
  }

  function removeCandidate(index: number, type: 'ai' | 'manual') {
    if (type === 'ai') {
      parsedCandidates = parsedCandidates.filter((_, i) => i !== index);
    } else {
      manualCandidates = manualCandidates.filter((_, i) => i !== index);
    }
  }

  // イベント確定・送信
  async function submitEvent(type: 'ai' | 'manual') {
    const title = type === 'ai' ? parsedTitle : manualTitle;
    const description = type === 'ai' ? parsedDescription : manualDescription;
    const candidates = type === 'ai' ? parsedCandidates : manualCandidates;

    if (!title.trim()) {
      toast.push('イベントのタイトルを入力してください', {
        theme: {
          '--toastBackground': 'var(--color-maybe)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
      return;
    }

    if (candidates.length === 0 || candidates.some(c => !c.event_date)) {
      toast.push('候補日を正しく設定してください（日付の入力は必須です）', {
        theme: {
          '--toastBackground': 'var(--color-maybe)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
      return;
    }

    try {
      const created = await api.post<Event>('/events', {
        title,
        description,
        candidates
      });

      // ダッシュボード更新し、イベント一覧タブに戻る
      events = [created, ...events];
      activeTab = 'list';
      
      toast.push(`イベント「${created.title}」を作成しました！`);
      
      // 入力値のクリア
      if (type === 'ai') {
        aiTextInput = '';
        parsedTitle = '';
        parsedDescription = '';
        parsedCandidates = [];
      } else {
        manualTitle = '';
        manualDescription = '';
        manualCandidates = [{ event_date: '', start_time: '19:00', end_time: '21:00' }];
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : '作成に失敗しました。';
      toast.push('イベントの作成に失敗しました: ' + msg, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    }
  }

  async function deleteEvent(id: string, title: string) {
    if (!confirm(`イベント「${title}」を削除しますか？\nこの操作は取り消せません。`)) {
      return;
    }

    try {
      await api.delete(`/events/${id}`);
      events = events.filter(e => e.id !== id);
      toast.push(`イベント「${title}」を削除しました`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : '削除に失敗しました';
      toast.push(`イベントの削除に失敗しました: ${msg}`, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    }
  }
</script>

<!-- noindex: 管理画面は検索エンジンに表示しない -->
<svelte:head>
  <title>ダッシュボード | 幹事ちゃん</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="container admin-container animate-fade-in">
  {#if loading}
    <div class="glass-panel loading-panel" role="status" aria-label="読み込み中">
      <div class="spinner"></div>
      <p>ダッシュボードデータをロード中...</p>
    </div>
  {:else}
    <div class="admin-grid">
      <!-- Sidebar / Tab Selector -->
      <aside class="admin-sidebar glass-panel">
        <div class="sidebar-header">
          <span class="material-symbols-rounded" aria-hidden="true">admin_panel_settings</span>
          <h3>幹事ダッシュボード</h3>
        </div>
        <div class="sidebar-menu" role="tablist" aria-label="ダッシュボードメニュー">
          <button 
            class="sidebar-btn" 
            class:active={activeTab === 'list'}
            role="tab"
            aria-selected={activeTab === 'list'}
            onclick={() => activeTab = 'list'}
          >
            <span class="material-symbols-rounded" aria-hidden="true">list_alt</span>
            イベント一覧
          </button>
          <button 
            class="sidebar-btn" 
            class:active={activeTab === 'create-ai'}
            role="tab"
            aria-selected={activeTab === 'create-ai'}
            onclick={() => activeTab = 'create-ai'}
          >
            <span class="material-symbols-rounded" aria-hidden="true">auto_awesome</span>
            AIでイベント作成
          </button>
          <button 
            class="sidebar-btn" 
            class:active={activeTab === 'create-manual'}
            role="tab"
            aria-selected={activeTab === 'create-manual'}
            onclick={() => activeTab = 'create-manual'}
          >
            <span class="material-symbols-rounded" aria-hidden="true">add_circle</span>
            手動でイベント作成
          </button>
          <button 
            class="sidebar-btn" 
            class:active={activeTab === 'settings'}
            role="tab"
            aria-selected={activeTab === 'settings'}
            onclick={() => activeTab = 'settings'}
          >
            <span class="material-symbols-rounded" aria-hidden="true">settings</span>
            AI設定 (APIキー)
          </button>
        </div>
      </aside>

      <!-- Main Content Area -->
      <section class="admin-main-content">
        {#if activeTab === 'list'}
          <div class="glass-panel">
            <h2 class="section-title">作成した調整イベント一覧</h2>
            {#if events.length === 0}
              <div class="empty-state">
                <span class="material-symbols-rounded empty-icon" aria-hidden="true">event_busy</span>
                <p>まだ作成されたイベントはありません。</p>
                <button class="btn btn-primary" onclick={() => activeTab = 'create-ai'}>
                  初のイベントを作成する
                </button>
              </div>
            {:else}
              <div class="events-list">
                {#each events as event}
                  <div class="event-row glass-panel">
                    <div class="event-info">
                      <h4>{event.title}</h4>
                      <p class="event-desc">{event.description || '説明はありません。'}</p>
                      <div class="event-meta">
                        <span class="status-badge" class:confirmed={event.status === 'confirmed'}>
                          {event.status === 'confirmed' ? '確定済み' : '調整中'}
                        </span>
                        <span class="date-meta">作成日: {new Date(event.created_at).toLocaleDateString()}</span>
                      </div>
                    </div>
                    <div class="event-actions">
                      <a href={`/event/${event.id}`} class="btn btn-secondary btn-sm">回答ページ</a>
                      <a href={`/admin/event/${event.id}`} class="btn btn-primary btn-sm">管理・AI提案</a>
                      <button 
                        class="btn btn-danger btn-sm-del" 
                        title="イベント削除"
                        aria-label={`イベント「${event.title}」を削除`}
                        onclick={() => deleteEvent(event.id, event.title)}
                      >
                        <span class="material-symbols-rounded" aria-hidden="true">delete</span>
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

        {:else if activeTab === 'create-ai'}
          <div class="glass-panel">
            <h2 class="section-title">
              <span class="gradient-text">AIアシスタントでイベント作成</span>
            </h2>
            <p class="tab-intro">やりたいイベントの内容や日時の希望を自然文で書くと、AIが調整用の日程候補を抽出します。</p>
            
            <form onsubmit={parseNaturalLanguage} class="ai-prompt-form">
              <div class="form-group">
                <label for="ai-text">イベントの希望内容</label>
                <textarea 
                  id="ai-text" 
                  rows="4" 
                  placeholder="e.g. 来週の平日（月曜〜水曜）の19時以降で、新宿付近で3人〜4人で懇親会をやりたい。時間は2時間程度で。候補日は3個くらい出して。"
                  bind:value={aiTextInput}
                ></textarea>
              </div>
              <button type="submit" class="btn btn-primary" disabled={isParsing}>
                <span class="material-symbols-rounded" aria-hidden="true">auto_awesome</span>
                {isParsing ? 'AIが解析中...' : 'AIに日程を抽出してもらう'}
              </button>
            </form>

            {#if parseError}
              <div class="error-banner glass-panel" role="alert">
                <span class="material-symbols-rounded" aria-hidden="true">error</span>
                <p>{parseError}</p>
              </div>
            {/if}

            {#if parsedTitle}
              <div class="parsed-result-area animate-fade-in">
                <hr class="divider" />
                <h3>AI抽出結果の確認・修正</h3>
                <p class="helper-text">必要に応じて、内容や日程候補を修正・追加して確定してください。</p>

                <div class="form-group">
                  <label for="parsed-title">イベント名</label>
                  <input type="text" id="parsed-title" bind:value={parsedTitle} />
                </div>

                <div class="form-group">
                  <label for="parsed-desc">イベントの説明</label>
                  <textarea id="parsed-desc" rows="3" bind:value={parsedDescription}></textarea>
                </div>

                <div class="candidates-editor">
                  <span class="form-label" id="ai-cand-label">日程候補スロット</span>
                  <div role="group" aria-labelledby="ai-cand-label">
                    {#each parsedCandidates as cand, index}
                      <div class="candidate-row">
                        <input type="date" bind:value={cand.event_date} aria-label={`AI抽出 候補日 ${index + 1}`} />
                        <input type="time" bind:value={cand.start_time} aria-label={`AI抽出 開始時刻 ${index + 1}`} />
                        <span class="time-separator" aria-hidden="true">〜</span>
                        <input type="time" bind:value={cand.end_time} aria-label={`AI抽出 終了時刻 ${index + 1}`} />
                        <button 
                          type="button" 
                          class="btn-icon" 
                          onclick={() => removeCandidate(index, 'ai')}
                          aria-label={`候補日 ${index + 1} を削除`}
                        >
                          <span class="material-symbols-rounded" aria-hidden="true">delete</span>
                        </button>
                      </div>
                    {/each}
                  </div>
                  <button 
                    type="button" 
                    class="btn btn-secondary btn-sm add-cand-btn"
                    onclick={() => addCandidate('ai')}
                  >
                    <span class="material-symbols-rounded" aria-hidden="true">add</span>
                    候補日時を追加
                  </button>
                </div>

                <button 
                  type="button" 
                  class="btn btn-primary btn-lg submit-event-btn"
                  onclick={() => submitEvent('ai')}
                >
                  <span class="material-symbols-rounded" aria-hidden="true">check_circle</span>
                  この内容でイベントを確定作成
                </button>
              </div>
            {/if}
          </div>

        {:else if activeTab === 'create-manual'}
          <div class="glass-panel">
            <h2 class="section-title">手動でイベントを作成</h2>
            <p class="tab-intro">通常の日程調整と同様に、すべての候補日を手動で入力してイベントを作成します。</p>

            <div class="form-group">
              <label for="manual-title">イベント名</label>
              <input 
                type="text" 
                id="manual-title" 
                placeholder="e.g. 開発チームミーティング" 
                bind:value={manualTitle} 
              />
            </div>

            <div class="form-group">
              <label for="manual-desc">イベントの説明</label>
              <textarea 
                id="manual-desc" 
                rows="3" 
                placeholder="e.g. 今期のアジェンダの整理を行います。" 
                bind:value={manualDescription}
              ></textarea>
            </div>

            <div class="candidates-editor">
              <span class="form-label" id="manual-cand-label">日程候補スロット</span>
              <div role="group" aria-labelledby="manual-cand-label">
                {#each manualCandidates as cand, index}
                  <div class="candidate-row">
                    <input type="date" bind:value={cand.event_date} aria-label={`手動入力 候補日 ${index + 1}`} />
                    <input type="time" bind:value={cand.start_time} aria-label={`手動入力 開始時刻 ${index + 1}`} />
                    <span class="time-separator" aria-hidden="true">〜</span>
                    <input type="time" bind:value={cand.end_time} aria-label={`手動入力 終了時刻 ${index + 1}`} />
                    <button 
                      type="button" 
                      class="btn-icon" 
                      onclick={() => removeCandidate(index, 'manual')}
                      aria-label={`候補日 ${index + 1} を削除`}
                    >
                      <span class="material-symbols-rounded" aria-hidden="true">delete</span>
                    </button>
                  </div>
                {/each}
              </div>
              <button 
                type="button" 
                class="btn btn-secondary btn-sm add-cand-btn"
                onclick={() => addCandidate('manual')}
              >
                <span class="material-symbols-rounded" aria-hidden="true">add</span>
                候補日時を追加
              </button>
            </div>

            <button 
              type="button" 
              class="btn btn-primary btn-lg submit-event-btn"
              onclick={() => submitEvent('manual')}
            >
              <span class="material-symbols-rounded" aria-hidden="true">check_circle</span>
              イベントを作成
            </button>
          </div>

        {:else if activeTab === 'settings'}
          <div class="glass-panel">
            <h2 class="section-title">AI設定 (Gemini APIキー)</h2>
            <p class="tab-intro">
              イベント作成のアシストや日程の絞り込み機能には AI (Gemini API) を使用します。
              幹事ごとのプライベートなAPIキーを設定できます（暗号化してデータベースに安全に保持されます）。
              未設定の場合は、システムデフォルトのキーが使用されます。
            </p>

            <form onsubmit={updateApiKey}>
              <div class="form-group">
                <label for="api-key">Gemini APIキー</label>
                <input 
                  type="password" 
                  id="api-key" 
                  placeholder="AI-key-xxxx..." 
                  bind:value={apiKeyInput} 
                />
                <p class="helper-text">Google AI Studioから取得したAPIキーを入力してください。</p>
              </div>

              {#if apiKeyUpdateSuccess}
                <p class="success-text" role="status">{apiKeyUpdateSuccess}</p>
              {/if}

              <button type="submit" class="btn btn-primary">
                <span class="material-symbols-rounded" aria-hidden="true">save</span>
                設定を保存する
              </button>
            </form>
          </div>
        {/if}
      </section>
    </div>
  {/if}
</div>

<style>
  .admin-container {
    padding: 1.5rem 0;
  }

  .loading-panel {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1.5rem;
    padding: 5rem 0;
  }

  .spinner {
    width: 50px;
    height: 50px;
    border: 5px solid var(--border-glass);
    border-top: 5px solid var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* Admin Grid Layout */
  .admin-grid {
    display: grid;
    grid-template-columns: 280px 1fr;
    gap: 2rem;
  }

  @media (max-width: 900px) {
    .admin-grid {
      grid-template-columns: 1fr;
    }
  }

  .admin-sidebar {
    padding: 1.5rem;
    height: fit-content;
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--border-glass);
  }

  .sidebar-header h3 {
    font-size: 1.1rem;
  }

  .sidebar-menu {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .sidebar-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.85rem 1rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-family: var(--font-display);
    font-weight: 600;
    text-align: left;
    transition: color var(--transition-fast), background-color var(--transition-fast);
  }

  .sidebar-btn:hover {
    background: var(--bg-secondary);
    color: var(--text-primary);
  }

  .sidebar-btn.active {
    background: var(--color-accent);
    color: #FAF8F5;
  }

  .sidebar-btn .material-symbols-rounded {
    font-size: 1.25rem;
  }

  .tab-intro {
    color: var(--text-secondary);
    font-size: 0.95rem;
    margin-bottom: 2rem;
  }

  /* Events List */
  .events-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .event-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.8rem 2rem;
    background: #FAF8F5;
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
  }

  @media (max-width: 600px) {
    .event-row {
      flex-direction: column;
      align-items: flex-start;
      gap: 1.5rem;
    }
  }

  .event-info h4 {
    font-size: 1.2rem;
    margin-bottom: 0.35rem;
  }

  .event-desc {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin-bottom: 0.75rem;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .event-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
    font-size: 0.8rem;
  }

  .status-badge {
    padding: 0.35rem 0.8rem;
    border-radius: var(--radius-full);
    background: rgba(208, 169, 126, 0.08);
    color: var(--color-maybe);
    border: 1px solid var(--color-maybe);
    font-weight: 600;
    font-size: 0.75rem;
    letter-spacing: 0.03em;
  }

  .status-badge.confirmed {
    background: rgba(94, 111, 98, 0.08);
    color: var(--color-ok);
    border-color: var(--color-ok);
  }

  .date-meta {
    color: var(--text-muted);
    font-size: 0.85rem;
  }

  .event-actions {
    display: flex;
    gap: 0.75rem;
  }

  .empty-state {
    text-align: center;
    padding: 4rem 2rem;
  }

  .empty-icon {
    font-size: 4rem;
    color: var(--text-muted);
    margin-bottom: 1rem;
  }

  .empty-state p {
    color: var(--text-secondary);
    margin-bottom: 1.5rem;
  }

  /* Form & Editor Styles */
  .ai-prompt-form {
    margin-bottom: 2rem;
  }

  .divider {
    border: 0;
    border-top: 1px solid var(--border-glass);
    margin: 2.5rem 0;
  }

  .helper-text {
    font-size: 0.8rem;
    color: var(--text-muted);
    margin-top: 0.4rem;
  }

  .error-banner {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    background: hsla(354, 90%, 60%, 0.1);
    border-color: var(--color-ng);
    color: var(--color-ng);
    padding: 1rem 1.5rem;
    margin-top: 1.5rem;
  }

  .candidates-editor {
    margin: 2rem 0;
  }

  .candidate-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
    animation: fadeIn 0.2s ease forwards;
  }

  .candidate-row input[type="date"] {
    flex: 1.5;
  }

  .candidate-row input[type="time"] {
    flex: 1;
  }

  .time-separator {
    color: var(--text-secondary);
  }

  .btn-icon {
    background: transparent;
    border: none;
    color: var(--text-muted);
    padding: 0.5rem;
    cursor: pointer;
    border-radius: var(--radius-sm);
    transition: color var(--transition-fast), background-color var(--transition-fast);
  }

  .btn-icon:hover {
    color: var(--color-ng);
    background: hsla(354, 90%, 60%, 0.1);
  }

  .add-cand-btn {
    margin-top: 0.5rem;
  }

  .submit-event-btn {
    margin-top: 1.5rem;
    width: 100%;
  }

  .success-text {
    color: var(--color-ok);
    font-size: 0.85rem;
    margin-bottom: 1rem;
    font-weight: 500;
  }

  .btn-danger {
    background-color: hsla(350, 89%, 60%, 0.15);
    border: 1px solid var(--color-ng);
    color: var(--color-ng);
  }

  .btn-danger:hover {
    background-color: var(--color-ng);
    color: #fff;
  }

  .btn-sm-del {
    padding: 0.5rem;
    font-size: 0.85rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: color var(--transition-fast), background-color var(--transition-fast), border-color var(--transition-fast);
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
</style>
