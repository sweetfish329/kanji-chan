<script lang="ts">
  import { Progress } from 'bits-ui';

  interface Props {
    value?: number; // 0..100
    max?: number;
    label?: string;
    class?: string;
  }

  let {
    value = 0,
    max = 100,
    label,
    class: className = ''
  }: Props = $props();

  let percentage = $derived(Math.min(100, Math.max(0, Math.round((value / max) * 100))));
</script>

<div class="bits-progress-wrapper {className}">
  {#if label}
    <div class="bits-progress-label">
      <span>{label}</span>
      <span class="bits-progress-val">{percentage}%</span>
    </div>
  {/if}

  <Progress.Root {value} {max} class="bits-progress-root">
    <div class="bits-progress-fill" style="width: {percentage}%"></div>
  </Progress.Root>
</div>

<style>
  .bits-progress-wrapper {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .bits-progress-label {
    display: flex;
    justify-content: space-between;
    font-size: 0.82rem;
    color: var(--text-secondary);
    font-weight: 500;
  }

  .bits-progress-val {
    font-family: var(--font-mono, monospace);
    color: var(--color-primary);
  }

  :global(.bits-progress-root) {
    position: relative;
    width: 100%;
    height: 8px;
    background: rgba(42, 64, 50, 0.12);
    border-radius: 9999px;
    overflow: hidden;
  }

  .bits-progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--color-primary-light, #2db367) 0%, var(--color-primary, #1e874b) 100%);
    border-radius: 9999px;
    transition: width 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }
</style>
