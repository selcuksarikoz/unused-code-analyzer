<script>
// Vue 2 Options API
// Test file for unused code detection

export default {
  name: 'Vue2Component',
  
  props: {
    title: {
      type: String,
      required: true
    },
    unusedProp: {
      type: Number,
      default: 0
    }
  },
  
  data() {
    return {
      // Used data
      counter: 0,
      message: 'Hello',
      
      // Unused data
      unusedData: 'not used',
      anotherUnused: 123
    };
  },
  
  computed: {
    // Used computed
    doubled() {
      return this.counter * 2;
    },
    
    // Unused computed
    tripled() {
      return this.counter * 3;
    }
  },
  
  methods: {
    // Used method
    increment() {
      this.counter++;
    },
    
    // Unused method
    decrement() {
      this.counter--;
    },
    
    // Method with unused parameter
    handleEvent(event, unusedParam) {
      console.log(event);
    }
  },
  
  mounted() {
    console.log('mounted');
  },
  
  // Unused lifecycle
  updated() {
    // Not used
  }
};
</script>

<template>
  <div>
    <h2>{{ message }}</h2>
    <p>{{ counter }}</p>
    <button @click="increment">+</button>
  </div>
</template>
