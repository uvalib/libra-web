<template>
   <div class="register">
      <div class="row">
         <label id="program">Program:</label>
         <Select v-model="program" :options="system.optPrograms" ariaLabelledby="program" />
      </div>
      <div class="row">
         <label id="degree">Degree:</label>
         <Select v-model="degree" :options="system.optDegrees" ariaLabelledby="degree" />
      </div>
      <FieldSet legend="Registrants">
         <div class="lookup">
            <div class="note">
               Add one ore more computing IDs to register. IDs can be separated by spaces, commas or newlines. Click 'Add'
               to validate the IDs and add them to the registration list.
            </div>
            <div class="user-lookup">
               <TextArea v-model="computeID"  rows="2" @update:modelValue="idChanged" fluid placeholder="Computing IDs"/>
               <Button label="Add" size="small" severity="secondary" @click="lookup" :loading="working" :disabled="computeID.length == 0"/>
            </div>
            <div class="users">
               <Chip v-for="u in users" removable @remove="removeUser(u.computeID)">
                  <span class="computeID">{{  u.computeID }}</span>
                  <span class="name">- {{  u.lastName }}, {{ u.firstName }}</span>
               </Chip>
            </div>
            <div class="errors">
               <div v-for="err in userErrors" class="err">{{ err }}</div>
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
import TextArea from 'primevue/textarea'
import Chip from 'primevue/chip'
import { useSystemStore } from "@/stores/system"
import axios from 'axios'

const system = useSystemStore()
const program = ref("")
const degree = ref("")
const computeID = ref("")
const userErrors = ref([])
const users = ref([])
const working = ref(false)

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
   userErrors.value = []
   users.value = []
})

const idChanged = (() => {
   userErrors.value=[]
})

const removeUser = ( (cID) => {
   const idx = users.value.findIndex( u => u.computeID == cID)
   if (idx > -1 ) {
      users.value.splice(idx,1)
   }
})

const lookup = ( () => {
   let normalized = computeID.value.replace(/\n|\s+/g, ',').trim()
   normalized = normalized.replace(/,+/g, ',')
   let request = [ normalized ]
   if ( normalized.includes(",")) {
      request = normalized.split(",")
   }

   working.value = true
   userErrors.value = []
   request.forEach( computeID => {
      if (computeID.length == 0) return
      axios.get(`/api/users/lookup/${computeID}`).then(r => {
         const idx = users.value.findIndex( u => u.computeID == r.data.cid)
         if ( idx == -1) {
            let user = {
               computeID: r.data.cid, firstName: r.data.first_name, lastName:
               r.data.last_name, email: r.data.email }
            users.value.push( user )
         } else {
            userErrors.value.push(`${computeID} already added`)
         }
      }).catch( () => {
         userErrors.value.push(`${computeID} is not a valid compute ID`)
      })
   })
   computeID.value = ""
   working.value = false
})

const submitClicked = (() => {
   emit('submit', {program: program.value, degree: degree.value, users: users.value})
   if ( props.cancel == false ) {
      resetForm()
   }
})
</script>

<style lang="scss" scoped>
.note {
   color: $uva-grey-A;
}
.register {
   display: flex;
   flex-direction: column;
   gap: 15px;
}
div.row {
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-start;
   align-items: center;
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
   gap: 20px;

   div.user-lookup {
      display: flex;
      flex-flow: row nowrap;
      gap: 0.5rem;
      align-items: flex-start;
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
.errors {
   display: flex;
   flex-direction: column;
   gap: 5px;
   .err {
      display: inline-block;
      color: $uva-red-A;
      margin-left: 10px;
      font-style: italic;
   }
}
.controls {
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-end;
   gap: 10px;
}
</style>