<script lang="ts">
  import { goto } from '$app/navigation';
  import { reveal } from '$lib/reveal';

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '';
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

<div class="container hero-container" use:reveal>
  <div class="hero-text-section">
    <div class="badge">Aesthetic & AI Scheduling</div>
    <h1 class="hero-title">
      日程調整を、<br />
      もっと美しく、心地よく。<br />
      <span class="gradient-text">幹事ちゃん</span>
    </h1>
    <p class="hero-description">
      「幹事ちゃん」は、シンプルで洗練されたデザインの日程調整ツール。<br />
      自然文を入力するだけでAIが候補日を自動で整理し、全員の回答から最適な日程を美しく提案します。忙しい日々に、少しのゆとりと調和を。
    </p>
    
    <div class="hero-actions">
      <a href="{apiBaseUrl}/api/auth/login" class="btn btn-primary btn-lg">
        <span class="material-symbols-rounded">login</span>
        幹事として始める (ログイン)
      </a>
    </div>
  </div>

  <div class="hero-form-section">
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
  </div>
</div>

<div class="container features-section" use:reveal>
  <h2 class="section-title">幹事ちゃん の心地よいサポート</h2>
  <div class="features-grid">
    <div class="feature-card glass-panel" use:reveal>
      <span class="material-symbols-rounded feature-icon">edit_note</span>
      <h4>言葉から、候補日を紡ぐ</h4>
      <p>「来週の平日夜、渋谷でランチかお茶。候補日を3つほど」といった自然な言葉から、AIが最適な候補日と時間帯をカレンダーから美しく提案・入力します。</p>
    </div>
    
    <div class="feature-card glass-panel" use:reveal>
      <span class="material-symbols-rounded feature-icon">psychology</span>
      <h4>調和を生み出す決定サポート</h4>
      <p>「仕事帰りに無理なく」「Aさんは必ず招待」といった、数字だけでは測れない幹事の想いと全員の都合をAIが汲み取り、一番心地よい日程を提案します。</p>
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
    gap: 4rem;
    align-items: center;
    padding: 4rem 0;
  }

  @media (max-width: 768px) {
    .hero-container {
      grid-template-columns: 1fr;
      gap: 3rem;
    }
  }

  .badge {
    background: rgba(94, 111, 98, 0.06);
    border: 1px solid rgba(94, 111, 98, 0.25);
    color: var(--color-accent);
    padding: 0.5rem 1.2rem;
    border-radius: var(--radius-full);
    display: inline-block;
    font-size: 0.75rem;
    font-weight: 600;
    margin-bottom: 2rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .hero-title {
    font-size: 3.2rem;
    line-height: 1.25;
    margin-bottom: 2rem;
    font-weight: 400;
  }

  .hero-description {
    font-size: 1.05rem;
    color: var(--text-secondary);
    margin-bottom: 2.5rem;
    line-height: 1.8;
  }

  .hero-actions {
    display: flex;
    gap: 1rem;
  }

  .btn-lg {
    padding: 1.1rem 2.2rem;
    font-size: 1rem;
  }

  .event-code-panel {
    display: flex;
    flex-direction: column;
    gap: 1.8rem;
    border-radius: var(--radius-lg);
  }

  .panel-title {
    font-size: 1.25rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .panel-subtitle {
    color: var(--text-muted);
    font-size: 0.85rem;
    margin-top: -1.2rem;
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
    padding: 6rem 0 3rem 0;
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
