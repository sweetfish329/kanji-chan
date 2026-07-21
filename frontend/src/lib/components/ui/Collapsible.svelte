<script lang="ts">
  import { Collapsible } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    open?: boolean;
    title?: string;
    class?: string;
    onOpenChange?: (open: boolean) => void;
    trigger?: Snippet;
    children?: Snippet;
  }

  let {
    open = $bindable(false),
    title = '詳細を表示',
    class: className = '',
    onOpenChange,
    trigger,
    children
  }: Props = $props();
</script>

<Collapsible.Root bind:open {onOpenChange} class="bits-collapsible-root {className}">
  <Collapsible.Trigger class="bits-collapsible-trigger">
    {#if trigger}
      {@render trigger()}
    {:else}
      <span>{title}</span>
      <span class="material-symbols-rounded chevron" class:open aria-hidden="true">
        expand_more
      </span>
    {/if}
  </Collapsible.Trigger>

  <Collapsible.Content class="bits-collapsible-content">
    <div class="bits-collapsible-inner">
      {@render children?.()}
    </div>
  </Collapsible.Content>
</Collapsible.Root>

<style>
  :global(.bits-collapsible-root) {
    width: 100%;
    display: flex;
    flex-direction: column;
  }

  :global(.bits-collapsible-trigger) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    background: none;
    border: none;
    padding: 0.5rem 0;
    color: var(--color-primary);
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    transition: color 0.15s ease;
  }

  :global(.bits-collapsible-trigger:hover) {
    color: var(--color-primary-dark, #166437);
  }

  .chevron {
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .chevron.open {
    transform: rotate(180deg);
  }

  :global(.bits-collapsible-content) {
    overflow: hidden;
    transition: height 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .bits-collapsible-inner {
    padding-top: 0.5rem;
  }
</style>
