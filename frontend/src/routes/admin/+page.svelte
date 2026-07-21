<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api';
  import { Accordion, AccordionItem } from '$lib';
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

  interface UserApiKey {
    id: number;
    name: string;
    key_prefix: string;
    created_at: string;
    last_used_at?: string;
  }

  // Svelte 5 Runes for Reactivity
  let user = $state<User | null>(null);
  let events = $state<Event[]>([]);
  let userApiKeys = $state<UserApiKey[]>([]);
  let activeTab = $state<'list' | 'create-ai' | 'create-manual' | 'settings'>('list');
  let loading = $state(true);

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '';

  // Gemini API Key state
  let apiKeyInput = $state('');
  let apiKeyUpdateSuccess = $state('');

  // 幹事ちゃん API Key state (MCP・外部API連携用)
  let newKeyName = $state('');
  let createdRawKey = $state<string | null>(null);
  let isGeneratingKey = $state(false);

  // AI Creation state
  let aiTextInput = $state('');
  let isParsing = $state(false);
  let parseError = $state('');
  let parsedTitle = $state('');
  let parsedDescription = $state('');
  let parsedCandidates = $state<CandidateInput[]>([]);

  // AI Prompt Templates
  const aiTemplates = [
    {
      label: '🍻 渋谷で懇親会',
      text: '来週の平日夜（月曜〜水曜）で、新宿・渋谷付近で3名で懇親会をやりたい。時間は19時〜21時で、候補日は3つ出してください。'
    },
    {
      label: '💻 チーム開発MTG',
      text: '来週の木曜日か金曜日の午後（14:00〜17:00）の枠で、オンライン開発ミーティングを1時間やりたい。候補を3つ挙げてください。'
    },
    {
      label: '☕ 週末カフェ勉強会',
      text: '今週末の土曜・日曜の13時〜16時で、カフェ勉強会を開催したいです。土日それぞれ候補を出して。'
    }
  ];

  let parsingStep = $state(0);
  let parsingTimer: ReturnType<typeof setInterval> | null = null;

  function applyAiTemplate(text: string) {
    aiTextInput = text;
    toast.push('テンプレートを入力エリアにセットしました');
  }

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

      const keysData = await api.get<UserApiKey[]>('/auth/apikeys');
      userApiKeys = keysData;
    } catch (err) {
      console.error('Failed to load dashboard data:', err);
      // 未ログインの場合はログインへ誘導
      window.location.href = `${apiBaseUrl}/api/auth/login`;
    } finally {
      loading = false;
    }
  });

  async function generateUserApiKey(e: SubmitEvent) {
    e.preventDefault();
    isGeneratingKey = true;
    try {
      const res = await api.post<{ id: number; name: string; key: string; key_prefix: string; created_at: string }>(
        '/auth/apikeys',
        { name: newKeyName }
      );
      createdRawKey = res.key;
      userApiKeys = [res, ...userApiKeys];
      newKeyName = '';
      toast.push('新しいAPIキーを発行しました！');
    } catch (err) {
      const msg = err instanceof Error ? err.message : '発行に失敗しました';
      toast.push('APIキーの発行に失敗しました: ' + msg, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    } finally {
      isGeneratingKey = false;
    }
  }

  async function deleteUserApiKey(id: number, name: string) {
    if (!confirm(`APIキー「${name}」を削除しますか？\nこのキーを使用している連携サービス（MCPやAPI）はアクセスできなくなります。`)) {
      return;
    }
    try {
      await api.delete(`/auth/apikeys/${id}`);
      userApiKeys = userApiKeys.filter(k => k.id !== id);
      toast.push(`APIキー「${name}」を削除しました`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : '削除に失敗しました';
      toast.push('APIキーの削除に失敗しました: ' + msg, {
        theme: {
          '--toastBackground': 'var(--color-ng)',
          '--toastBarBackground': 'rgba(255, 255, 255, 0.3)'
        }
      });
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    toast.push('APIキーをクリップボードにコピーしました！');
  }

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
    parsingStep = 0;

    if (parsingTimer) clearInterval(parsingTimer);
    parsingTimer = setInterval(() => {
      parsingStep = (parsingStep + 1) % 3;
    }, 1000);
    
    try {
      const result = await api.post<{ title: string; description: string; candidates: CandidateInput[] }>(
        '/ai/parse-event', 
        { text: aiTextInput }
      );
      
      parsedTitle = result.title;
      parsedDescription = result.description;
      parsedCandidates = result.candidates;
      toast.push('✨ AIによる自然文の解析が完了しました！');
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

<svelte:head>
  <title>幹事ダッシュボード | 幹事ちゃん</title>
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
      <!-- Mobile Accordion Tab Switcher (スマホ用アコーディオンメニュー) -->
      <div class="admin-mobile-menu">
        <div class="sidebar-header">
          <span class="material-symbols-rounded" aria-hidden="true">admin_panel_settings</span>
          <h3>幹事ダッシュボード</h3>
        </div>
        <Accordion>
          <AccordionItem title="イベント一覧" icon="list_alt" badge={`${events.length}件`} open={activeTab === 'list'}>
            <button class="btn btn-secondary btn-sm w-full mobile-tab-select-btn" onclick={() => activeTab = 'list'}>
              {activeTab === 'list' ? '✓ 表示中' : 'このタブを表示する'}
            </button>
          </AccordionItem>
          <AccordionItem title="AIでイベント作成" icon="auto_awesome" open={activeTab === 'create-ai'}>
            <button class="btn btn-primary btn-sm w-full mobile-tab-select-btn" onclick={() => activeTab = 'create-ai'}>
              {activeTab === 'create-ai' ? '✓ 表示中' : 'AI作成フォームを表示'}
            </button>
          </AccordionItem>
          <AccordionItem title="手動でイベント作成" icon="add_circle" open={activeTab === 'create-manual'}>
            <button class="btn btn-secondary btn-sm w-full mobile-tab-select-btn" onclick={() => activeTab = 'create-manual'}>
              {activeTab === 'create-manual' ? '✓ 表示中' : '手動作成フォームを表示'}
            </button>
          </AccordionItem>
          <AccordionItem title="ユーザー設定 & APIキー" icon="settings" open={activeTab === 'settings'}>
            <button class="btn btn-secondary btn-sm w-full mobile-tab-select-btn" onclick={() => activeTab = 'settings'}>
              {activeTab === 'settings' ? '✓ 表示中' : '設定パネルを表示'}
            </button>
          </AccordionItem>
        </Accordion>
      </div>

      <!-- Desktop Sidebar / Tab Selector -->
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
            イベント一覧 ({events.length})
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
            ユーザー設定 & APIキー
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
              <span class="gradient-text">✨ AIアシスタントでイベント作成</span>
            </h2>
            <p class="tab-intro">やりたいイベントの内容や日時の希望を自然文で書くと、AIが調整用の日程候補を自動抽出します。</p>

            <!-- Quick Template Chips -->
            <div class="template-chips-area">
              <span class="chips-label">💡 ワンタップ入力テンプレート:</span>
              <div class="chips-list">
                {#each aiTemplates as tpl}
                  <button 
                    type="button" 
                    class="chip-btn"
                    onclick={() => applyAiTemplate(tpl.text)}
                  >
                    {tpl.label}
                  </button>
                {/each}
              </div>
            </div>
            
            <form onsubmit={parseNaturalLanguage} class="ai-prompt-form">
              <div class="form-group">
                <label for="ai-text">イベントの希望内容</label>
                <textarea 
                  id="ai-text" 
                  rows="4" 
                  placeholder="e.g. 来週の平日（月曜〜水曜）の19時以降で、新宿付近で3名で懇親会をやりたい。時間は2時間程度で。候補日は3個くらい出して。"
                  bind:value={aiTextInput}
                  disabled={isParsing}
                ></textarea>
              </div>
              <button type="submit" class="btn btn-primary ai-submit-btn" class:parsing={isParsing} disabled={isParsing}>
                <span class="material-symbols-rounded ai-sparkle-icon" class:spin={isParsing} aria-hidden="true">auto_awesome</span>
                {isParsing ? 'AIが解析中...' : 'AIに日程を抽出してもらう'}
              </button>
            </form>

            {#if isParsing}
              <div class="ai-loading-card glass-panel animate-fade-in" role="status">
                <div class="shimmer-bar"></div>
                <div class="loading-status-content">
                  <div class="ai-pulse-orb"></div>
                  <div class="loading-text-wrapper">
                    {#if parsingStep === 0}
                      <p class="step-text animate-slide-up">🧠 自然文の文章・希望条件を解析中...</p>
                    {:else if parsingStep === 1}
                      <p class="step-text animate-slide-up">📅 日時フォーマット・候補日スロットを抽出中...</p>
                    {:else}
                      <p class="step-text animate-slide-up">✍️ イベントタイトルと説明文を生成中...</p>
                    {/if}
                  </div>
                </div>
              </div>
            {/if}

            {#if parseError}
              <div class="error-banner glass-panel" role="alert">
                <span class="material-symbols-rounded" aria-hidden="true">error</span>
                <p>{parseError}</p>
              </div>
            {/if}

            {#if parsedTitle && !isParsing}
              <div class="parsed-result-area animate-fade-in">
                <hr class="divider" />
                <div class="result-header-badge">
                  <span class="material-symbols-rounded spark-icon" aria-hidden="true">auto_awesome</span>
                  <h3>AI抽出結果の確認・修正</h3>
                </div>
                <p class="helper-text">抽出された内容と日程候補です。必要に応じて微調整して確定してください。</p>

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
                      <div class="candidate-row animated-row" style="animation-delay: {index * 0.08}s">
                        <span class="row-num font-mono">{index + 1}</span>
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
                  class="btn btn-primary btn-lg submit-event-btn pulse-glow"
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
          <div class="glass-panel settings-container">
            <h2 class="section-title">ユーザー設定 & 連携</h2>

            <!-- 幹事ちゃん APIキー (MCP・外部連携) セクション -->
            <div class="settings-section">
              <div class="section-header">
                <span class="material-symbols-rounded section-icon" aria-hidden="true">key</span>
                <div>
                  <h3>幹事ちゃん APIキー (MCP & REST API)</h3>
                  <p class="section-desc">
                    外部ツール（MCPクライアント、Pythonスクリプト、curl等）から幹事ちゃんのAPIを利用するためのAPIキーを発行します。
                  </p>
                </div>
              </div>

              {#if createdRawKey}
                <div class="key-created-banner glass-panel animate-fade-in" role="alert">
                  <div class="banner-header">
                    <span class="material-symbols-rounded" aria-hidden="true">verified</span>
                    <strong>APIキーが正常に発行されました</strong>
                  </div>
                  <p class="warning-text">
                    ⚠️ 安全のため、このキーは<strong>今しか表示されません</strong>。必ずコピーして安全な場所に保管してください。
                  </p>
                  <div class="key-copy-box">
                    <code class="raw-key-code">{createdRawKey}</code>
                    <button class="btn btn-primary btn-sm" onclick={() => copyToClipboard(createdRawKey!)}>
                      <span class="material-symbols-rounded" aria-hidden="true">content_copy</span>
                      コピー
                    </button>
                  </div>
                  <button class="btn btn-secondary btn-sm close-key-btn" onclick={() => createdRawKey = null}>
                    閉じる
                  </button>
                </div>
              {/if}

              <form onsubmit={generateUserApiKey} class="generate-key-form">
                <div class="form-row">
                  <div class="form-group flex-grow">
                    <label for="new-key-name">APIキーの識別名</label>
                    <input 
                      type="text" 
                      id="new-key-name" 
                      placeholder="例: My MCP Client, Claude Desktop" 
                      bind:value={newKeyName} 
                    />
                  </div>
                  <button type="submit" class="btn btn-primary generate-btn" disabled={isGeneratingKey}>
                    <span class="material-symbols-rounded" aria-hidden="true">add_key</span>
                    {isGeneratingKey ? '発行中...' : 'APIキーを発行'}
                  </button>
                </div>
              </form>

              <div class="apikeys-list-wrapper">
                <h4>発行済みAPIキー</h4>
                {#if userApiKeys.length === 0}
                  <p class="no-keys-text">まだ発行されたAPIキーはありません。</p>
                {:else}
                  <div class="apikeys-table">
                    {#each userApiKeys as k}
                      <div class="apikey-row">
                        <div class="apikey-info">
                          <span class="apikey-name">{k.name}</span>
                          <span class="apikey-prefix"><code>{k.key_prefix}</code></span>
                        </div>
                        <div class="apikey-meta">
                          <span class="key-date">作成日: {new Date(k.created_at).toLocaleDateString()}</span>
                          {#if k.last_used_at}
                            <span class="key-date">最終利用: {new Date(k.last_used_at).toLocaleDateString()}</span>
                          {:else}
                            <span class="key-date">未利用</span>
                          {/if}
                        </div>
                        <button 
                          class="btn btn-danger btn-sm-del" 
                          title="APIキーを削除" 
                          aria-label={`APIキー「${k.name}」を削除`}
                          onclick={() => deleteUserApiKey(k.id, k.name)}
                        >
                          <span class="material-symbols-rounded" aria-hidden="true">delete</span>
                        </button>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>

              <div class="api-usage-guide">
                <h4>🔑 APIキー & MCPサーバーの使い方</h4>
                <p><strong>MCP サーバー URL:</strong> <code>/mcp</code> (Streamable HTTP)</p>
                <p>発行したAPIキーは、HTTPリクエストのヘッダーに設定してご利用ください：</p>
                <pre class="code-block"><code>Authorization: Bearer kc_your_api_key_here
# または
X-API-Key: kc_your_api_key_here</code></pre>
              </div>
            </div>

            <hr class="divider" />

            <!-- Gemini APIキー 設定セクション (既存機能) -->
            <div class="settings-section">
              <div class="section-header">
                <span class="material-symbols-rounded section-icon" aria-hidden="true">psychology</span>
                <div>
                  <h3>AI機能用 Gemini APIキー</h3>
                  <p class="section-desc">
                    イベント作成のアシストや日程の自動絞り込みに使用する Gemini APIキーを登録します（暗号化してデータベースに保存されます）。
                  </p>
                </div>
              </div>

              <form onsubmit={updateApiKey}>
                <div class="form-group">
                  <label for="api-key">Gemini APIキー</label>
                  <input 
                    type="password" 
                    id="api-key" 
                    placeholder="AI-key-xxxx..." 
                    bind:value={apiKeyInput} 
                  />
                  <p class="helper-text">Google AI Studioから取得したAPIキーを入力してください。未設定の場合はシステム共有キーが適用されます。</p>
                </div>

                {#if apiKeyUpdateSuccess}
                  <p class="success-text" role="status">{apiKeyUpdateSuccess}</p>
                {/if}

                <button type="submit" class="btn btn-primary">
                  <span class="material-symbols-rounded" aria-hidden="true">save</span>
                  Gemini APIキーを保存
                </button>
              </form>
            </div>
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

  /* Settings Page Styles */
  .settings-container {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .settings-section {
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
  }

  .section-header {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
  }

  .section-icon {
    font-size: 2rem;
    color: var(--color-primary);
    padding: 0.5rem;
    background: rgba(94, 111, 98, 0.1);
    border-radius: var(--radius-md);
  }

  .section-desc {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin-top: 0.25rem;
  }

  .key-created-banner {
    background: rgba(94, 111, 98, 0.12);
    border: 1px solid var(--color-primary);
    padding: 1.25rem;
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .banner-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--color-primary);
    font-size: 1.05rem;
  }

  .warning-text {
    font-size: 0.88rem;
    color: var(--text-primary);
  }

  .key-copy-box {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    background: #FAF8F5;
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-glass);
    overflow-x: auto;
  }

  .raw-key-code {
    font-family: monospace;
    font-size: 0.95rem;
    color: var(--color-accent);
    word-break: break-all;
    flex: 1;
  }

  .close-key-btn {
    align-self: flex-end;
  }

  .generate-key-form .form-row {
    display: flex;
    align-items: flex-end;
    gap: 1rem;
  }

  @media (max-width: 600px) {
    .generate-key-form .form-row {
      flex-direction: column;
      align-items: stretch;
    }
  }

  .flex-grow {
    flex: 1;
  }

  .generate-btn {
    white-space: nowrap;
  }

  .apikeys-list-wrapper {
    margin-top: 1rem;
  }

  .apikeys-list-wrapper h4 {
    font-size: 1rem;
    margin-bottom: 0.75rem;
  }

  .no-keys-text {
    color: var(--text-muted);
    font-size: 0.9rem;
    font-style: italic;
  }

  .apikeys-table {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .apikey-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.85rem 1.25rem;
    background: #FAF8F5;
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    gap: 1rem;
  }

  @media (max-width: 600px) {
    .apikey-row {
      flex-direction: column;
      align-items: flex-start;
    }
  }

  .apikey-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .apikey-name {
    font-weight: 600;
    font-size: 0.95rem;
  }

  .apikey-prefix code {
    background: rgba(0,0,0,0.05);
    padding: 0.2rem 0.4rem;
    border-radius: var(--radius-sm);
    font-size: 0.85rem;
    color: var(--text-secondary);
  }

  .apikey-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .api-usage-guide {
    background: rgba(0,0,0,0.03);
    padding: 1rem 1.25rem;
    border-radius: var(--radius-md);
    border-left: 4px solid var(--color-accent);
    font-size: 0.88rem;
    margin-top: 1rem;
  }

  .api-usage-guide h4 {
    font-size: 0.95rem;
    margin-bottom: 0.5rem;
  }

  .code-block {
    background: #2d3748;
    color: #edf2f7;
    padding: 0.75rem 1rem;
    border-radius: var(--radius-sm);
    margin-top: 0.5rem;
    overflow-x: auto;
    font-size: 0.85rem;
  }

  /* AI UI Animation & Enhancements */
  .template-chips-area {
    margin-bottom: 1.25rem;
  }

  .chips-label {
    font-size: 0.85rem;
    color: var(--text-secondary);
    font-weight: 500;
    display: block;
    margin-bottom: 0.5rem;
  }

  .chips-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .chip-btn {
    background: rgba(94, 111, 98, 0.08);
    border: 1px solid rgba(94, 111, 98, 0.2);
    color: var(--color-primary);
    padding: 0.4rem 0.85rem;
    border-radius: var(--radius-full);
    font-size: 0.83rem;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .chip-btn:hover {
    background: var(--color-primary);
    color: #fff;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(94, 111, 98, 0.25);
  }

  .ai-submit-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    position: relative;
    overflow: hidden;
  }

  .ai-submit-btn.parsing {
    background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  }

  .ai-sparkle-icon.spin {
    animation: rotateSparkle 1.5s infinite linear;
  }

  @keyframes rotateSparkle {
    0% { transform: rotate(0deg) scale(1); }
    50% { transform: rotate(180deg) scale(1.2); }
    100% { transform: rotate(360deg) scale(1); }
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
    left: -100%;
    width: 100%;
    height: 4px;
    background: linear-gradient(90deg, transparent, var(--color-accent), transparent);
    animation: shimmerSlide 2s infinite;
  }

  @keyframes shimmerSlide {
    0% { left: -100%; }
    100% { left: 100%; }
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

  .result-header-badge {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--color-primary);
    margin-bottom: 0.25rem;
  }

  .spark-icon {
    color: var(--color-accent);
  }

  .candidate-row.animated-row {
    animation: rowFadeIn 0.4s ease forwards;
    opacity: 0;
  }

  @keyframes rowFadeIn {
    to { opacity: 1; transform: translateY(0); }
    from { opacity: 0; transform: translateY(8px); }
  }

  /* Mobile Menu Styles for Admin Dashboard */
  .admin-mobile-menu {
    display: none;
    margin-bottom: 1rem;
  }

  .mobile-tab-select-btn {
    margin-top: 0.5rem;
  }

  @media (max-width: 768px) {
    .admin-mobile-menu {
      display: block;
    }
    .admin-sidebar {
      display: none;
    }
    .admin-grid {
      grid-template-columns: 1fr;
    }
  }

  .pulse-glow {
    animation: pulseButtonGlow 2.5s infinite;
  }

  @keyframes pulseButtonGlow {
    0%, 100% { box-shadow: 0 0 0 0 rgba(94, 111, 98, 0.4); }
    50% { box-shadow: 0 0 0 10px rgba(94, 111, 98, 0); }
  }
</style>
