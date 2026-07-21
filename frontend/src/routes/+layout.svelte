<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { api } from '$lib/api';
  import { Accordion, AccordionItem, DropdownMenu, Toast, type MenuItem } from '$lib';

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '';

  interface User {
    id: number;
    name: string;
    email: string;
    gemini_api_key?: string;
  }

  let user = $state<User | null>(null);
  let loading = $state(true);
  let isMobile = $state(false);
  let mobileMenuOpen = $state(false);

  let userMenuItems = $derived<MenuItem[]>(user ? [
    { id: 'dashboard', label: '幹事ダッシュボード', icon: 'dashboard', onSelect: () => window.location.href = '/admin' },
    { id: 'create', label: '新規イベント作成', icon: 'add_circle', onSelect: () => window.location.href = '/admin' },
    { id: 'logout', label: 'ログアウト', icon: 'logout', danger: true, onSelect: logout }
  ] : []);

  onMount(() => {
    // デバイスタイプ判定: UA + 画面幅の両方で判定
    const ua = navigator.userAgent;
    const isMobileUA = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(ua);
    const isMobileScreen = window.innerWidth <= 768;
    isMobile = isMobileUA || isMobileScreen;

    if (isMobile) {
      document.body.classList.add('is-mobile');
    }

    // 画面回転・リサイズ時に再判定
    const handleResize = () => {
      const resizeMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) || window.innerWidth <= 768;
      isMobile = resizeMobile;
      if (resizeMobile) {
        document.body.classList.add('is-mobile');
      } else {
        document.body.classList.remove('is-mobile');
      }
    };
    window.addEventListener('resize', handleResize);

    api.get<User>('/auth/me')
      .then((data) => {
        user = data;
      })
      .catch(() => {
        user = null;
      })
      .finally(() => {
        loading = false;
      });

    return () => window.removeEventListener('resize', handleResize);
  });

  async function logout() {
    try {
      await api.post('/auth/logout', {});
      user = null;
      window.location.href = '/';
    } catch (err) {
      console.error('Logout failed:', err);
    }
  }

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  let { children } = $props();
</script>

<!-- デフォルト: 全ページをnoindex（トップページのみ上書きでindex） -->
<svelte:head>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<a href="#main-content" class="skip-link">メインコンテンツへスキップ</a>

