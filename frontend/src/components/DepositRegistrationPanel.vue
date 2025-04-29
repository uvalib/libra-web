<template>
   <div class="register">
      <div class="row">
         <label>Program:</label>
         <Select v-model="program" :options="system.optPrograms" />
      </div>
      <div class="row">
         <label>Degree:</label>
         <Select v-model="degree" :options="system.optDegrees" />
      </div>
      <FieldSet legend="User List">
         <div class="lookup">
            <div class="user-lookup">
               <InputText v-model="computeID" placeholder="Compute ID" @update:modelValue="idChanged"/>
               <Button class="check" icon="pi pi-search" severity="secondary" @click="lookupComputeID"/>
            </div>
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
         <Button v-if="props.cancel" label="Cancel" severity="secondary" @click="emit('cancel')"/>
         <Button label="Submit" @click="submitClicked" :disabled="submitDisabled"/>
      </div>
   </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Select from 'primevue/select'
import FieldSet from 'primevue/fieldset'
import InputText from 'primevue/inputtext'
import Chip from 'primevue/chip'
import { useSystemStore } from "@/stores/system"
import axios from 'axios'

const system = useSystemStore()
const program = ref("")
const degree = ref("")
const computeID = ref("")
const userError = ref("")
const users = ref([])

const emit = defineEmits( ['cancel', 'submit'])
const props = defineProps({
   cancel: {
      type: Boolean,
      default: true,
   },
})

const submitDisabled = computed( () => {
   return program.value == "" || degree.value == "" || users.value.length == 0
})

onMounted( () => {
   resetForm()
})

const resetForm = (() => {
   program.value = ""
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

const submitClicked = (() => {
   emit('submit', {program: program.value, degree: degree.value, users: users.value})
   if ( props.cancel == false ) {
      resetForm()
   }
})
</script>

<style lang="scss" scoped>
div.row {
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-start;
   align-items: center;
   margin-bottom: 15px;
   gap: 0.5rem;
   .p-select {
      flex-grow: 1;
   }
   label {
      font-weight: bold;
      text-align: right;
      width: 90px;
   }
}

.lookup {
   display: flex;
   flex-direction: column;
   gap: 1rem;

   div.user-lookup {
      display: flex;
      flex-flow: row nowrap;
      gap: 5px;
      align-items: stretch;
   }
   .users {
      padding: 10px;
      border-radius: 0.3rem;
      min-height: 100px;
      border: 1px solid $uva-grey-100;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-start;
      align-items: flex-start;
      gap: 5px;
   }
}
span.err {
   display: inline-block;
   color: $uva-red-A;
   margin-left: 10px;
   font-style: italic;
}
.controls {
   margin-top: 15px;
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-end;
   gap: 10px;
}
</style>