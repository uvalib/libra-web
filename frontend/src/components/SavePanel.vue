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
         <!-- note; use props.visibility here to capture the visibility when the work was loaded, not when changed during edit -->
         <div v-if="props.visibility == 'embargo'" class="embargo no-pad">
            <span class="embargo-note">This work is under embargo.</span>
            <p>Files will be unavavilble to others until:</p>
            <Calendar v-model="releaseDate" showIcon iconDisplay="input" dateFormat="yy-mm-dd"/>
            <p>After that, files will be be available:</p>
            <Dropdown v-model="releaseVisibility" :options="oaVisibilities" optionLabel="label" optionValue="value" />
            <Button severity="danger" label="Lift Embargo" @click="liftEmbargo()" />
         </div>
         <div v-else v-for="v in visibilityOptions" :key="v.value" class="visibility-opt">
            <RadioButton v-model="visibility" :inputId="v.value"  :value="v.value"  class="visibility" :disabled="v.value=='restricted' && props.disablePrivate" />
            <label :for="v.value" class="visibility" :class="v.value">{{ v.label }}</label>
            <div v-if="v.value=='restricted' && props.disablePrivate" class="license">
               Contact libra@virginia.edu for this visibility change.
            </div>
            <div v-if="showLicense(v)" class="license">
               <a :href="v.license.url">{{ v.license.label }}</a>
            </div>
            <div  v-if="showETDEmbargo(v)" class="embargo">
               <p>Files will be available to UVA only for:</p>
               <Dropdown v-model="limitedDuration" :options="limitedDurations" optionLabel="label" optionValue="value" />
               <p>After that, files will be be available worldwide.</p>
            </div>
            <div v-if="showOAEmbargo(v)" class="embargo">
               <p>Files will be unavavilble to others until:</p>
               <Calendar v-model="releaseDate" showIcon iconDisplay="input" dateFormat="yy-mm-dd"/>
               <p>After that, files will be be available:</p>
               <Dropdown v-model="releaseVisibility" :options="oaVisibilities" optionLabel="label" optionValue="value" />
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
         <template v-if="props.type=='oa'">
            <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
            <Button label="Submit" @click="oaSubmitClicked()" :disabled="!canSubmit"/>
         </template>
         <template v-else>
            <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
            <Button label="Save" @click="etdSubmitClicked()" :disabled="!canSubmit"/>
         </template>
      </div>
   </Panel>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Checkbox from 'primevue/checkbox'
import Fieldset from 'primevue/fieldset'
import RadioButton from 'primevue/radiobutton'
import Calendar from 'primevue/calendar'
import Dropdown from 'primevue/dropdown'
import Panel from 'primevue/panel'
import { useSystemStore } from "@/stores/system"
import { useConfirm } from "primevue/useconfirm"

import dayjs from 'dayjs'
import customParseFormat from 'dayjs/plugin/customParseFormat'
dayjs.extend(customParseFormat)

const confirm = useConfirm()
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
      required: true,
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
      default: ""
   },
   releaseVisibility: {
      type: String,
      default: ""
   },
   disablePrivate: {
      type: Boolean,
      default: false
   },
   described: {
      type: Boolean,
      required: true
   }
})

const system = useSystemStore()
const visibility = ref(props.visibility)
const releaseVisibility = ref("open")
const releaseDate = ref(new Date())
const limitedDuration = ref("6-months")
const agree = ref(false)


const oaVisibilities = ref([
   {label: "Worldwide", value: "open"}, {label: "UVA Only", value: "uva"}
])
const limitedDurations = ref([
   {label: "6 Months", value: "6-months"}, {label: "1 Year", value: "1-year"},
   {label: "2 Years", value: "2-years"}, {label: "5 Years", value: "5-years"}
])

onMounted( () => {
   visibility.value = props.visibility
   if ( props.releaseDate != "") {
      releaseDate.value = dayjs(props.releaseDate, "YYYY-MM-DD").toDate()
   }
   if ( props.releaseVisibility != "") {
      releaseVisibility.value = props.releaseVisibility
   }
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

const liftEmbargo = ( () => {
   confirm.require({
      message: `Are you sure you want to lift the embargo on this work?`,
      header: 'Confirm Release Embargo',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         releaseDate.value = new Date()
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
   return (vis.value == 'limited' && visibility.value == vis.value)
})
const showOAEmbargo = ((vis) =>{
   return (vis.value == 'embargo' && visibility.value == vis.value)
})

const oaSubmitClicked = (() => {
   if ( visibility.value == "embargo") {
      emit('submit', visibility.value, releaseDate.value, releaseVisibility.value)
   } else {
      emit('submit', visibility.value)
   }
})
const etdSubmitClicked = (() => {
   if ( visibility.value == "limited") {
      let endDate = new Date()
      if ( limitedDuration.value == "6-months") {
         endDate.setMonth( endDate.getMonth()+6)
      } else {
         let numYears = parseInt(limitedDuration.value.split("-")[0], 10)
         endDate.setFullYear( endDate.getFullYear()+numYears)
      }
      emit("submit", visibility.value, endDate, "open")
   } else {
      emit("submit", visibility.value)
   }
})
</script>

<style lang="scss" scoped>
.save-panel {
   .help {
      font-size: 0.9em;
      margin-top:15px;
   }
   div.embargo.no-pad {
      margin: 0 0 0 0;
   }
   div.embargo {
      font-size: 0.9em;
      margin: 15px 0 30px 30px;
      .embargo-note {
         font-style: italic;
         font-weight: bold;
         color: #bababa;
      }
      button {
         width: 100%;
         margin-top: 15px;
      }
      p {
         margin: 10px 0 2px 0;
         padding: 0;
      }
      .p-calendar, .p-dropdown {
         width: 100%;
         .p-inputtext {
            margin-bottom: 0;
         }
      }
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
         margin: 7px 0px 10px 35px;
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
      justify-content: center;
      align-items: stretch;
      button {
         font-size: 0.85em;
         padding: 5px 10px;
         margin-left: 5px;
      }
   };
}
</style>