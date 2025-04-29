<template>
   <div class="registration">
      <h1>ETD Deposit Registration</h1>
      <Card>
         <template #title>Add deposit registration for one or more users</template>
         <template #content>
            <DepositRegistrationPanel :cancel="false" @submit="submitRegistration"/>
         </template>
      </Card>
   </div>
</template>

<script setup>
import DepositRegistrationPanel from "@/components/DepositRegistrationPanel.vue"
import { useSystemStore } from "@/stores/system"
import { useAdminStore } from "@/stores/admin"
import Card from 'primevue/card'

const system = useSystemStore()
const admin = useAdminStore()

const submitRegistration = ( async ( evt ) => {
   await admin.addRegistrations(evt.program, evt.degree, evt.users)
   if (system.error == "") {
      system.toastMessage("Registration success", "All specified users have been registered.")
   }
})
</script>

<style lang="scss" scoped>
.registration {
   margin: 0 auto;
   min-height: 600px;
   text-align: left;
}
@media only screen and (min-width: 768px) {
   .registration {
      width: 75%;
   }
}
@media only screen and (max-width: 768px) {
   .registration {
      width: 95%;
   }
}
</style>