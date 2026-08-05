<template>
  <div class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5)">
    <div class="modal-dialog" :class="dialogClass">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ title }}</h5>
          <button type="button" class="btn-close" @click="$emit('close')" />
        </div>
        <div class="modal-body">
          <slot />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
// BaseModal is the dialog shell every modal in the app shares: backdrop,
// header with a title and a close button, and a slot for the body.
export default {
    name: 'BaseModal',
    props: {
        title: {
            type: String,
            required: true
        },
        dialogClass: {
            type: String,
            default: ''
        }
    },
    emits: ['close']
};
</script>

<style scoped>
.modal {
    display: block !important;
}
</style>
