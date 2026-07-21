<script lang="ts" module>
  export interface ToastItem {
    id: string;
    message: string;
    type?: 'info' | 'success' | 'warning' | 'error';
    duration?: number;
  }

  // Toast Store / Runes State
  let toasts = $state<ToastItem[]>([]);

  export const toast = {
    show(message: string, type: ToastItem['type'] = 'info', duration = 3000) {
      const id = Math.random().toString(36).substring(2, 9);
      const item: ToastItem = { id, message, type, duration };
      toasts = [...toasts, item];

      if (duration > 0) {
        setTimeout(() => {
          this.dismiss(id);
        }, duration);
      }
      return id;
    },
    push(message: string, opts?: { theme?: any }) {
      return this.show(message, 'info');
    },
    success(message: string, duration = 3000) {
      return this.show(message, 'success', duration);
    },
    error(message: string, duration = 4000) {
      return this.show(message, 'error', duration);
    },
    info(message: string, duration = 3000) {
      return this.show(message, 'info', duration);
    },
    warning(message: string, duration = 3500) {
      return this.show(message, 'warning', duration);
    },
    dismiss(id: string) {
      toasts = toasts.filter(t => t.id !== id);
    }
  };
</script>

<script lang="ts">
  const typeIcons: Record<NonNullable<ToastItem['type']>, string> = {
    info: 'info',
    success: 'check_circle',
    warning: 'warning',
    error: 'error'
  };
</script>

<div class="bits-toaster-container" aria-live="polite" aria-atomic="true">
  {#each toasts as t (t.id)}
    <div class="bits-toast bits-toast-{t.type || 'info'}" role="status">
      <span class="material-symbols-rounded toast-icon" aria-hidden="true">
        {typeIcons[t.type || 'info']}
      </span>
      <span class="toast-message">{t.message}</span>
      <button class="toast-close" onclick={() => toast.dismiss(t.id)} aria-label="通知を閉じる">
        <span class="material-symbols-rounded" aria-hidden="true">close</span>
      </button>
    </div>
  {/each}
</div>

<style>
  .bits-toaster-container {
    position: fixed;
    bottom: 1.5rem;
    right: 1.5rem;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    max-width: 400px;
    width: calc(100vw - 3rem);
    pointer-events: none;
  }

  .bits-toast {
    pointer-events: auto;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.85rem 1rem;
    background: var(--bg-glass);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    color: var(--text-primary);
    font-size: 0.9rem;
    line-height: 1.4;
    animation: toastSlideIn 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes toastSlideIn {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .toast-icon {
    font-size: 1.25rem;
    flex-shrink: 0;
  }

  .bits-toast-success .toast-icon { color: #2db367; }
  .bits-toast-error .toast-icon { color: #e55353; }
  .bits-toast-warning .toast-icon { color: #f59e0b; }
  .bits-toast-info .toast-icon { color: #3b82f6; }

  .toast-message {
    flex: 1;
    font-weight: 500;
  }

  .toast-close {
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

  .toast-close:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.1);
  }
</style>
