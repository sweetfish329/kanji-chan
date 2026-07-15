<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api';
  import { toast } from '@zerodevx/svelte-toast';
  import { reveal } from '$lib/reveal';

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '';

  // 閲覧・回答用ID
  let eventIdInput = $state('');
  let errorMsg = $state('');

  // 新規イベント作成状態
  let title = $state('');
  let description = $state('');
  let candidates = $state<{ event_date: string; start_time: string; end_time: string }[]>([
    { event_date: '', start_time: '19:00', end_time: '21:00' }
  ]);
  let submitting = $state(false);

  function addCandidate() {
    candidates = [...candidates, { event_date: '', start_time: '19:00', end_time: '21:00' }];
  }

  function removeCandidate(index: number) {
    candidates = candidates.filter((_, i) => i !== index);
  }

  async function createEvent(e: SubmitEvent) {
    e.preventDefault();
    if (!title.trim()) {
      toast.push('イベント名を入力してください');
      return;
    }
    if (candidates.length === 0 || candidates.some(c => !c.event_date)) {
      toast.push('日付をすべて指定してください');
      return;
    }

    submitting = true;
    try {
      const created = await api.post<{ id: string; title: string }>('/events', {
        title,
        description,
        candidates
      });
      toast.push(`イベント「${created.title}」を作成しました！`);
      goto(`/event/${created.id}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : '作成に失敗しました。';
      toast.push('イベント作成に失敗しました: ' + msg);
    } finally {
      submitting = false;
    }
  }

  function navigateToEvent(e: SubmitEvent) {
    e.preventDefault();
    if (!eventIdInput.trim()) {
      errorMsg = 'イベントIDを入力してください';
      return;
    }
    
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(eventIdInput.trim())) {
      errorMsg = '正しいイベントIDの形式で入力してください';
      return;
    }

    goto(`/event/${eventIdInput.trim()}`);
  }
</script>

<div class="container hero-container" use:reveal>
  <!-- 左カラム：いきなり予定作成 -->
  <div class="hero-text-section">
    <div class="glass-panel creation-panel">
      <h2 class="panel-title-large">
        <span class="material-symbols-rounded icon-accent">add_circle</span>
        日程調整を新しくつくる
      </h2>
      <p class="panel-subtitle-large">日程候補とイベント名を入力するだけで、すぐに調整ページを作成できます（ログイン不要）</p>

      <form onsubmit={createEvent}>
        <div class="form-group">
          <label for="event-title">イベント名</label>
          <input 
            type="text" 
            id="event-title" 
            placeholder="e.g. 女子会ランチ、週末カフェ会" 
            bind:value={title} 
            required
          />
        </div>

        <div class="form-group">
          <label for="event-desc">説明（場所や会費など）</label>
          <textarea 
            id="event-desc" 
            rows="2" 
            placeholder="e.g. 渋谷付近でランチしましょう。予算は3000円くらいです。" 
            bind:value={description}
          ></textarea>
        </div>

        <div class="candidates-editor">
          <span class="form-label">候補日時スロット</span>
          <div class="candidate-list">
            {#each candidates as cand, index}
              <div class="candidate-row">
                <input type="date" bind:value={cand.event_date} required />
                <input type="time" bind:value={cand.start_time} required />
                <span class="time-separator">〜</span>
                <input type="time" bind:value={cand.end_time} required />
                {#if candidates.length > 1}
                  <button 
                    type="button" 
                    class="btn-icon" 
                    onclick={() => removeCandidate(index)}
                    title="削除"
                  >
                    <span class="material-symbols-rounded">delete</span>
                  </button>
                {/if}
              </div>
            {/each}
          </div>
          <button 
            type="button" 
            class="btn btn-secondary btn-sm add-cand-btn"
            onclick={addCandidate}
          >
            <span class="material-symbols-rounded">add</span>
            候補日時を追加
          </button>
        </div>

        <button type="submit" class="btn btn-primary btn-lg w-full submit-btn" disabled={submitting}>
          <span class="material-symbols-rounded">check_circle</span>
          {submitting ? '作成中...' : '予定作成 ＆ 調整ページを開く'}
        </button>
      </form>
    </div>
  </div>

  <!-- 右カラム：回答・閲覧 ＆ AIログイン -->
  <div class="hero-form-section">
    <div class="sidebar-cards">
      <!-- 回答・閲覧 -->
      <div class="glass-panel event-code-panel">
        <h3 class="panel-title">イベントへの回答・閲覧</h3>
        <p class="panel-subtitle">招待されたイベントIDを入力してください</p>
        
        <form onsubmit={navigateToEvent}>
          <div class="form-group">
            <label for="event-id">イベントID (UUID)</label>
            <input 
              type="text" 
              id="event-id" 
              placeholder="e.g. 12345678-abcd-1234-ef00-1234567890ab" 
              bind:value={eventIdInput}
            />
            {#if errorMsg}
              <p class="error-text">{errorMsg}</p>
            {/if}
          </div>
          
          <button type="submit" class="btn btn-secondary w-full">
            <span class="material-symbols-rounded">arrow_forward</span>
            イベントページを開く
          </button>
        </form>
      </div>

      <!-- AI機能を使うにはログイン -->
      <div class="ai-login-card glass-panel">
        <div class="ai-card-header">
          <span class="material-symbols-rounded ai-icon">auto_awesome</span>
          <h4>AI日程決定 ＆ アシスト</h4>
        </div>
        <p class="ai-card-text">
          自然文からの候補日自動抽出や、回答結果からAIが最適な日程を自動分析・決定する機能を利用する場合は、ログインしてご利用ください。
        </p>
        <a href="{apiBaseUrl}/api/auth/login" class="btn btn-primary w-full login-btn">
          <span class="material-symbols-rounded">login</span>
          AI機能を使うにはログイン
        </a>
      </div>
    </div>
  </div>
</div>

<div class="container features-section" use:reveal>
  <h2 class="section-title">幹事ちゃん の心地よいサポート</h2>
  <div class="features-grid">
    <div class="feature-card glass-panel" use:reveal>
      <span class="material-symbols-rounded feature-icon">edit_note</span>
      <h4>言葉から、候補日を紡ぐ</h4>
      <p>「来週の平日夜、渋谷でランチかお茶。候補日を3つほど」といった自然な言葉から、AIが最適な候補日と時間帯をカレンダーから美しく提案・入力します。（ログインが必要です）</p>
    </div>
    
    <div class="feature-card glass-panel" use:reveal>
      <span class="material-symbols-rounded feature-icon">psychology</span>
      <h4>調和を生み出す決定サポート</h4>
      <p>「仕事帰りに無理なく」「Aさんは必ず招待」といった、数字だけでは測れない幹事の想いと全員の都合をAIが汲み取り、一番心地よい日程を提案します。（ログインが必要です）</p>
    </div>

    <div class="feature-card glass-panel" use:reveal>
      <span class="material-symbols-rounded feature-icon">person_check</span>
      <h4>おもてなしのシンプル回答</h4>
      <p>回答するメンバーは登録やログインが不要。馴染み深い「〇・△・×」のシンプルなテーブルで、どのデバイスからも迷わずすぐに回答できます。</p>
    </div>
  </div>
</div>

<style>
  .hero-container {
    display: grid;
    grid-template-columns: 1.2fr 0.8fr;
    gap: 3rem;
    align-items: flex-start;
    padding: 3rem 0;
  }

  @media (max-width: 900px) {
    .hero-container {
      grid-template-columns: 1fr;
      gap: 2.5rem;
    }
  }

  .creation-panel {
    border-radius: var(--radius-lg);
    background: var(--bg-glass);
    padding: 3rem 2.5rem;
  }

  .panel-title-large {
    font-size: 1.8rem;
    font-weight: 500;
    margin-bottom: 0.5rem;
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .panel-subtitle-large {
    color: var(--text-secondary);
    font-size: 0.95rem;
    margin-bottom: 2.5rem;
  }

  .icon-accent {
    color: var(--color-accent);
  }

  .sidebar-cards {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .event-code-panel, .ai-login-card {
    border-radius: var(--radius-lg);
    padding: 2.2rem;
  }

  .panel-title {
    font-size: 1.25rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .panel-subtitle {
    color: var(--text-muted);
    font-size: 0.85rem;
    margin-top: -0.2rem;
    margin-bottom: 1.5rem;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .w-full {
    width: 100%;
  }

  .error-text {
    color: var(--color-ng);
    font-size: 0.8rem;
    margin-top: 0.5rem;
  }

  /* Candidates Editor */
  .candidates-editor {
    margin: 2rem 0;
  }

  .candidate-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 0.6rem;
  }

  .candidate-row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .candidate-row input[type="date"] {
    flex: 1.4;
  }

  .candidate-row input[type="time"] {
    flex: 1;
  }

  .time-separator {
    color: var(--text-muted);
  }

  .btn-icon {
    background: transparent;
    border: none;
    color: var(--text-muted);
    padding: 0.5rem;
    cursor: pointer;
    border-radius: var(--radius-sm);
    transition: all var(--transition-fast);
  }

  .btn-icon:hover {
    color: var(--color-ng);
    background: rgba(194, 134, 127, 0.08);
  }

  .add-cand-btn {
    margin-top: 0.8rem;
  }

  .submit-btn {
    margin-top: 1rem;
  }

  /* AI Login Card */
  .ai-card-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .ai-icon {
    color: var(--color-accent);
    font-size: 1.5rem;
  }

  .ai-card-header h4 {
    font-size: 1.15rem;
    font-weight: 500;
  }

  .ai-card-text {
    font-size: 0.85rem;
    color: var(--text-secondary);
    margin-bottom: 1.5rem;
    line-height: 1.7;
  }

  .login-btn {
    background: var(--color-accent);
  }

  .login-btn:hover {
    background: #4D5C51;
  }

  /* Features Section */
  .features-section {
    padding: 6rem 0 3rem 0;
  }

  .section-title {
    text-align: center;
    font-size: 2rem;
    margin-bottom: 4rem;
    font-weight: 400;
  }

  .features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 2rem;
  }

  .feature-card {
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
    padding: 3rem 2.2rem;
    border-radius: var(--radius-lg);
  }

  .feature-icon {
    font-size: 2.2rem;
    color: var(--color-accent);
  }

  .feature-card h4 {
    font-size: 1.2rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .feature-card p {
    color: var(--text-secondary);
    font-size: 0.9rem;
    line-height: 1.7;
  }
</style>
