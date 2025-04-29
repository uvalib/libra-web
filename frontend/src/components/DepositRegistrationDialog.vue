<template>
   <Button @click="show" label="ETD Deposit Registration"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="ETD Deposit Registration" position="top" style="width:40%">
      <DepositRegistrationPanel @cancel="isOpen= false" @submit="submitRegistration"/>
   </Dialog>
</template>

<script setup>
import { ref } from 'vue'
import Dialog from 'primevue/dialog'
import { useSystemStore } from "@/stores/system"
import { useAdminStore } from "@/stores/admin"
import DepositRegistrationPanel from "@/components/DepositRegistrationPanel.vue"

const system = useSystemStore()
const admin = useAdminStore()
const isOpen = ref(false)

const show = (() => {
   isOpen.value = true
})

const submitRegistration = ( async ( evt ) => {
   await admin.addRegistrations(evt.program, evt.degree, evt.users)
   if (system.error == "") {
      isOpen.value = false
   }
})
</script>

<style lang="scss" scoped>
</style>