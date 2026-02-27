<template>
  <div>{{ usedVariable }}</div>
</template>

<script>
import { usedFunction, unusedExport } from './utils';
import Vue from 'vue';

export default {
  name: 'UsedComponent',
  data() {
    return {
      usedVariable: 'test'
    };
  },
  methods: {
    usedMethod() {
      usedFunction();
    }
  },
  computed: {
    usedComputed() {
      return this.usedVariable;
    }
  }
};
</script>

<script setup>
import { ref } from 'vue';

const usedRef = ref('test');
const unusedRef = ref('unused');

function usedFunction() {
  console.log('used');
}

function unusedFunction() {
  console.log('unused');
}
</script>

<style scoped>
div {
  color: red;
}
</style>
