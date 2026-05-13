<template>
   <Fieldset class="advisors" pt:contentContainer:aria-labelledby="">
      <template #legend>
         <span class="legend">Advisors</span><span class="required" v-if="user.isAdmin==false"><span class="star">*</span>(required)</span>
      </template>
      <div v-for="(item, index) in etdRepo.work.advisors" class="advisor">
         <div v-if="index==0" class="note">Lookup@ a UVA Computing ID to automatically fill the remaining fields for an advisor.</div>
         <div class="id-field">
            <label :for="`work.advisors[${index}].computeID`">Computing ID</label>
            <div class="control-group">
               <InputText type="text" v-model="advisorLookup[index]" :name="`work.advisors[${index}].computeID`" :id="`work.advisors[${index}].computeID`"/>
               <Button class="check" label="Lookup Advisor"  severity="secondary" @click="checkAdvisorID(index)"/>
               <Button v-if="etdRepo.work.advisors.length > 1 || user.isAdmin" class="remove" severity="danger" :label="removeAdvisorLabel(index)" @click="emit('remove-advisor',index)"/>
            </div>
         </div>
         <Message v-if="etdRepo.work.advisors[index].msg" severity="error" size="small" variant="simple">{{ etdRepo.work.advisors[index].msg }}</Message>
         <div class="two-col">
            <div class="field" >
               <LabeledInput label="First Name" :name="`work.advisors[${index}].firstName`" :required="true" v-model="item.firstName"/>
               <Message v-if="props.form.work?.advisors?.[index]?.firstName?.invalid" severity="error" size="small" variant="simple">
                  {{ props.form.work.advisors[index].firstName.error.message }}
               </Message>
            </div>
            <div class="field" >
               <LabeledInput label="Last Name" :name="`work.advisors[${index}].lastName`" :required="true" v-model="item.lastName"/>
               <Message v-if="props.form.work?.advisors?.[index]?.lastName?.invalid" severity="error" size="small" variant="simple">
                  {{ props.form.work.advisors[index].lastName.error.message }}
               </Message>
            </div>
         </div>
         <div class="two-col">
            <div class="field" >
               <LabeledInput label="Department" :name="`work.advisors[${index}].department`" v-model="item.department"/>
            </div>
            <div class="field" >
               <LabeledInput label="Institution" :name="`work.advisors[${index}].institution`" v-model="item.institution"/>
            </div>
         </div>
      </div>
      <div class="acts">
         <Button label="Add Advisor" @click="emit('add-advisor')" :disabled="addAdvisorDisabled"/>
      </div>
   </Fieldset>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Fieldset from 'primevue/fieldset'
import LabeledInput from '@/components/editform/LabeledInput.vue'
import Message from 'primevue/message'
import InputText from 'primevue/inputtext'
import { useETDStore } from "@/stores/etd"
import { useUserStore } from "@/stores/user"
import axios from 'axios'

const etdRepo = useETDStore()
const user = useUserStore()

const advisorLookup = ref([])

const emit = defineEmits( ['add-advisor', 'remove-advisor', 'set-advisor'])

const props = defineProps({
   form: {
      type: Object,
      required: true 
   }
})

onMounted( async () => {
   console.log("MOUNT ADVISOR PANEL")
   etdRepo.work.advisors.forEach( a => {
      advisorLookup.value.push( a.computeID )
   })
})

const addAdvisorDisabled = computed(() => {
   if ( etdRepo.work.advisors.length == 0) return false
   let lastIdx = etdRepo.work.advisors.length -1
   return etdRepo.work.advisors[lastIdx].lastName == ""
})

const removeAdvisorLabel = ( (index) => {
   if ( etdRepo.work.advisors[index].computeID != "" ) {
      return `Remove ${etdRepo.work.advisors[index].computeID}`
   }
   return "Remove Advisor"
})

const checkAdvisorID = ((idx) => {
   etdRepo.work.advisors[idx].msg = ""
   let cID = advisorLookup.value[idx]
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      if ( etdRepo.work.author.computeID == r.data.cid) {
         etdRepo.work.advisors[idx].msg = cID +" is the author and cannot be an advisor"
         return
      }

      let existing = etdRepo.work.advisors.find( a => a.computeID == r.data.cid)
      if (existing) {
         etdRepo.work.advisors[idx].msg = cID+" is already an advisor"
         return
      }

      let department = ""
      if ( r.data.department && r.data.department.length > 0 ) {
         department = r.data.department[0]
      }
      let adv = {index: idx, computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: department, institution: "University of Virginia"}
      emit("set-advisor", adv)
     
   }).catch( (e) => {
      console.error(e)
      etdRepo.work.advisors[idx].msg = cID+" is not a valid computing ID"
   })
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .remove {
      margin-left: auto;
   }
}
@media only screen and (max-width: 768px) {
   .control-group {
      button, input {
         flex-grow: 1;
      }
   }
}
.acts {
   text-align: right;
}
.advisors {
   .legend {
      font-weight: bold;
   }
   .advisor {
      border-bottom: 1px solid $uva-grey-100;
      padding: 10px 0 20px 0;
      display: flex;
      flex-direction: column;
      gap: 10px;
      .note {
         font-style: italic;
         border-bottom: 1px solid $uva-grey-100;
         padding-bottom: 10px;
         margin-bottom: 10px;
      }
      .id-field {
         display: flex;
         gap: 5px;
         flex-direction: column;
         .control-group {
            display: flex;
            flex-flow: row wrap;
            gap: 10px;
         }
      }
   }
}
</style>