<script lang="ts">
  import { goto } from '$app/navigation';

  let eventIdInput = $state('');
  let errorMsg = $state('');

  function navigateToEvent(e: SubmitEvent) {
    e.preventDefault();
    if (!eventIdInput.trim()) {
      errorMsg = 'イベントIDを入力してください';
      return;
    }
    
    // UUIDの簡易チェック
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(eventIdInput.trim())) {
      errorMsg = '正しいイベントID (UUID) の形式で入力してください';
      return;
    }

    goto(`/event/${eventIdInput.trim()}`);
  }
</script>

<div class="container hero-container animate-fade-in">
  <div class="hero-text-section">
    <div class="badge">AI-Powered Scheduling</div>
    <h1 class="hero-title">
      日程調整を、もっとスマートに。<br />
      <span class="gradient-text">幹事ちゃん</span> がお手伝い。
    </h1>
    <p class="hero-description">
      「幹事ちゃん」は、日本の定番日程調整ツールのようなシンプルさをベースに、
      自然文からのイベント自動作成や、回答データから最適な日程をAIが絞り込んでくれる、
      AIサポート付き日程調整アプリケーションです。
    </p>
    
    <div class="hero-actions">
      <a href="http://localhost:8080/api/auth/login" class="btn btn-primary btn-lg">
        <span class="material-symbols-rounded">login</span>
        幹事として始める (ログイン)
      </a>
    </div>
  </div>

  <div class="hero-form-section">
    <div class="glass-panel event-code-panel">
      <h3 class="panel-title">イベントへの回答・閲覧</h3>
      <p class="panel-subtitle">幹事から共有されたイベントIDを入力してください</p>
      
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
  </div>
</div>

<div class="container features-section animate-fade-in" style="animation-delay: 0.1s">
  <h2 class="section-title">幹事ちゃん の主なAI機能</h2>
  <div class="features-grid">
    <div class="feature-card glass-panel">
      <span class="material-symbols-rounded feature-icon">edit_note</span>
      <h4>自然文で一発作成</h4>
      <p>「来週の平日夜、渋谷で飲み会。候補日3つくらい」と入力するだけで、AIがタイトル、説明、カレンダー候補日を自動抽出してフォームを埋めます。</p>
    </div>
    
    <div class="feature-card glass-panel">
      <span class="material-symbols-rounded feature-icon">psychology</span>
      <h4>AIによる最適日程提案</h4>
      <p>参加者全員の〇△×が出揃ったら、AIが優先順位（「平日のほうが良い」「Aさんは必須」など）を加味して、おすすめ候補日のスコアと詳細な理由を提案します。</p>
    </div>

    <div class="feature-card glass-panel">
      <span class="material-symbols-rounded feature-icon">person_check</span>
      <h4>かんたんログイン不要回答</h4>
      <p>参加者はアカウント登録やログインなしで、従来の調整さんと同じように直感的な〇△×テーブルで日程を入力できます。</p>
    </div>
  </div>
</div>

<style>
  .hero-container {
    display: grid;
    grid-template-columns: 1.2fr 0.8fr;
    gap: 3rem;
    align-items: center;
    padding: 3rem 0;
  }

  @media (max-width: 768px) {
    .hero-container {
      grid-template-columns: 1fr;
      gap: 2rem;
    }
  }

  .badge {
    background: hsla(263, 90%, 65%, 0.15);
    border: 1px solid var(--color-primary);
    color: var(--text-primary);
    padding: 0.4rem 1rem;
    border-radius: var(--radius-full);
    display: inline-block;
    font-size: 0.8rem;
    font-weight: 600;
    margin-bottom: 1.5rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .hero-title {
    font-size: 3rem;
    line-height: 1.15;
    margin-bottom: 1.5rem;
  }

  .hero-description {
    font-size: 1.1rem;
    color: var(--text-secondary);
    margin-bottom: 2rem;
  }

  .hero-actions {
    display: flex;
    gap: 1rem;
  }

  .btn-lg {
    padding: 1rem 2rem;
    font-size: 1.1rem;
  }

  .event-code-panel {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .panel-title {
    font-size: 1.3rem;
  }

  .panel-subtitle {
    color: var(--text-secondary);
    font-size: 0.85rem;
    margin-top: -1rem;
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

  /* Features Section */
  .features-section {
    padding: 5rem 0 2rem 0;
  }

  .section-title {
    text-align: center;
    font-size: 2rem;
    margin-bottom: 3rem;
  }

  .features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
  }

  .feature-card {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 2.5rem 2rem;
    transition: transform var(--transition-normal);
  }

  .feature-card:hover {
    transform: translateY(-5px);
  }

  .feature-icon {
    font-size: 2.5rem;
    color: var(--color-primary);
  }

  .feature-card h4 {
    font-size: 1.25rem;
  }

  .feature-card p {
    color: var(--text-secondary);
    font-size: 0.95rem;
  }
</style>
