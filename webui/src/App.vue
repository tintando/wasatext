<script setup>
import { RouterView } from 'vue-router'
import authService from './services/authService.js'
</script>

<script>
export default {
	mounted() {
		// Initialize auth service with existing token
		authService.init()
	}
}
</script>

<template>
  <div id="app">
    <RouterView />
  </div>
</template>

<style>
#app {
	min-height: 100vh;
}
</style>
