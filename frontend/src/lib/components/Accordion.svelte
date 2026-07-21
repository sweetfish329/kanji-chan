<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    class?: string;
    children?: Snippet;
  }

  let {
    class: className = '',
    children
  }: Props = $props();
</script>

<div class="accordion-group {className}">
  {@render children?.()}
</div>

<style>
  .accordion-group {
    display: flex;
    flex-direction: column;
    width: 100%;
  }
</style>
