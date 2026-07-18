<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { api } from '$lib/api';
  import { SvelteToast } from '@zerodevx/svelte-toast';

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
      <span class="material-symbols-rounded logo-icon" aria-hidden="true">calendar_today</span>
      <span class="logo-text gradient-text">幹事ちゃん</span>
    </a>

    <!-- Desktop Nav -->
    <nav class="desktop-nav" aria-label="メインナビゲーション">
      {#if loading}
        <span class="loading-dots" role="status">読み込み中...</span>
      {:else if user}
        <div class="user-menu">
          <span class="welcome-text">こんにちは、<strong>{user.name}</strong> さん</span>
          <a href="/admin" class="btn btn-secondary btn-sm-nav">ダッシュボード</a>
          <button onclick={logout} class="btn btn-secondary btn-sm-nav">ログアウト</button>
        </div>
      {:else}
        <a href="{apiBaseUrl}/api/auth/login" class="btn btn-primary btn-sm-nav">
          <span class="material-symbols-rounded" aria-hidden="true">auto_awesome</span>
          AI機能を使うにはログイン
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

  <!-- Mobile Dropdown Menu -->
  {#if mobileMenuOpen}
    <div class="mobile-menu" aria-label="モバイルナビゲーション">
      <div class="container mobile-menu-inner">
        {#if loading}
          <span class="loading-dots" role="status">読み込み中...</span>
        {:else if user}
          <div class="mobile-user-info">
            <span class="material-symbols-rounded" aria-hidden="true">person</span>
            <span>{user.name} さん</span>
          </div>
          <a href="/admin" class="mobile-nav-item" onclick={() => mobileMenuOpen = false}>
            <span class="material-symbols-rounded" aria-hidden="true">dashboard</span>
            ダッシュボード
          </a>
          <button onclick={() => { mobileMenuOpen = false; logout(); }} class="mobile-nav-item mobile-nav-logout">
            <span class="material-symbols-rounded" aria-hidden="true">logout</span>
            ログアウト
          </button>
        {:else}
          <div class="mobile-ai-banner">
            <span class="material-symbols-rounded" aria-hidden="true">auto_awesome</span>
            <div>
              <p class="mobile-ai-title">AI機能を使うには</p>
              <p class="mobile-ai-sub">自然文での日程作成やAI分析が利用できます</p>
            </div>
          </div>
          <a href="{apiBaseUrl}/api/auth/login" class="btn btn-primary w-full mobile-login-btn" onclick={() => mobileMenuOpen = false}>
            <span class="material-symbols-rounded" aria-hidden="true">login</span>
            ログインしてAIを使う
          </a>
        {/if}
      </div>
    </div>
  {/if}
</header>

<SvelteToast />

<main id="main-content" tabindex="-1">
  {@render children()}
</main>

<footer class="main-footer">
  <div class="container footer-content">
    <p>&copy; 2026 幹事ちゃん - AIサポート日程調整ツール</p>
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
    gap: 0.6rem;
    text-decoration: none;
  }

  .logo-icon {
    color: var(--color-accent);
    font-size: 1.5rem;
  }

  .logo-text {
    font-family: var(--font-display);
    font-size: 1.4rem;
    font-weight: 500;
    letter-spacing: -0.01em;
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

  .mobile-user-info {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.8rem 0;
    font-size: 0.9rem;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border-glass);
    margin-bottom: 0.5rem;
  }

  .mobile-user-info .material-symbols-rounded {
    font-size: 1.2rem;
    color: var(--color-accent);
  }

  .mobile-nav-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem 0;
    text-decoration: none;
    color: var(--text-primary);
    font-weight: 500;
    font-size: 1rem;
    border: none;
    background: none;
    width: 100%;
    cursor: pointer;
    text-align: left;
    border-radius: var(--radius-sm);
    min-height: var(--touch-target);
    -webkit-tap-highlight-color: transparent;
    transition: color var(--transition-fast);
  }

  .mobile-nav-item:active {
    color: var(--color-accent);
  }

  .mobile-nav-logout {
    color: var(--color-ng);
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
    padding: 2rem 0;
    text-align: center;
    color: var(--text-muted);
    font-size: 0.8rem;
    letter-spacing: 0.02em;
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
