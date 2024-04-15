<template>
   <Button @click="show" label="ETD Deposit Registration"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="ETD Deposit Registration" position="top">
      <div class="row">
         <label>Department:</label>
         <Dropdown v-model="department" :options="system.departments" />
      </div>
      <div class="row">
         <label>Degree:</label>
         <Dropdown v-model="degree" :options="system.degrees" />
      </div>
      <FieldSet legend="User List">
         <div class="lookup">
            <InputText v-model="computeID" placeholder="Compute ID" @update:modelValue="idChanged"/>
            <Button class="check" icon="pi pi-search" severity="secondary" @click="lookupComputeID"/>
            <span v-if="userError" class="err">{{ userError }}</span>
            <div class="users">
               <Chip v-for="u in users" removable @remove="removeUser(u.computeID)">
                  <span class="computeID">{{  u.computeID }}</span>
                  <span class="name">- {{  u.lastName }}, {{ u.firstName }}</span>
               </Chip>
            </div>
         </div>
      </FieldSet>
      <div class="controls">
         <Button label="Cancel" severity="secondary" @click="cancelClicked"/>
         <Button label="Submit" @click="submitRegistration" :disabled="submitDisabled"/>
      </div>
   </Dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import Dialog from 'primevue/dialog'
import Dropdown from 'primevue/dropdown'
import FieldSet from 'primevue/fieldset'
import InputText from 'primevue/inputtext'
import Chip from 'primevue/chip'
import { useSystemStore } from "@/stores/system"
import { useAdminStore } from "@/stores/admin"
import axios from 'axios'

const system = useSystemStore()
const admin = useAdminStore()

const isOpen = ref(false)
const department = ref("")
const degree = ref("")
const computeID = ref("")
const userError = ref("")
const users = ref([])

const submitDisabled = computed( () => {
   return department.value == "" || degree.value == "" || users.value.length == 0
})
const show = (() => {
   isOpen.value = true
   department.value = ""
   degree.value = ""
   computeID.value = ""
   userError.value = ""
   users.value = []
})

const idChanged = (() => {
   userError.value=""
})

const removeUser = ( (cID) => {
   const idx = users.value.findIndex( u => u.computeID == cID)
   if (idx > -1 ) {
      users.value.splice(idx,1)
   }
})

const lookupComputeID = ( () => {
   userError.value = ""
   axios.get(`/api/users/lookup/${computeID.value}`).then(r => {
      const idx = users.value.findIndex( u => u.computeID == r.data.cid)
      if ( idx == -1) {
         let user = {
            computeID: r.data.cid, firstName: r.data.first_name, lastName:
            r.data.last_name, email: r.data.email }
         users.value.push( user )
         computeID.value = ""
      } else {
         userError.value = `${computeID.value} already added`
      }
   }).catch( () => {
      userError.value = `${computeID.value} is not a valid compute ID`
   })
})

const cancelClicked = ( () => {
   isOpen.value = false
})

const submitRegistration = ( async () => {
   await admin.addRegistrations(department.value, degree.value, users.value)
   if (system.error == "") {
      isOpen.value = false
   }
})
</script>

<style lang="scss" scoped>
div.row {
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-start;
   align-items: baseline;
   margin-bottom: 15px;
   label {
      font-weight: bold;
      width: 110px;
      text-align: right;
   }
   .p-dropdown {
      margin-left: 10px;
      flex-grow: 1;
   }
}
.lookup {
   input::placeholder {
      color: var(--uvalib-grey-light);
   }
   .p-inputtext {
      width: 130px;
   }
   button {
      margin-left: 5px;
   }
}
.users {
   padding: 10px;
   margin: 10px 0;
   border-radius: 5px;
   min-height: 100px;
   border: 1px solid var(--uvalib-grey-lightest);
   .p-chip {
      margin: 0 5px 5px 0;
   }
   span {
      display: inline-block;
      margin: 4px 0 4px 4px;
   }
   .name {
      font-size: 0.85em;
   }
   .computeID {
      font-weight: bold;
   }
}
span.err {
   display: inline-block;
   color: var(--uvalib-red-emergency);
   margin-left: 10px;
   font-size: 0.9em;
   font-style: italic;
}
.controls {
   text-align: right;
   button {
      margin-left: 10px;
   }
}
</style>