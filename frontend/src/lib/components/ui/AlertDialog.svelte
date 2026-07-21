<script lang="ts">
  import { AlertDialog } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    open?: boolean;
    title?: string;
    description?: string;
    confirmText?: string;
    cancelText?: string;
    danger?: boolean;
    class?: string;
    onConfirm?: () => void;
    onCancel?: () => void;
    onOpenChange?: (open: boolean) => void;
    trigger?: Snippet;
    children?: Snippet;
  }

  let {
    open = $bindable(false),
    title = '確認',
    description,
    confirmText = '実行する',
    cancelText = 'キャンセル',
    danger = false,
    class: className = '',
    onConfirm,
    onCancel,
    onOpenChange,
    trigger,
    children
  }: Props = $props();

  function handleConfirm() {
    onConfirm?.();
    open = false;
  }

  function handleCancel() {
    onCancel?.();
    open = false;
  }
</script>

<AlertDialog.Root bind:open {onOpenChange}>
  {#if trigger}
    <AlertDialog.Trigger>
      {#snippet child({ props })}
        {@render trigger()}
      {/snippet}
    </AlertDialog.Trigger>
  {/if}

  <AlertDialog.Portal>
    <AlertDialog.Overlay class="bits-alert-overlay" />
    <AlertDialog.Content class="bits-alert-content {className}">
      <AlertDialog.Title class="bits-alert-title">{title}</AlertDialog.Title>
      {#if description}
        <AlertDialog.Description class="bits-alert-desc">{description}</AlertDialog.Description>
      {/if}

      {#if children}
        <div class="bits-alert-body">
          {@render children()}
        </div>
      {/if}

      <div class="bits-alert-actions">
        <AlertDialog.Cancel class="btn btn-secondary" onclick={handleCancel}>
          {cancelText}
        </AlertDialog.Cancel>
        <AlertDialog.Action class="btn {danger ? 'btn-danger' : 'btn-primary'}" onclick={handleConfirm}>
          {confirmText}
        </AlertDialog.Action>
      </div>
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>

<style>
  :global(.bits-alert-overlay) {
    position: fixed;
    inset: 0;
    z-index: 550;
    background: rgba(20, 27, 22, 0.55);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    animation: alertOverlayFade 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes alertOverlayFade {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  :global(.bits-alert-content) {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    z-index: 551;
    width: 90%;
    max-width: 480px;
    background: var(--bg-glass);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    padding: 1.75rem;
    animation: alertContentShow 0.22s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes alertContentShow {
    from {
      opacity: 0;
      transform: translate(-50%, -46%) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
    }
  }

  :global(.bits-alert-title) {
    font-family: var(--font-display);
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
  }

  :global(.bits-alert-desc) {
    font-size: 0.9rem;
    color: var(--text-secondary);
    line-height: 1.5;
    margin-bottom: 1.25rem;
  }

  .bits-alert-body {
    margin-bottom: 1.25rem;
  }

  .bits-alert-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
  }
</style>
