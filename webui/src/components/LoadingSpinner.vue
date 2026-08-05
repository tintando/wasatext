<script>
export default {
	props: {
		loading: {
			type: Boolean,
			required: true
		}
	}
}
</script>

<template>
  <div v-if="loading">
    <div style="text-align: center">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>
  </div>
  <div v-if="!loading"><slot /></div>
</template>

<style></style>
