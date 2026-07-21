<script lang="ts">
  import { ToggleGroup } from 'bits-ui';

  export interface ToggleItemOption {
    value: string;
    label: string;
    icon?: string;
    variant?: 'ok' | 'maybe' | 'ng' | 'default';
  }

  interface Props {
    value?: string;
    options: ToggleItemOption[];
    disabled?: boolean;
    onValueChange?: (val: string) => void;
  }

  let {
    value = $bindable(''),
    options = [],
    disabled = false,
    onValueChange
  }: Props = $props();
</script>

<ToggleGroup.Root type="single" bind:value {disabled} {onValueChange} class="bits-toggle-group-root">
  {#each options as opt}
    <ToggleGroup.Item
      value={opt.value}
      class="bits-toggle-item variant-{opt.variant || 'default'}"
    >
      {#if opt.icon}
        <span class="material-symbols-rounded item-icon">{opt.icon}</span>
      {/if}
      <span>{opt.label}</span>
    </ToggleGroup.Item>
  {/each}
</ToggleGroup.Root>

<style>
  :global(.bits-toggle-group-root) {
    display: inline-flex;
    gap: 0.35rem;
    padding: 0.25rem;
    background: var(--bg-secondary);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
  }

  :global(.bits-toggle-item) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.35rem;
    padding: 0.45rem 0.85rem;
    border: none;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-secondary);
    font-family: var(--font-sans);
    font-size: 0.88rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.18s cubic-bezier(0.16, 1, 0.3, 1), color 0.18s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.18s cubic-bezier(0.16, 1, 0.3, 1);
    -webkit-tap-highlight-color: transparent;
  }

  :global(.bits-toggle-item:hover) {
    background: rgba(255, 255, 255, 0.5);
    color: var(--text-primary);
  }

  :global(.bits-toggle-item[data-state="on"]) {
    box-shadow: var(--shadow-sm);
  }

  :global(.bits-toggle-item.variant-ok[data-state="on"]) {
    background: var(--color-ok);
    color: #ffffff;
  }

  :global(.bits-toggle-item.variant-maybe[data-state="on"]) {
    background: var(--color-maybe);
    color: #ffffff;
  }

  :global(.bits-toggle-item.variant-ng[data-state="on"]) {
    background: var(--color-ng);
    color: #ffffff;
  }

  .item-icon {
    font-size: 1.1rem;
  }
</style>
