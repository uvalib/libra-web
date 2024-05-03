<template>
   <Panel header="Save Work" class="save-panel">
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
            <span v-if="type=='etd'">View <a target="_blank" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist</a> for help.</span>
            <span v-else>View the <a href="https://www.library.virginia.edu/libra/open/oc-checklist" target="_blank">Libra Open Checklist</a> for help.</span>
         </div>
      </Fieldset>
      <Fieldset legend="Visibility">
         <div v-if="props.visibility == 'embargo' && props.type=='etd'">
            <!-- ETD can only be embargoed by an admin. When this happens, lock out the visibility for the user with a message -->
            <div class="embargo-note">This work is under embargo.</div>
            <div class="embargo-note">Files will NOT be available to anyone until {{ $formatDate(releaseDate) }}.</div>
         </div>
         <div v-else-if="props.visibility == 'embargo' && props.type=='oa' && props.draft == false" class="embargo no-pad">
            <!-- Users and admins set an embargo for OA works. In this case show a different UI after publication. -->
            <div>This work is under embargo.</div>
            <div>Files will NOT be available to anyone until:</div>
            <div class="embargo-date">
               <span>{{ $formatDate(releaseDate) }}</span>
               <DatePickerDialog :type="props.type" :endDate="releaseDate" :admin="false" :visibility="props.visibility" @picked="endDatePicked" />
            </div>
            <div>After that, files will be be available:</div>
            <div class="embargo-col">
               <Dropdown v-model="releaseVisibility" :options="oaVisibilities" optionLabel="label" optionValue="value" />
               <Button severity="danger" label="Lift Embargo" @click="liftEmbargo()" />
            </div>
         </div>
         <div v-else v-for="v in visibilityOptions" :key="v.value" class="visibility-opt">
            <RadioButton v-model="visibility" :inputId="v.value"  :value="v.value"  class="visibility" @update:model-value="visibilityUpdated"/>
            <label :for="v.value" class="left-margin visibility" :class="v.value">{{ v.label }}</label>
            <div v-if="showLicense(v)" class="license">
               <a :href="v.license.url">{{ v.license.label }}</a>
            </div>
            <div  v-if="showETDEmbargo(v)" class="limited">
               <div class="note">Files available to UVA only until:</div>
               <div class="embargo-date small">
                  <span>{{ $formatDate(releaseDate) }}</span>
                  <DatePickerDialog :type="props.type" :endDate="releaseDate" :admin="false"
                     :visibility="props.visibility" @picked="endDatePicked"
                     :degree="props.degree" :department="props.department" />
               </div>
               <div class="note">After that, files will be be available worldwide.</div>
            </div>
            <div v-if="showOAEmbargo(v)" class="embargo">
               <div>Files will NOT be available to anyone until:</div>
               <div class="embargo-date small">
                  <span>{{ $formatDate(releaseDate) }}</span>
                  <DatePickerDialog :type="props.type" :endDate="releaseDate" :admin="false" :visibility="props.visibility" @picked="endDatePicked" />
               </div>
               <div class="embargo-col">
                  <div>After that, files will be be available:</div>
                  <Dropdown v-model="releaseVisibility" :options="oaVisibilities" optionLabel="label" optionValue="value" />
               </div>
            </div>
         </div>
      </Fieldset>
      <div class="agree">
         <Checkbox inputId="agree-cb" v-model="agree" :binary="true" :disabled="props.create == false" />
         <label v-if="type=='oa'" for="agree-cb">
            By saving this work, I agree to the
            <a href="https://www.library.virginia.edu/libra/open/libra-deposit-license" target="_blank">Libra Deposit Agreement</a>
         </label>
         <label v-else for="agree-cb">
            I have read and agree to the
            <a href="https://www.library.virginia.edu/libra/etds/etd-license" target="_blank">Libra Deposit License</a>,
            including discussing my deposit access options with my faculty advisor.
         </label>
      </div>
      <div class="button-bar">
         <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
         <Button label="Submit" @click="submitClicked()" :disabled="!canSubmit"/>
      </div>
   </Panel>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import DatePickerDialog from "@/components/DatePickerDialog.vue"
import Checkbox from 'primevue/checkbox'
import Fieldset from 'primevue/fieldset'
import RadioButton from 'primevue/radiobutton'
import Dropdown from 'primevue/dropdown'
import Panel from 'primevue/panel'
import { useSystemStore } from "@/stores/system"
import { useConfirm } from "primevue/useconfirm"

