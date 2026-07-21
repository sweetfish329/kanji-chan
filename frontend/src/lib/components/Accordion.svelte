<script lang="ts">
  import { Accordion } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    type?: 'single' | 'multiple';
    value?: string | string[];
    class?: string;
    children?: Snippet;
  }

  let {
    type = 'multiple',
    value = $bindable(),
    class: className = '',
    children
  }: Props = $props();
</script>

<Accordion.Root {type} bind:value class="accordion-group {className}">
  {@render children?.()}
</Accordion.Root>

<style>
  :global(.accordion-group) {
    display: flex;
    flex-direction: column;
    width: 100%;
  }
</style>
