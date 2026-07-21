<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api';
  import { Accordion, AccordionItem } from '$lib';
  import { toast } from '@zerodevx/svelte-toast';
  import { reveal } from '$lib/reveal';

  // サイトURL (VITE_PUBLIC_SITE_URL 環境変数から取得。未設定時は空文字 = 相対URL)
  const siteUrl = (import.meta.env.VITE_PUBLIC_SITE_URL ?? '').replace(/\/$/, '');

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

<svelte:head>
  <!-- ==========================================
       SEO: トップページのみ index, follow
       VITE_PUBLIC_SITE_URL 環境変数でURLを制御
       ========================================== -->
  <title>幹事ちゃん — AIが日程調整をおしゃれに、スマートに</title>
  <meta name="robots" content="index, follow" />
  <meta name="description" content="幹事ちゃんは、AIが候補日時を自動提案・分析するモダンな日程調整サービスです。完全無料で登録・ログイン不要。〇△×形式のシンプルな操作で、グループの予定調整をスムーズかつおしゃれに解決します。" />
  <meta name="keywords" content="予定調整, 日程調整, AI, 幹事, 調整さん, スケジュール, グループ調整, おしゃれ, モダン, 無料, ログイン不要" />
  {#if siteUrl}
    <link rel="canonical" href="{siteUrl}/" />
  {/if}

  <!-- Open Graph (SNS シェア用) -->
  <meta property="og:type" content="website" />
  {#if siteUrl}
    <meta property="og:url" content="{siteUrl}/" />
    <meta property="og:image" content="{siteUrl}/ogp.jpg" />
  {:else}
    <meta property="og:image" content="/ogp.jpg" />
  {/if}
  <meta property="og:site_name" content="幹事ちゃん" />
  <meta property="og:title" content="幹事ちゃん — AIが日程調整をおしゃれに、スマートに" />
  <meta property="og:description" content="自然な言葉で候補日を提案、AIが最適な日程を分析。登録不要ですぐ使えるモダンな日程調整サービス。" />
  <meta property="og:locale" content="ja_JP" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="幹事ちゃん — AIが日程調整をおしゃれに、スマートに" />
  <meta name="twitter:description" content="自然な言葉で候補日を提案、AIが最適な日程を分析。登録不要ですぐ使えるモダンな日程調整サービス。" />
  {#if siteUrl}
    <meta name="twitter:image" content="{siteUrl}/ogp.jpg" />
  {:else}
    <meta name="twitter:image" content="/ogp.jpg" />
  {/if}

  <!-- JSON-LD 構造化データ (Google 検索エンジン向け) -->
  {@html `<script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "WebApplication",
    "name": "幹事ちゃん",
    ${siteUrl ? `"url": "${siteUrl}/",` : ''}
    "image": "${siteUrl ? siteUrl : ''}/ogp.jpg",
    "description": "AIが候補日を自動提案・分析するモダンな日程調整サービス。登録不要で〇△×方式のグループスケジュール調整ができます。",
    "applicationCategory": "BusinessApplication",
    "operatingSystem": "Web",
    "offers": {
      "@type": "Offer",
      "price": "0",
      "priceCurrency": "JPY"
    },
    "featureList": [
      "AI自然文からの候補日自動提案",
      "〇△×方式のグループ日程調整",
      "登録・ログイン不要",
      "AIによる最適日程分析",
      "モバイル対応"
    ],
    "inLanguage": "ja"
  }
  <\/script>`}
</svelte:head>

<div class="container hero-container" use:reveal>
  <!-- 左カラム：いきなり予定作成 -->
  <div class="hero-text-section">
    <div class="glass-panel creation-panel">
      <h1 class="panel-title-large">
        <span class="material-symbols-rounded icon-accent" aria-hidden="true">add_circle</span>
        日程調整を新しくつくる
      </h1>
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
          <span class="form-label" id="candidates-label">候補日時スロット</span>
          <div class="candidate-list" role="group" aria-labelledby="candidates-label">
            {#each candidates as cand, index}
              <div class="candidate-row">
                <input 
                  type="date" 
                  bind:value={cand.event_date} 
                  required 
                  aria-label={`候補日 ${index + 1}`} 
                />
                <input 
                  type="time" 
                  bind:value={cand.start_time} 
                  required 
                  aria-label={`開始時刻 ${index + 1}`} 
                />
                <span class="time-separator" aria-hidden="true">〜</span>
                <input 
                  type="time" 
                  bind:value={cand.end_time} 
                  required 
                  aria-label={`終了時刻 ${index + 1}`} 
                />
                {#if candidates.length > 1}
                  <button 
                    type="button" 
                    class="btn-icon" 
                    onclick={() => removeCandidate(index)}
                    title="削除"
                    aria-label={`候補日 ${index + 1} を削除`}
                  >
                    <span class="material-symbols-rounded" aria-hidden="true">delete</span>
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
            <span class="material-symbols-rounded" aria-hidden="true">add</span>
            候補日時を追加
          </button>
        </div>

        <button type="submit" class="btn btn-primary btn-lg w-full submit-btn" disabled={submitting}>
          <span class="material-symbols-rounded" aria-hidden="true">check_circle</span>
          {submitting ? '作成中...' : '予定作成 ＆ 調整ページを開く'}
        </button>
      </form>
    </div>
  </div>

  <!-- 右カラム：回答・閲覧 ＆ AIログイン -->
  <div class="hero-form-section">
    <div class="sidebar-cards">
      <Accordion>
        <!-- 回答・閲覧 -->
        <AccordionItem title="イベントへの回答・閲覧" icon="search" open={true}>
          <div class="event-code-panel-inner">
            <p class="panel-subtitle">招待されたイベントIDを入力してください</p>
            
            <form onsubmit={navigateToEvent}>
              <div class="form-group">
                <label for="event-id">イベントID (UUID)</label>
                <input 
                  type="text" 
                  id="event-id" 
                  placeholder="e.g. 12345678-abcd-1234-ef00-1234567890ab" 
                  bind:value={eventIdInput}
                  aria-describedby={errorMsg ? "event-id-error" : undefined}
                  aria-invalid={!!errorMsg}
                />
                {#if errorMsg}
                  <p class="error-text" id="event-id-error" role="alert">{errorMsg}</p>
                {/if}
              </div>
              
              <button type="submit" class="btn btn-secondary w-full">
                <span class="material-symbols-rounded" aria-hidden="true">arrow_forward</span>
                イベントページを開く
              </button>
            </form>
          </div>
        </AccordionItem>

        <!-- AI機能を使うにはログイン -->
        <AccordionItem title="✨ AI日程決定 ＆ アシスト機能" icon="auto_awesome" open={true}>
          <div class="ai-login-card-inner">
            <p class="ai-card-text">
              自然文からの候補日自動抽出や、回答結果からAIが最適な日程を自動分析・決定する機能を利用する場合は、ログインしてご利用ください。
            </p>
            <a href="{apiBaseUrl}/api/auth/login" class="btn btn-primary w-full login-btn">
              <span class="material-symbols-rounded" aria-hidden="true">login</span>
              AI機能を使うにはログイン
            </a>
          </div>
        </AccordionItem>

        <!-- 幹事ちゃんの特長・心地よいサポート -->
        <AccordionItem title="幹事ちゃん の使いかた・特長" icon="help" open={false}>
          <div class="mobile-features-accordion">
            <div class="feature-sub-card">
              <span class="material-symbols-rounded feature-icon" aria-hidden="true">edit_note</span>
              <div>
                <strong>言葉から、候補日を紡ぐ</strong>
                <p>「来週の平日夜、渋谷でランチ」のように書くだけでAIが自動生成（要ログイン）</p>
              </div>
            </div>
            <div class="feature-sub-card">
              <span class="material-symbols-rounded feature-icon" aria-hidden="true">psychology</span>
              <div>
                <strong>調和を生み出す決定サポート</strong>
                <p>回答結果と「Aさん必須」等の条件から一番良い候補をAIが推奨（要ログイン）</p>
              </div>
            </div>
            <div class="feature-sub-card">
              <span class="material-symbols-rounded feature-icon" aria-hidden="true">person_check</span>
              <div>
                <strong>おもてなしのシンプル回答</strong>
                <p>回答者はログイン不要。スマホから〇・△・×を選ぶだけの簡単操作</p>
              </div>
            </div>
          </div>
        </AccordionItem>
      </Accordion>
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
      gap: 2rem;
    }
  }

  /* Candidate Row Mobile Friendly */
  @media (max-width: 640px) {
    .hero-container {
      padding: 1.5rem 0;
      gap: 1.5rem;
    }

    .candidate-row {
      flex-wrap: wrap;
      background: rgba(255, 255, 255, 0.4);
      padding: 0.75rem;
      border-radius: var(--radius-sm);
      border: 1px solid var(--border-glass);
    }

    .candidate-row input[type="date"] {
      flex: 1 1 100%;
    }

    .candidate-row input[type="time"] {
      flex: 1 1 42%;
    }

    .time-separator {
      display: inline-block;
      text-align: center;
    }
  }

  /* Mobile Features Accordion */
  .mobile-features-accordion {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    padding: 0.25rem 0;
  }

  .feature-sub-card {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.6rem 0;
  }

  .feature-sub-card .feature-icon {
    font-size: 1.4rem;
    color: var(--color-accent);
    margin-top: 0.1rem;
    flex-shrink: 0;
  }

  .feature-sub-card strong {
    display: block;
    font-size: 0.9rem;
    color: var(--text-primary);
  }

  .feature-sub-card p {
    font-size: 0.82rem;
    color: var(--text-secondary);
    line-height: 1.4;
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
    transition: color var(--transition-fast), background-color var(--transition-fast);
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

  .ai-card-header h3 {
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

  .feature-card h3 {
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