const emit = defineEmits( ['submit', 'cancel'])
const props = defineProps({
   type: {
      type: String,
      required: true,
      validator(value) {
         return ['oa', 'etd'].includes(value)
      },
   },
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
   department: {
      type: String,
      default: "",
   },
})

const confirm = useConfirm()
const system = useSystemStore()
const visibility = ref(props.visibility)
const releaseVisibility = ref("open")
const releaseDate = ref("")
const agree = ref(false)

const oaVisibilities = ref([
   {label: "Worldwide", value: "open"}, {label: "UVA Only", value: "uva"}
])

onMounted( () => {
   visibility.value = props.visibility
   releaseDate.value = props.releaseDate
   releaseVisibility.value = props.releaseVisibility
   agree.value = !props.create
})

const visibilityOptions = computed( () => {
   if ( props.type == 'oa') {
      return system.oaVisibility
   }
   return system.etdVisibility
})

const canSubmit = computed(() =>{
   if (props.described == false ) return false
   return agree.value == true && visibility.value != "" && props.files
})

const endDatePicked = ( (newDate) => {
   releaseDate.value = newDate
})

const visibilityUpdated = (() => {
   if (visibility.value == "embargo" || visibility.value == "uva" && props.type == "etd") {
      releaseVisibility.value = "open"
      let endDate = new Date()
      endDate.setMonth( endDate.getMonth() + 6)
      releaseDate.value = endDate.toJSON()
   }
})

const liftEmbargo = ( () => {
   confirm.require({
      message: `Are you sure you want to lift the embargo on this work?`,
      header: 'Confirm Release Embargo',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         const now = new Date()
         releaseDate.value = now.toJSON()
         emit('submit', visibility.value, releaseDate.value, releaseVisibility.value)
      },
   })
})

const showLicense = ( (vis) => {
   if (vis.license) {
      return visibility.value == vis.value
   }
   return false
})

const showETDEmbargo = ((vis) =>{
   if (props.type == "oa") return false
   return (vis.value == 'uva' && visibility.value == vis.value)
})

const showOAEmbargo = ((vis) =>{
   if (props.type == "etd") return false
   return (vis.value == 'embargo' && visibility.value == vis.value)
})

const submitClicked = (() => {
   emit('submit', visibility.value, releaseDate.value, releaseVisibility.value)
})
</script>

<style lang="scss" scoped>
.save-panel {
   :deep(.p-panel-title) {
      font-weight: normal;
   }
   .embargo-note {
      margin: 4px 0;
   }
   .help {
      font-size: 0.9em;
      margin-top:15px;
   }
   div.embargo.no-pad {
      margin: 0 0 0 0;
   }
   .embargo-date {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      margin: 15px 0;
   }
   .embargo-date.small {
      margin: 10px 0;
   }
   .embargo-col {
      display: flex;
      flex-direction: column;
      justify-content: flex-start;
      button{
         margin-top: 20px;
      }
      .p-dropdown {
         margin-top: 5px;
      }
   }
   div.limited {
      font-size: 0.9em;
      margin: 15px 0 0px 30px;
      .date-row {
         margin-bottom: 10px;
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
      }
      .note {
         margin-bottom: 5px;
      }
   }
   div.embargo {
      font-size: 0.9em;
      margin: 15px 0 0px 30px;
   }
   .requirement {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: center;
      i {
         display: inline-block;
         margin-right: 10px;
         font-size: 1.25rem;
      }
      .not-done {
         color: var(--uvalib-red-darker);
      }
      .done {
         color: var(--uvalib-green-dark);
      }
   }
   .requirement:first-of-type {
      margin-bottom: 5px;
   }
   .visibility-opt {
      margin: 5px 0;
      div.visibility {
         padding: 0;
         margin-left: 0;
      }
      .license {
         font-size: 0.8em;
         margin: 10px 0px 15px 35px;
      }
      label.left-margin {
         margin-left: 10px;
         width: 250px;
      }
   }
   .agree {
      display: flex;
      flex-direction: row;
      align-items: flex-start;
      margin: 25px 0;
      label {
         margin-left: 15px;
      }
   }
   .button-bar {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      align-items: stretch;
      button {
         font-size: 0.85em;
         padding: 5px 10px;
         margin-left: 5px;
      }
   };
}
</style>