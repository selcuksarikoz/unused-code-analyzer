<!-- Svelte Component -->
<!-- Test file for unused code detection -->

<script lang="ts">
  // ========== IMPORTS ==========
  import { onMount, createEventDispatcher } from 'svelte';
  
  // Unused imports
  import { tick, onDestroy, onUpdate } from 'svelte';
  import type { SvelteComponent } from 'svelte';
  
  // ========== PROPS ==========
  export let title: string;
  export let count: number = 0;
  
  // Unused prop
  export let unusedProp: string = 'default';
  
  // ========== INTERNAL STATE ==========
  // Used state
  let internalCount = count;
  let message = 'Hello Svelte';
  
  // Unused state
  let unusedVariable = 'never used';
  let anotherUnused = 42;
  
  // ========== REACTIVE STATEMENTS ==========
  $: doubled = internalCount * 2;
  $: console.log('count changed:', internalCount);
  
  // Unused reactive
  $: unusedReactive = internalCount * 3;
  
  // ========== FUNCTIONS ==========
  // Used function
  function increment() {
    internalCount += 1;
  }
  
  // Unused function
  function decrement() {
    internalCount -= 1;
  }
  
  // Function with unused parameter
  function handleClick(event: MouseEvent, unusedParam: string): void {
    console.log('clicked');
  }
  
  // ========== EVENTS ==========
  const dispatch = createEventDispatcher();
  
  function notify() {
    dispatch('update', internalCount);
  }
  
  // Unused event handler
  function unusedHandler() {
    dispatch('unused');
  }
  
  // ========== LIFECYCLE ==========
  onMount(() => {
    console.log('Component mounted');
  });
  
  // ========== TYPES ==========
  interface UnusedInterface {
    id: number;
    name: string;
  }
  
  type UnusedType = string | number;
  
  // Unused type annotation
  let typedUnused: UnusedType = 'test';
</script>

<main>
  <h1>{message}</h1>
  <p>Count: {internalCount}</p>
  <p>Doubled: {doubled}</p>
  <button on:click={increment}>Increment</button>
  <button on:click={notify}>Notify</button>
</main>

<style>
  main {
    padding: 20px;
  }
</style>
