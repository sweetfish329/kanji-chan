<script lang="ts">
  import { Dialog } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    open?: boolean;
    title?: string;
    description?: string;
    class?: string;
    onOpenChange?: (open: boolean) => void;
    trigger?: Snippet;
    children?: Snippet;
  }

  let {
    open = $bindable(false),
    title,
    description,
    class: className = '',
    onOpenChange,
    trigger,
    children
  }: Props = $props();
</script>

<Dialog.Root bind:open {onOpenChange}>
  {#if trigger}
    <Dialog.Trigger>
      {#snippet child({ props })}
        {@render trigger()}
      {/snippet}
    </Dialog.Trigger>
  {/if}

  <Dialog.Portal>
    <Dialog.Overlay class="bits-dialog-overlay" />
    <Dialog.Content class="bits-dialog-content {className}">
      {#if title}
        <Dialog.Title class="bits-dialog-title">{title}</Dialog.Title>
      {/if}
      {#if description}
        <Dialog.Description class="bits-dialog-desc">{description}</Dialog.Description>
      {/if}

      <div class="bits-dialog-body">
        {@render children?.()}
      </div>

      <Dialog.Close class="bits-dialog-close" aria-label="ダイアログを閉じる">
        <span class="material-symbols-rounded" aria-hidden="true">close</span>
      </Dialog.Close>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<style>
  :global(.bits-dialog-overlay) {
    position: fixed;
    inset: 0;
    z-index: 500;
    background: rgba(28, 36, 30, 0.45);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    animation: dialogOverlayFade 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes dialogOverlayFade {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  :global(.bits-dialog-content) {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    z-index: 501;
    width: 90%;
    max-width: 520px;
    max-height: 85vh;
    overflow-y: auto;
    background: var(--bg-glass);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    padding: 2rem;
    animation: dialogContentShow 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes dialogContentShow {
    from {
      opacity: 0;
      transform: translate(-50%, -46%) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
    }
  }

  :global(.bits-dialog-title) {
    font-family: var(--font-display);
    font-size: 1.35rem;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.4rem;
  }

  :global(.bits-dialog-desc) {
    font-size: 0.88rem;
    color: var(--text-secondary);
    margin-bottom: 1.25rem;
  }

  .bits-dialog-body {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  :global(.bits-dialog-close) {
    position: absolute;
    top: 1.25rem;
    right: 1.25rem;
    background: none;
    border: none;
    color: var(--text-muted);
    padding: 0.35rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s ease, background-color 0.15s ease;
  }

  :global(.bits-dialog-close:hover) {
    color: var(--text-primary);
    background: rgba(42, 64, 50, 0.08);
  }
</style>
