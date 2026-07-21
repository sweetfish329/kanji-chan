<script lang="ts">
  import { Tabs } from 'bits-ui';
  import type { Snippet } from 'svelte';

  export interface TabItem {
    value: string;
    label: string;
    icon?: string;
    badge?: string;
  }

  interface Props {
    value?: string;
    items: TabItem[];
    class?: string;
    onValueChange?: (val: string) => void;
    children?: Snippet;
  }

  let {
    value = $bindable(),
    items = [],
    class: className = '',
    onValueChange,
    children
  }: Props = $props();

  $effect(() => {
    if (!value && items.length > 0) {
      value = items[0].value;
    }
  });
</script>

<Tabs.Root bind:value {onValueChange} class="bits-tabs-root {className}">
  <Tabs.List class="bits-tabs-list">
    {#each items as item}
      <Tabs.Trigger value={item.value} class="bits-tabs-trigger">
        {#if item.icon}
          <span class="material-symbols-rounded tab-icon" aria-hidden="true">{item.icon}</span>
        {/if}
        <span class="tab-label">{item.label}</span>
        {#if item.badge}
          <span class="tab-badge">{item.badge}</span>
        {/if}
      </Tabs.Trigger>
    {/each}
  </Tabs.List>

  {@render children?.()}
</Tabs.Root>

<style>
  :global(.bits-tabs-root) {
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  :global(.bits-tabs-list) {
    display: flex;
    gap: 0.5rem;
    padding: 0.35rem;
    background: var(--bg-secondary);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    margin-bottom: 1.5rem;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  :global(.bits-tabs-trigger) {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem 1.2rem;
    background: transparent;
    border: none;
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-family: var(--font-sans);
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    white-space: nowrap;
    transition: color 0.2s cubic-bezier(0.16, 1, 0.3, 1), background-color 0.2s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    -webkit-tap-highlight-color: transparent;
  }

  :global(.bits-tabs-trigger:hover) {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.4);
  }

  :global(.bits-tabs-trigger[data-state="active"]) {
    background: var(--bg-primary);
    color: var(--color-primary);
    font-weight: 600;
    box-shadow: var(--shadow-sm);
  }

  .tab-icon {
    font-size: 1.15rem;
    color: var(--color-accent);
  }

  .tab-badge {
    font-size: 0.72rem;
    padding: 0.1rem 0.45rem;
    border-radius: var(--radius-full);
    background: rgba(212, 140, 56, 0.15);
    color: var(--color-accent);
    font-weight: 600;
  }
</style>
