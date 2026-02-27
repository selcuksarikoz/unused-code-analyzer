<script setup lang="ts">
// Vue 3 Composition API with TypeScript
// Test file for unused code detection

// ========== IMPORTS ==========
// Used imports
import { ref, computed, onMounted } from 'vue';

// Unused imports
import { watch, watchEffect, provide, inject } from 'vue';
import type { PropType } from 'vue';

// Unused type import
import type { ComponentPublicInstance } from 'vue';

// ========== PROPS ==========
interface Props {
  title: string;
  count?: number;
}

const props = defineProps<Props>();

// ========== EMITS ==========
const emit = defineEmits<{
  (e: 'update', value: string): void;
  (e: 'delete', id: number): void;
}>();

// ========== REACTIVE STATE ==========
// Used state
const counter = ref(0);
const message = ref('Hello Vue');

// Unused state
const unusedRef = ref('never used');
const unusedComputed = computed(() => unusedRef.value.toUpperCase());

// ========== COMPUTED ==========
// Used computed
const doubled = computed(() => counter.value * 2);

// Unused computed
const unusedDoubled = computed(() => counter.value * 3);

// ========== METHODS ==========
// Used method
function increment() {
  counter.value++;
}

// Unused method
function decrement() {
  counter.value--;
}

// Unused arrow function
const unusedHandler = () => {
  console.log('never called');
};

// ========== LIFECYCLE ==========
onMounted(() => {
  console.log('Component mounted');
});

// Unused lifecycle
// watch(counter, () => {}); // commented out but imported watch is unused

// ========== UNUSED VARIABLES ==========
const unusedConst = 'not used';
let unusedLet = 42;

// ========== UNUSED FUNCTION WITH PARAMS ==========
function handleAction(unusedParam: string, action: string): void {
  console.log(action);
}

// ========== UNUSED INTERFACE/TYPE ==========
interface UnusedInterface {
  id: number;
  name: string;
}

type UnusedType = string | number;
</script>

<template>
  <div class="vue-component">
    <h1>{{ message }}</h1>
    <p>Count: {{ counter }}</p>
    <p>Doubled: {{ doubled }}</p>
    <button @click="increment">Increment</button>
  </div>
</template>

<style scoped>
.vue-component {
  padding: 20px;
}
</style>
