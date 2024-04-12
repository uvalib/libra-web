<template>
   <Button @click="show" label="ETD Deposit Registration"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="ETD Deposit Registration" position="top">
   </Dialog>
</template>

<script setup>
import { ref } from 'vue'
import Dialog from 'primevue/dialog'

const isOpen = ref(false)

const show = (() => {
   isOpen.value = true
})
</script>

<style lang="scss" scoped>
</style>