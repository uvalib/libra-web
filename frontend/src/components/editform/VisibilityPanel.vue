<template>
   <Panel header="Access and Visibility" toggleable pt:title:id="visibilty-panel" pt:contentContainer:aria-labelledby="visibilty-panel">
      <div class="license">
         <div class="note">
            For more information, see the
            <a href="https://uvapolicy.virginia.edu/policy/PROV-014" target="_blank" aria-describedby="new-window">Provost Policy on Access Levels for Libra ETD deposits</a>.
         </div>

         <div v-if="etdRepo.visibility == 'embargo' && user.isAdmin == false" class="embargo">
            <!-- ETD can only be embargoed by an admin. When this happens, lock out the visibility for the user with a message -->
            <div>This work is under embargo.</div>
            <div>Files will NOT be available to anyone until {{ $formatDate(etdRepo.embargoReleaseDate) }}.</div>
            <div>After that, files will be be available worldwide.</div>
         </div>
         <template v-else>
            <fieldset>
               <legend>Access and Visibility</legend>
               <div v-for="v in visibilityOpts" :key="v.value" class="visibility-opt">
                  <RadioButton v-model="etdRepo.visibility" name="visibility" :inputId="v.value"  :value="v.value" size="large" @update:model-value="visibilityUpdated"/>
                  <label :for="v.value" class="visibility" :class="v.value">{{ v.label }}</label>
               </div>
            </fieldset>
            <div v-if="etdRepo.visibility == 'uva' || (user.isAdmin && etdRepo.visibility == 'embargo')" class="visibility-info">
               <div v-if="etdRepo.visibility == 'uva'">Files available to UVA only until:</div>
               <div v-else>Files unavailable to anyone until:</div>
               <div class="embargo-date">
                  <span v-if="etdRepo.embargoReleaseDate">{{ $formatDate(etdRepo.embargoReleaseDate) }}</span>
                  <span v-else>Never</span>
                  <DatePickerDialog :endDate="etdRepo.embargoReleaseDate" :admin="user.isAdmin"
                     :visibility="etdRepo.visibility" @picked="endDatePicked"
                     :degree="etdRepo.work.degree" :program="etdRepo.work.program" />
               </div>
               <div>After that, files will be be available worldwide.</div>
            </div>
            <div v-else-if="etdRepo.visibility == 'open'" class="visibility-info">
               All files will be available worldwide.
            </div>
         </template>
         <Message v-if="props.form.visibility?.invalid" severity="error" size="small" variant="simple">{{ props.form.visibility.error.message }}</Message>
      </div>

      <template #icons>
         <span v-if="etdRepo.visibility != ''" class="complete">
            <i class="pi pi-check-circle"></i>
            <span>Complete</span>
         </span>
         <span v-else class="incomplete">
            <i class="pi pi-exclamation-circle"></i>
            <span>Incomplete</span>
         </span>
      </template>
   </Panel>
</template>

<script setup>
import { computed } from 'vue'
import Panel from 'primevue/panel'
import Message from 'primevue/message'
import RadioButton from 'primevue/radiobutton'
import DatePickerDialog from "@/components/editform/DatePickerDialog.vue"
import { useETDStore } from "@/stores/etd"
import { useUserStore } from "@/stores/user"
import { useSystemStore } from "@/stores/system"

const etdRepo = useETDStore()
const user = useUserStore()
const system = useSystemStore()

const emit = defineEmits( ['embargo-changed'])

const props = defineProps({
   form: {
      type: Object,
      required: true 
   }
})

const visibilityOpts = computed( () => {
   if (user.isAdmin) {
      return system.visibility
   }
   return system.userVisibility
})

const visibilityUpdated = (() => {
   if (etdRepo.visibility == "embargo" || etdRepo.visibility == "uva") {
      etdRepo.embargoReleaseVisibility = "open"
      let endDate = new Date()
      endDate.setMonth( endDate.getMonth() + 6)
      etdRepo.embargoReleaseDate = endDate.toJSON()
   }
})

const endDatePicked = ( (newDate) => {
   etdRepo.embargoReleaseDate = newDate
   emit('embargo-changed')
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .visibility-panel {
      min-width: 375px;
   }
}
.license {
   display: flex;
   flex-direction: column;
   gap: 10px;
   fieldset {
      display: flex;
      flex-direction: column;
      gap: 10px;
      border: none;
      outline: none;
      legend {
         display: none;
      }
   }

   .visibility-opt {
      display: flex;
      flex-flow: row nowrap;
      gap: 15px;
      align-items: center;
      .visibility {
         width: 200px;
      }
   }
   .visibility-info {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      gap: 10px;
      margin-top: 15px;
      .embargo-date {
         span {
            margin-right: 20px;
         }
      }
   }
}
</style>