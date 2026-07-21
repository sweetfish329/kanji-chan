<script lang="ts">
  import { Popover } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    open?: boolean;
    side?: 'top' | 'right' | 'bottom' | 'left';
    align?: 'start' | 'center' | 'end';
    class?: string;
    trigger?: Snippet;
    children?: Snippet;
    onOpenChange?: (open: boolean) => void;
  }

  let {
    open = $bindable(false),
    side = 'bottom',
    align = 'center',
    class: className = '',
    trigger,
    children,
    onOpenChange
  }: Props = $props();
</script>

<Popover.Root bind:open {onOpenChange}>
  {#if trigger}
    <Popover.Trigger class="bits-popover-trigger">
      {#snippet child({ props })}
        {@render trigger()}
      {/snippet}
    </Popover.Trigger>
  {/if}

  <Popover.Portal>
    <Popover.Content {side} {align} sideOffset={8} class="bits-popover-content {className}">
      {@render children?.()}
      <Popover.Close class="bits-popover-close" aria-label="閉じる">
        <span class="material-symbols-rounded" aria-hidden="true">close</span>
      </Popover.Close>
      <Popover.Arrow class="bits-popover-arrow" />
    </Popover.Content>
  </Popover.Portal>
</Popover.Root>

<style>
  :global(.bits-popover-trigger) {
    display: inline-flex;
    align-items: center;
    background: none;
    border: none;
    padding: 0;
    margin: 0;
    cursor: pointer;
  }

  :global(.bits-popover-content) {
    z-index: 520;
    width: 280px;
    max-width: 90vw;
    background: var(--bg-glass);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    padding: 1rem;
    position: relative;
    animation: popoverFade 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes popoverFade {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  :global(.bits-popover-close) {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: none;
    border: none;
    color: var(--text-muted);
    padding: 0.25rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :global(.bits-popover-close:hover) {
    color: var(--text-primary);
    background: rgba(42, 64, 50, 0.08);
  }
</style>
