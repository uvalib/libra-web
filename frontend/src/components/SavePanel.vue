<template>
   <Panel header="Save Work" class="save-panel" id="save-panel">
      <div class="panel-content">
         <Fieldset legend="Requirements">
            <div class="requirement">
               <i v-if="props.described" class="done pi pi-check"></i>
               <i v-else class="not-done pi pi-exclamation-circle"></i>
               <span>Describe your work</span>
            </div>
            <div class="requirement">
               <i v-if="props.files" class="done pi pi-check"></i>
               <i v-else class="not-done pi pi-exclamation-circle"></i>
               <span>Add files</span>
            </div>
            <div class="help">
               <span>View <a target="_blank" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist</a> for help.</span>
            </div>
         </Fieldset>
         <Fieldset legend="Visibility">
            <div v-if="props.visibility == 'embargo'" class="embargo">
               <!-- ETD can only be embargoed by an admin. When this happens, lock out the visibility for the user with a message -->
               <div>This work is under embargo.</div>
               <div>Files will NOT be available to anyone until {{ $formatDate(releaseDate) }}.</div>
            </div>
            <div v-else v-for="v in system.userVisibility" :key="v.value" class="visibility-opt">
               <div class="visibility-picker">
                  <RadioButton v-model="visibility" :inputId="v.value"  :value="v.value"  class="visibility-radio-btn" @update:model-value="visibilityUpdated"/>
                  <label :for="v.value" class="visibility" :class="v.value">{{ v.label }}</label>
               </div>
               <div v-if="showLicense(v)" class="license">
                  <a :href="v.license.url">{{ v.license.label }}</a>
               </div>
               <div  v-if="showETDEmbargo(v)" class="limited">
                  <div class="note">Files available to UVA only until:</div>
                  <div class="embargo-date">
                     <span>{{ $formatDate(releaseDate) }}</span>
                     <DatePickerDialog :endDate="releaseDate" :admin="false"
                        :visibility="props.visibility" @picked="endDatePicked"
                        :degree="props.degree" :program="props.program" />
                  </div>
                  <div class="note">After that, files will be be available worldwide.</div>
               </div>
            </div>
         </Fieldset>
         <div class="agree">
            <Checkbox inputId="agree-cb" v-model="agree" :binary="true" />
            <label for="agree-cb">
               I have read and agree to the
               <a href="https://www.library.virginia.edu/libra/etds/etd-license" target="_blank">Libra Deposit License</a>,
               including discussing my deposit access options with my faculty advisor.
            </label>
         </div>
         <div class="button-bar">
            <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
            <Button label="Save" @click="saveClicked()" :disabled="!canSave"/>
         </div>
      </div>
   </Panel>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import DatePickerDialog from "@/components/DatePickerDialog.vue"
import Checkbox from 'primevue/checkbox'
import Fieldset from 'primevue/fieldset'
import RadioButton from 'primevue/radiobutton'
import Panel from 'primevue/panel'
import { useSystemStore } from "@/stores/system"

const emit = defineEmits( ['save', 'cancel'])
const props = defineProps({
   create: {
      type: Boolean,
      default: false,
   },
   files: {
      type: Boolean,
      required: true
   },
   visibility: {
      type: String,
      required: true
   },
   releaseDate: {
      type: String,
      default: null
   },
   releaseVisibility: {
      type: String,
      default: ""
   },
   draft: {
      type: Boolean,
      required: true
   },
   described: {
      type: Boolean,
      required: true
   },
   degree: {
      type: String,
      default: "",
   },
   program: {
      type: String,
      default: "",
   },
})

const system = useSystemStore()
const visibility = ref(props.visibility)
const releaseVisibility = ref("open")
const releaseDate = ref("")
const agree = ref(false)

onMounted( () => {
   visibility.value = props.visibility
   releaseDate.value = props.releaseDate
   releaseVisibility.value = props.releaseVisibility
   agree.value = false
})

const canSave = computed(() =>{
   if (props.described == false ) return false
   return agree.value == true && visibility.value != "" && props.files
})

const endDatePicked = ( (newDate) => {
   releaseDate.value = newDate
})

const visibilityUpdated = (() => {
   if (visibility.value == "embargo" || visibility.value == "uva") {
      releaseVisibility.value = "open"
      let endDate = new Date()
      endDate.setMonth( endDate.getMonth() + 6)
      releaseDate.value = endDate.toJSON()
   }
})

const showLicense = ( (vis) => {
   if (vis.license) {
      return visibility.value == vis.value
   }
   return false
})

const showETDEmbargo = ((vis) =>{
   return (vis.value == 'uva' && visibility.value == vis.value)
})

const saveClicked = (() => {
   emit('save', visibility.value, releaseDate.value, releaseVisibility.value)
})
</script>

<style lang="scss" scoped>
.save-panel {
   font-size: 0.9em;
   .panel-content {
      display: flex;
      flex-direction: column;
      gap: 25px;
   }
   .help {
      font-size: 0.9em;
      margin-top:15px;
   }
   .embargo {
      display: flex;
      flex-direction: column;
      gap: 10px;
   }
   .embargo-date {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      margin: 15px 0;
   }
   div.limited {
      margin: 15px 0 0px 30px;
      .date-row {
         margin-bottom: 10px;
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
      }
   }
   .requirement {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: center;
      gap: 20px;
      .not-done {
         color: $uva-red-B;
      }
      .done {
         color: $uva-green-A;
      }
   }
   .requirement:first-of-type {
      margin-bottom: 5px;
   }
   .visibility-opt {
      margin: 5px 0;
      .visibility-picker {
         display: flex;
         flex-flow: row nowrap;
         align-items: center;
         gap: 10px;
         .visibility {
            flex-grow: 1;
         }
      }
   }
   .agree {
      display: flex;
      flex-direction: row;
      align-items: flex-start;
      gap: 10px;
   }

   .button-bar {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      align-items: stretch;
      gap: 10px;
   };
}
</style>