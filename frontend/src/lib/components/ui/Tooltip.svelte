<script lang="ts">
  import { Tooltip } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    content: string | Snippet;
    side?: 'top' | 'right' | 'bottom' | 'left';
    align?: 'start' | 'center' | 'end';
    delayDuration?: number;
    class?: string;
    trigger?: Snippet;
    children?: Snippet;
  }

  let {
    content,
    side = 'top',
    align = 'center',
    delayDuration = 200,
    class: className = '',
    trigger,
    children
  }: Props = $props();
</script>

<Tooltip.Provider {delayDuration}>
  <Tooltip.Root>
    <Tooltip.Trigger class="bits-tooltip-trigger">
      {#snippet child({ props })}
        {@render (trigger || children)?.()}
      {/snippet}
    </Tooltip.Trigger>
    <Tooltip.Content {side} {align} class="bits-tooltip-content {className}">
      {#if typeof content === 'string'}
        {content}
      {:else if content}
        {@render content()}
      {/if}
      <Tooltip.Arrow class="bits-tooltip-arrow" />
    </Tooltip.Content>
  </Tooltip.Root>
</Tooltip.Provider>

<style>
  :global(.bits-tooltip-trigger) {
    display: inline-flex;
    align-items: center;
    background: none;
    border: none;
    padding: 0;
    margin: 0;
    cursor: pointer;
  }

  :global(.bits-tooltip-content) {
    z-index: 600;
    background: rgba(20, 27, 22, 0.92);
    color: var(--text-primary);
    font-size: 0.78rem;
    font-weight: 500;
    padding: 0.35rem 0.65rem;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-glass);
    box-shadow: var(--shadow-md);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    animation: tooltipFade 0.15s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes tooltipFade {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
