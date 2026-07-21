<script lang="ts">
  import { Switch } from 'bits-ui';

  interface Props {
    checked?: boolean;
    disabled?: boolean;
    label?: string;
    onCheckedChange?: (checked: boolean) => void;
  }

  let {
    checked = $bindable(false),
    disabled = false,
    label,
    onCheckedChange
  }: Props = $props();
</script>

<div class="bits-switch-wrapper">
  <Switch.Root
    bind:checked
    {disabled}
    {onCheckedChange}
    class="bits-switch-root"
  >
    <Switch.Thumb class="bits-switch-thumb" />
  </Switch.Root>
  {#if label}
    <span class="switch-label">{label}</span>
  {/if}
</div>

<style>
  .bits-switch-wrapper {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
  }

  :global(.bits-switch-root) {
    width: 44px;
    height: 24px;
    background-color: var(--bg-secondary);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-full);
    position: relative;
    cursor: pointer;
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
    -webkit-tap-highlight-color: transparent;
  }

  :global(.bits-switch-root[data-state="checked"]) {
    background-color: var(--color-primary);
    border-color: var(--color-primary);
  }

  :global(.bits-switch-thumb) {
    display: block;
    width: 18px;
    height: 18px;
    background-color: #ffffff;
    border-radius: 50%;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    transform: translateX(2px);
    transition: transform var(--transition-fast);
  }

  :global(.bits-switch-root[data-state="checked"] .bits-switch-thumb) {
    transform: translateX(22px);
  }

  .switch-label {
    font-size: 0.88rem;
    font-weight: 500;
    color: var(--text-primary);
  }
</style>
