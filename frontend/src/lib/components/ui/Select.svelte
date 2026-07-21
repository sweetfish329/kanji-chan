<script lang="ts">
  import { Select } from 'bits-ui';

  export interface SelectOptionItem {
    value: string;
    label: string;
    disabled?: boolean;
  }

  interface Props {
    value?: string;
    options: SelectOptionItem[];
    placeholder?: string;
    disabled?: boolean;
    onValueChange?: (val: string) => void;
  }

  let {
    value = $bindable(''),
    options = [],
    placeholder = '選択してください...',
    disabled = false,
    onValueChange
  }: Props = $props();

  let selectedLabel = $derived.by(() => {
    const item = options.find(o => o.value === value);
    return item ? item.label : placeholder;
  });
</script>

<Select.Root type="single" bind:value {disabled} {onValueChange}>
  <Select.Trigger class="bits-select-trigger">
    <span class="bits-select-value">{selectedLabel}</span>
    <span class="material-symbols-rounded select-arrow" aria-hidden="true">unfold_more</span>
  </Select.Trigger>

  <Select.Portal>
    <Select.Content class="bits-select-content" sideOffset={4}>
      <Select.Viewport class="bits-select-viewport">
        {#each options as opt}
          <Select.Item
            value={opt.value}
            disabled={opt.disabled}
            label={opt.label}
            class="bits-select-item"
          >
            {opt.label}
          </Select.Item>
        {/each}
      </Select.Viewport>
    </Select.Content>
  </Select.Portal>
</Select.Root>

<style>
  :global(.bits-select-trigger) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0.75rem 1rem;
    background: #FAF8F5;
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    font-family: var(--font-sans);
    font-size: 0.95rem;
    color: var(--text-primary);
    cursor: pointer;
    transition: border-color 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
    -webkit-tap-highlight-color: transparent;
  }

  :global(.bits-select-trigger:hover) {
    border-color: var(--color-accent);
  }

  .bits-select-value {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .select-arrow {
    font-size: 1.25rem;
    color: var(--text-muted);
    flex-shrink: 0;
  }

  :global(.bits-select-content) {
    z-index: 600;
    width: var(--bits-select-anchor-width);
    max-height: 280px;
    background: var(--bg-glass);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    box-shadow: var(--shadow-md);
    padding: 0.35rem;
    overflow-y: auto;
    animation: selectShow 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes selectShow {
    from { opacity: 0; transform: translateY(-4px); }
    to { opacity: 1; transform: translateY(0); }
  }

  :global(.bits-select-item) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.65rem 0.85rem;
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text-primary);
    border-radius: var(--radius-sm);
    cursor: pointer;
    outline: none;
    transition: background-color var(--transition-fast), color var(--transition-fast);
  }

  :global(.bits-select-item:hover), :global(.bits-select-item[data-highlighted]) {
    background: rgba(42, 64, 50, 0.08);
    color: var(--color-primary);
  }

  :global(.bits-select-item[data-state="checked"]) {
    font-weight: 600;
    color: var(--color-primary);
  }
</style>
