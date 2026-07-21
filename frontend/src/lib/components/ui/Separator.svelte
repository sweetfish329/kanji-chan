<script lang="ts">
  import { Separator } from 'bits-ui';

  interface Props {
    orientation?: 'horizontal' | 'vertical';
    class?: string;
  }

  let { orientation = 'horizontal', class: className = '' }: Props = $props();
</script>

<Separator.Root {orientation} class="bits-separator {orientation} {className}" />

<style>
  :global(.bits-separator) {
    background-color: var(--border-glass);
    flex-shrink: 0;
  }

  :global(.bits-separator.horizontal) {
    height: 1px;
    width: 100%;
    margin: 1.25rem 0;
  }

  :global(.bits-separator.vertical) {
    width: 1px;
    height: 100%;
    margin: 0 1rem;
  }
</style>
