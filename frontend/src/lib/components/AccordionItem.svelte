<script lang="ts">
  import { Accordion } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    title: string;
    value?: string;
    icon?: string;
    badge?: string;
    open?: boolean; // 互換性のためのフラグ
    disabled?: boolean;
    class?: string;
    children?: Snippet;
  }

  let {
    title,
    value = `item-${Math.random().toString(36).substring(2, 10)}`,
    icon,
    badge,
    open = false,
    disabled = false,
    class: className = '',
    children
  }: Props = $props();
</script>

<Accordion.Item {value} {disabled} class="accordion-item {className}">
  <Accordion.Header class="accordion-header-wrapper">
    <Accordion.Trigger class="accordion-header">
      {#snippet children({ open })}
        <div class="accordion-title-wrapper">
          {#if icon}
            <span class="material-symbols-rounded accordion-icon" aria-hidden="true">{icon}</span>
          {/if}
          <span class="accordion-title-text">{title}</span>
          {#if badge}
            <span class="accordion-badge">{badge}</span>
          {/if}
        </div>
        <span class="material-symbols-rounded accordion-arrow" class:rotated={open} aria-hidden="true">
          expand_more
        </span>
      {/snippet}
    </Accordion.Trigger>
  </Accordion.Header>

  <Accordion.Content class="accordion-body">
    <div class="accordion-content-inner">
      {@render children?.()}
    </div>
  </Accordion.Content>
</Accordion.Item>

<style>
  :global(.accordion-item) {
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    background: var(--bg-glass);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    overflow: hidden;
    transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
    margin-bottom: 0.75rem;
  }

  :global(.accordion-item:last-child) {
    margin-bottom: 0;
  }

  :global(.accordion-item[data-state="open"]) {
    border-color: rgba(42, 64, 50, 0.2);
    box-shadow: var(--shadow-sm);
  }

  :global(.accordion-header-wrapper) {
    display: flex;
    margin: 0;
  }

  :global(.accordion-header) {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem 1.25rem;
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: left;
    color: var(--text-primary);
    font-family: var(--font-sans);
    font-size: 0.98rem;
    font-weight: 600;
    min-height: var(--touch-target);
    transition: background-color var(--transition-fast);
    -webkit-tap-highlight-color: transparent;
  }

  :global(.accordion-header:hover) {
    background: rgba(42, 64, 50, 0.03);
  }

  .accordion-title-wrapper {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex: 1;
    min-width: 0;
  }

  .accordion-icon {
    font-size: 1.25rem;
    color: var(--color-accent);
    flex-shrink: 0;
  }

  .accordion-title-text {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .accordion-badge {
    font-size: 0.72rem;
    padding: 0.15rem 0.5rem;
    border-radius: var(--radius-full);
    background: rgba(212, 140, 56, 0.12);
    color: var(--color-accent);
    font-weight: 600;
    flex-shrink: 0;
  }

  .accordion-arrow {
    font-size: 1.4rem;
    color: var(--text-muted);
    transition: transform 0.25s var(--transition-spring), color var(--transition-fast);
    flex-shrink: 0;
    margin-left: 0.5rem;
  }

  .accordion-arrow.rotated {
    transform: rotate(180deg);
    color: var(--color-primary);
  }

  :global(.accordion-body) {
    animation: accordionSlideDown 0.22s var(--transition-normal);
  }

  @keyframes accordionSlideDown {
    from {
      opacity: 0;
      transform: translateY(-4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .accordion-content-inner {
    padding: 0.5rem 1.25rem 1.25rem;
    border-top: 1px dashed var(--border-glass);
    color: var(--text-secondary);
    font-size: 0.92rem;
  }
</style>