<header class="main-header">
  <div class="container nav-container">
    <a href="/" class="logo" aria-label="幹事ちゃん トップページ">
      <div class="logo-badge">
        <span class="material-symbols-rounded logo-icon" aria-hidden="true">event_seat</span>
      </div>
      <div class="logo-text-wrapper">
        <span class="logo-text">幹事ちゃん</span>
        <span class="logo-sub">AI スケジュール調整</span>
      </div>
    </a>

    <!-- Desktop Nav -->
    <nav class="desktop-nav" aria-label="メインナビゲーション">
      {#if loading}
        <span class="loading-dots" role="status">読み込み中...</span>
      {:else if user}
        <DropdownMenu items={userMenuItems}>
          {#snippet trigger(props)}
            <button type="button" {...props} class="user-menu-chip" aria-label="幹事アカウントメニュー">
              <span class="material-symbols-rounded" aria-hidden="true">account_circle</span>
              <span class="welcome-name">{user?.name} 幹事</span>
              <span class="material-symbols-rounded arrow-icon" aria-hidden="true">arrow_drop_down</span>
            </button>
          {/snippet}
        </DropdownMenu>
      {:else}
        <a href="{apiBaseUrl}/api/auth/login" class="btn btn-primary btn-sm-nav">
          <span class="material-symbols-rounded" aria-hidden="true">auto_awesome</span>
          幹事ログイン / AI機能
        </a>
      {/if}
    </nav>

    <!-- Mobile Nav Toggle -->
    <button
      class="mobile-menu-btn"
      onclick={toggleMobileMenu}
      aria-label={mobileMenuOpen ? "メニューを閉じる" : "メニューを開く"}
      aria-expanded={mobileMenuOpen}
    >
      <span class="material-symbols-rounded" aria-hidden="true">
        {mobileMenuOpen ? 'close' : 'menu'}
      </span>
    </button>
  </div>

  <!-- Mobile Accordion Menu -->
  {#if mobileMenuOpen}
    <div class="mobile-menu" aria-label="モバイルナビゲーション">
      <div class="container mobile-menu-inner">
        <div class="mobile-menu-header">
          <span class="mobile-menu-title">メニュー一覧</span>
          <button class="mobile-menu-close-btn" onclick={() => mobileMenuOpen = false} aria-label="メニューを閉じる">
            <span class="material-symbols-rounded" aria-hidden="true">close</span>
          </button>
        </div>

        <Accordion>
          <!-- メニュー項目 1: 幹事アカウント / マイページ -->
          <AccordionItem title={user ? `${user.name} 幹事アカウント` : "幹事ログイン / AI機能"} icon="person" open={true}>
            {#if loading}
              <span class="loading-dots" role="status">読み込み中...</span>
            {:else if user}
              <div class="mobile-user-card">
                <div class="mobile-user-details">
                  <span class="mobile-user-email">{user.email}</span>
                </div>
                <div class="mobile-nav-actions">
                  <a href="/admin" class="btn btn-primary w-full" onclick={() => mobileMenuOpen = false}>
                    <span class="material-symbols-rounded" aria-hidden="true">dashboard</span>
                    幹事ダッシュボード
                  </a>
                  <button onclick={() => { mobileMenuOpen = false; logout(); }} class="btn btn-secondary w-full btn-logout">
                    <span class="material-symbols-rounded" aria-hidden="true">logout</span>
                    ログアウト
                  </button>
                </div>
              </div>
            {:else}
              <div class="mobile-ai-banner">
                <span class="material-symbols-rounded" aria-hidden="true">auto_awesome</span>
                <div>
                  <p class="mobile-ai-title">幹事ログイン（完全無料）</p>
                  <p class="mobile-ai-sub">ログインするとイベントの一括管理やAI候補自動提案が使えます</p>
                </div>
              </div>
              <a href="{apiBaseUrl}/api/auth/login" class="btn btn-primary w-full mobile-login-btn" onclick={() => mobileMenuOpen = false}>
                <span class="material-symbols-rounded" aria-hidden="true">login</span>
                Google / GitHub でログイン
              </a>
            {/if}
          </AccordionItem>

          <!-- メニュー項目 2: イベントIDで回答ページを探す -->
          <AccordionItem title="イベントIDでページを開く" icon="search" open={false}>
            <form onsubmit={(e) => {
              e.preventDefault();
              const form = e.currentTarget as HTMLFormElement;
              const input = form.querySelector('input') as HTMLInputElement;
              if (input && input.value.trim()) {
                mobileMenuOpen = false;
                window.location.href = `/event/${input.value.trim()}`;
              }
            }} class="mobile-search-form">
              <p class="mobile-search-desc">共有されたイベントIDを入力して回答・調整ページへ移動できます</p>
              <div class="mobile-search-input-group">
                <input 
                  type="text" 
                  placeholder="例: 12345678-abcd-..." 
                  required 
                  aria-label="イベントID"
                />
                <button type="submit" class="btn btn-primary btn-sm">
                  <span class="material-symbols-rounded" aria-hidden="true">arrow_forward</span>
                </button>
              </div>
            </form>
          </AccordionItem>

          <!-- メニュー項目 3: AI機能と使い方ガイド -->
          <AccordionItem title="幹事ちゃんのAI機能ガイド" icon="auto_awesome" open={false}>
            <ul class="mobile-guide-list">
              <li>
                <span class="material-symbols-rounded" aria-hidden="true">chat</span>
                <div>
                  <strong>自然文からのイベント作成</strong>
                  <p>「来週平日夜に渋谷で懇親会」のように書くだけでAIが候補日時を抽出</p>
                </div>
              </li>
              <li>
                <span class="material-symbols-rounded" aria-hidden="true">analytics</span>
                <div>
                  <strong>AI最適日程おすすめ分析</strong>
                  <p>回答結果と「〇〇さん必須」などの条件を元に最高の日程を提案</p>
                </div>
              </li>
            </ul>
          </AccordionItem>

          <!-- メニュー項目 4: トップページへ -->
          <AccordionItem title="トップページに戻る" icon="home" open={false}>
            <a href="/" class="btn btn-secondary w-full" onclick={() => mobileMenuOpen = false}>
              <span class="material-symbols-rounded" aria-hidden="true">home</span>
              トップページヘ
            </a>
          </AccordionItem>
        </Accordion>
      </div>
    </div>
  {/if}
</header>

<Toast />

<main id="main-content" tabindex="-1">
  {@render children()}
</main>

<footer class="main-footer">
  <div class="container footer-content">
    <div class="footer-brand">
      <span class="footer-logo-text">幹事ちゃん</span>
      <span class="footer-tagline">集いと調和を届ける、AI日程調整プラットフォーム</span>
    </div>
    <p class="copyright">&copy; 2026 幹事ちゃん (Kanji-Chan). All rights reserved.</p>
  </div>
</footer>

<style>
  .main-header {
    background: rgba(250, 248, 245, 0.85);
    backdrop-filter: blur(15px);
    -webkit-backdrop-filter: blur(15px);
    border-bottom: 1px solid var(--border-glass);
    position: sticky;
    top: 0;
    z-index: 200;
    padding: 1.2rem 0;
  }

  .nav-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    text-decoration: none;
  }

  .logo-badge {
    width: 38px;
    height: 38px;
    border-radius: var(--radius-sm);
    background: var(--gradient-brand);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 12px rgba(42, 64, 50, 0.2);
  }

  .logo-icon {
    color: #F8F6F0;
    font-size: 1.3rem;
  }

  .logo-text-wrapper {
    display: flex;
    flex-direction: column;
  }

  .logo-text {
    font-family: var(--font-display);
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1.1;
  }

  .logo-sub {
    font-size: 0.68rem;
    color: var(--text-muted);
    letter-spacing: 0.08em;
  }

  /* Desktop Nav */
  .desktop-nav {
    display: flex;
    align-items: center;
  }

  .user-menu {
    display: flex;
    align-items: center;
    gap: 1.2rem;
  }

  .welcome-text {
    font-size: 0.85rem;
    color: var(--text-secondary);
  }

  .btn-sm-nav {
    padding: 0.6rem 1.2rem;
    font-size: 0.8rem;
    min-height: 38px;
  }

  .loading-dots {
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  /* Mobile Menu Toggle */
  .mobile-menu-btn {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-primary);
    padding: 0.5rem;
    border-radius: var(--radius-sm);
    min-width: var(--touch-target);
    min-height: var(--touch-target);
    align-items: center;
    justify-content: center;
    -webkit-tap-highlight-color: transparent;
    transition: background-color var(--transition-fast);
  }

  .mobile-menu-btn:active {
    background: var(--bg-secondary);
  }

  .mobile-menu-btn .material-symbols-rounded {
    font-size: 1.6rem;
  }

  /* Mobile Dropdown */
  .mobile-menu {
    display: none;
    border-top: 1px solid var(--border-glass);
    background: rgba(250, 248, 245, 0.97);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    animation: slideDown 0.2s ease;
  }

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-8px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  .mobile-menu-inner {
    padding: 1.2rem 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  /* Mobile Menu Header & Accordion Styles */
  .mobile-menu-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.25rem 0 0.85rem;
    border-bottom: 1px dashed var(--border-glass);
    margin-bottom: 0.75rem;
  }

  .mobile-menu-title {
    font-weight: 700;
    font-size: 0.95rem;
    color: var(--color-primary);
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .mobile-menu-close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    padding: 0.2rem;
    cursor: pointer;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .mobile-user-card {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 0.25rem 0;
  }

  .mobile-user-details {
    display: flex;
    flex-direction: column;
  }

  .mobile-user-email {
    font-size: 0.82rem;
    color: var(--text-muted);
  }

  .mobile-nav-actions {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-top: 0.25rem;
  }

  .btn-logout {
    color: var(--color-ng);
    border-color: rgba(184, 74, 65, 0.3);
  }

  .mobile-search-form {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  .mobile-search-desc {
    font-size: 0.82rem;
    color: var(--text-secondary);
  }

  .mobile-search-input-group {
    display: flex;
    gap: 0.5rem;
  }

  .mobile-search-input-group input {
    font-size: 0.88rem;
    padding: 0.6rem 0.8rem;
  }

  .mobile-guide-list {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 0.25rem 0;
  }

  .mobile-guide-list li {
    display: flex;
    align-items: flex-start;
    gap: 0.6rem;
    font-size: 0.85rem;
  }

  .mobile-guide-list li .material-symbols-rounded {
    color: var(--color-accent);
    font-size: 1.2rem;
    margin-top: 0.15rem;
    flex-shrink: 0;
  }

  .mobile-guide-list strong {
    display: block;
    color: var(--text-primary);
    font-weight: 600;
  }

  .mobile-guide-list p {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  .mobile-ai-banner {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 1rem;
    background: var(--bg-secondary);
    border-radius: var(--radius-sm);
    margin-bottom: 0.75rem;
  }

  .mobile-ai-banner .material-symbols-rounded {
    color: var(--color-accent);
    font-size: 1.4rem;
    margin-top: 0.1rem;
  }

  .mobile-ai-title {
    font-weight: 600;
    font-size: 0.9rem;
  }

  .mobile-ai-sub {
    font-size: 0.8rem;
    color: var(--text-muted);
    margin-top: 0.2rem;
  }

  .mobile-login-btn {
    width: 100%;
  }

  .w-full {
    width: 100%;
  }

  main {
    min-height: calc(100vh - 140px);
    padding: 3rem 0;
  }

  .main-footer {
    border-top: 1px solid var(--border-glass);
    padding: 2.5rem 0;
    background: rgba(239, 232, 220, 0.4);
  }

  .footer-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
    text-align: center;
  }

  .footer-brand {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
    justify-content: center;
  }

  .footer-logo-text {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 1.1rem;
    color: var(--color-primary);
  }

  .footer-tagline {
    font-size: 0.82rem;
    color: var(--text-secondary);
  }

  .user-menu-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.4rem 0.85rem;
    background: var(--bg-glass);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-full);
    color: var(--text-primary);
    font-family: var(--font-sans);
    font-size: 0.88rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .user-menu-chip:hover {
    background: rgba(42, 64, 50, 0.08);
    border-color: rgba(42, 64, 50, 0.2);
  }

  .welcome-name {
    color: var(--color-primary);
  }

  .arrow-icon {
    font-size: 1.2rem;
    color: var(--text-muted);
  }

  .copyright {
    color: var(--text-muted);
    font-size: 0.78rem;
  }

  /* ==============================
     Mobile-specific overrides
     ============================== */
  @media (max-width: 768px) {
    .desktop-nav {
      display: none;
    }

    .mobile-menu-btn {
      display: flex;
    }

    .mobile-menu {
      display: block;
    }

    main {
      padding: 1.5rem 0 3rem;
    }

    .main-footer {
      padding-bottom: max(1.5rem, env(safe-area-inset-bottom));
    }
  }
</style>
