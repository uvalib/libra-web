<template>
   <div class="registration">
      <Panel header="ETD Deposit Registration">
         <DepositRegistrationPanel :cancel="false" @submit="submitRegistration"/>
      </Panel>
   </div>
</template>

<script setup>
import DepositRegistrationPanel from "@/components/DepositRegistrationPanel.vue"
import { useSystemStore } from "@/stores/system"
import { useAdminStore } from "@/stores/admin"
import Panel from 'primevue/panel'

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
   margin: 2% auto;
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