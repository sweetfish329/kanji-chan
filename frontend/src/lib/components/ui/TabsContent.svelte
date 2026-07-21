<script lang="ts">
  import { Tabs } from 'bits-ui';
  import type { Snippet } from 'svelte';

  interface Props {
    value: string;
    class?: string;
    children?: Snippet;
  }

  let { value, class: className = '', children }: Props = $props();
</script>

<Tabs.Content {value} class="bits-tabs-content {className}">
  {@render children?.()}
</Tabs.Content>

<style>
  :global(.bits-tabs-content) {
    animation: tabContentFade 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes tabContentFade {
    from {
      opacity: 0;
      transform: translateY(4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
