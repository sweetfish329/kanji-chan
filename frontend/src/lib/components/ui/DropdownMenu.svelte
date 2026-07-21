<script lang="ts">
  import { DropdownMenu } from 'bits-ui';
  import type { Snippet } from 'svelte';

  export interface MenuItem {
    id: string;
    label: string;
    icon?: string;
    danger?: boolean;
    disabled?: boolean;
    onSelect?: () => void;
  }

  interface Props {
    items: MenuItem[];
    class?: string;
    trigger?: Snippet<[Record<string, any>]>;
  }

  let { items = [], class: className = '', trigger }: Props = $props();
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="bits-dropdown-trigger {className}">
    {#snippet child({ props })}
      {#if trigger}
        {@render trigger(props)}
      {:else}
        <button type="button" {...props} class="btn btn-secondary btn-sm-menu" aria-label="操作メニュー">
          <span class="material-symbols-rounded" aria-hidden="true">more_vert</span>
        </button>
      {/if}
    {/snippet}
  </DropdownMenu.Trigger>

  <DropdownMenu.Portal>
    <DropdownMenu.Content class="bits-dropdown-content" sideOffset={6}>
      {#each items as item}
        <DropdownMenu.Item
          class="bits-dropdown-item {item.danger ? 'is-danger' : ''}"
          disabled={item.disabled}
          onSelect={() => item.onSelect?.()}
        >
          {#if item.icon}
            <span class="material-symbols-rounded item-icon" aria-hidden="true">{item.icon}</span>
          {/if}
          <span>{item.label}</span>
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  :global(.bits-dropdown-trigger) {
    display: inline-flex;
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
  }

  :global(.bits-dropdown-content) {
    z-index: 600;
    min-width: 180px;
    background: var(--bg-glass);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    box-shadow: var(--shadow-md);
    padding: 0.35rem;
    animation: dropdownShow 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes dropdownShow {
    from {
      opacity: 0;
      transform: translateY(-4px) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  :global(.bits-dropdown-item) {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.6rem 0.85rem;
    font-size: 0.88rem;
    font-weight: 500;
    color: var(--text-primary);
    border-radius: var(--radius-sm);
    cursor: pointer;
    outline: none;
    transition: background-color var(--transition-fast), color var(--transition-fast);
  }

  :global(.bits-dropdown-item:hover), :global(.bits-dropdown-item[data-highlighted]) {
    background: rgba(42, 64, 50, 0.08);
    color: var(--color-primary);
  }

  :global(.bits-dropdown-item.is-danger) {
    color: var(--color-ng);
  }

  :global(.bits-dropdown-item.is-danger:hover), :global(.bits-dropdown-item.is-danger[data-highlighted]) {
    background: rgba(184, 74, 65, 0.1);
    color: var(--color-ng);
  }

  .item-icon {
    font-size: 1.15rem;
    color: var(--text-muted);
  }

  .is-danger .item-icon {
    color: var(--color-ng);
  }
</style>
