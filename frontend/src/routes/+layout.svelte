<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { api } from '$lib/api';
  import { SvelteToast } from '@zerodevx/svelte-toast';

  // Svelte 5 Runes for Global Auth State
  interface User {
    id: number;
    name: string;
    email: string;
    gemini_api_key?: string;
  }

  let user = $state<User | null>(null);
  let loading = $state(true);

  onMount(async () => {
    try {
      const data = await api.get<User>('/auth/me');
      user = data;
    } catch {
      user = null;
    } finally {
      loading = false;
    }
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

  let { children } = $props();
</script>

<header class="main-header">
  <div class="container nav-container">
    <a href="/" class="logo">
      <span class="material-symbols-rounded logo-icon">calendar_today</span>
      <span class="logo-text gradient-text">幹事ちゃん</span>
    </a>
    <nav>
      {#if loading}
        <span class="loading-dots">読み込み中...</span>
      {:else if user}
        <div class="user-menu">
          <span class="welcome-text">こんにちは、<strong>{user.name}</strong> さん</span>
          <a href="/admin" class="btn btn-secondary btn-sm-nav">ダッシュボード</a>
          <button onclick={logout} class="btn btn-secondary btn-sm-nav">ログアウト</button>
        </div>
      {:else}
        <a href="http://localhost:8080/api/auth/login" class="btn btn-primary btn-sm-nav">
          <span class="material-symbols-rounded">login</span>
          幹事ログイン
        </a>
      {/if}
    </nav>
  </div>
</header>

<SvelteToast />

<main>
  {@render children()}
</main>

<footer class="main-footer">
  <div class="container footer-content">
    <p>&copy; 2026 幹事ちゃん - AIサポート日程調整ツール</p>
  </div>
</footer>

<style>
  .main-header {
    background: hsla(223, 47%, 4%, 0.4);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border-bottom: 1px solid var(--border-glass);
    position: sticky;
    top: 0;
    z-index: 100;
    padding: 1rem 0;
  }

  .nav-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    text-decoration: none;
  }

  .logo-icon {
    color: var(--color-primary);
    font-size: 1.8rem;
  }

  .logo-text {
    font-family: var(--font-display);
    font-size: 1.5rem;
    font-weight: 800;
  }

  .user-menu {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .welcome-text {
    font-size: 0.9rem;
    color: var(--text-secondary);
  }

  .btn-sm-nav {
    padding: 0.5rem 1rem;
    font-size: 0.85rem;
  }

  .loading-dots {
    font-size: 0.9rem;
    color: var(--text-muted);
  }

  main {
    min-height: calc(100vh - 140px);
    padding: 2rem 0;
  }

  .main-footer {
    border-top: 1px solid var(--border-glass);
    padding: 1.5rem 0;
    text-align: center;
    color: var(--text-muted);
    font-size: 0.85rem;
  }
</style>
